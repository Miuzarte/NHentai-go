package NHentai

import (
	"context"
	"iter"
	"net/http"
	"path"
	"strconv"

	"github.com/Miuzarte/NHentai-go/internal/utils"
)

const API_URL = "https://nhentai.net"

var ApiUrl = API_URL

// 负载均衡到所有 cdn
var UP UrlProviderType = &urlProvider{}

const (
	API_SEARCH        = "/api/galleries/search"
	API_SEARCH_TAGGED = "/api/galleries/tagged"
	API_GALLERY       = "/api/gallery"
)

type Sort = string

const (
	SORT_POPULAR = "popular"
	SORT_DATE    = "date"
)

var threads = 4 // 下载并发数

// SetThreads 设置下载并发数
func SetThreads(n int) {
	if n <= 0 {
		n = 1
	}
	threads = n
}

// SetUseEnvProxy 设置是否使用系统环境变量中的代理
//
// 默认为 true
func SetUseEnvProxy(b bool) {
	ht := httpClient.Transport.(*http.Transport)
	if b {
		ht.Proxy = http.ProxyFromEnvironment
	} else {
		ht.Proxy = nil
	}
}

type ImageInfo struct {
	T string `json:"t"` // type // "w"
	W int    `json:"w"`
	H int    `json:"h"`
}

// [TODO] 尝试兼容 [EHentai.TranslateMulti]
// 处理 "tag" 域
type Tag struct {
	Id    int    `json:"id"`
	Type  string `json:"type"` // "tag" | "language" | "category" | "parody" | "artist" | "group" ... // [TODO] TagSet type
	Name  string `json:"name"`
	Url   string `json:"url"` // "/{.Type}/{.Name}"
	Count int    `json:"count"`
}

func (t Tag) String() string {
	return t.Type + ":" + t.Name
}

func (t Tag) Search(ctx context.Context, page int, sort Sort) (*SearchResp, error) {
	return SearchTagged(ctx, t.Id, page, sort)
}

type Tags []Tag

func (t Tags) Namespaces() (namespaces []string) {
	namespaces = make([]string, len(t))
	for i := range t {
		namespaces[i] = t[i].Type
	}
	s := make(utils.Set[string])
	return s.Clean(namespaces)
}

func (t Tags) Set() (ts []TagSet) {
	setPos := make(map[string]int)
	for _, tag := range t {
		i, ok := setPos[tag.Type]
		if !ok {
			i = len(ts)
			setPos[tag.Type] = i
			ts = append(ts, TagSet{Namespace: tag.Type})
		}
		ts[i].Tags = append(ts[i].Tags, tag.Name)
	}
	return ts
}

type TagSet struct {
	Namespace string
	Tags      []string
}

func (ts Tags) Strings() []string {
	ss := make([]string, 0, len(ts))
	for _, tag := range ts {
		ss = append(ss, tag.String())
	}
	return ss
}

type Gallery struct {
	Id      int    `json:"id"`
	MediaId string `json:"media_id"`
	Title   struct {
		English  string `json:"english"`
		Japanese string `json:"japanese"`
		Pretty   string `json:"pretty"`
	} `json:"title"`
	Images struct {
		Pages     []ImageInfo `json:"pages"`
		Cover     ImageInfo   `json:"cover"`
		Thumbnail ImageInfo   `json:"thumbnail"`
	} `json:"images"`
	Scanlator    string `json:"scanlator"`
	UploadDate   int    `json:"upload_date"`
	Tags         Tags   `json:"tags"`
	NumPages     int    `json:"num_pages"`
	NumFavorites int    `json:"num_favorites"`
}

type Gallerys []*Gallery

func (gs Gallerys) DownloadCoversIter(ctx context.Context) iter.Seq2[Image, error] {
	return newDownloader(ctx, newCoversDownload(gs)).downloadIter()
}

func (g *Gallery) DownloadThumbsIter(ctx context.Context) iter.Seq2[Image, error] {
	return newDownloader(ctx, newThumbsDownload(g)).downloadIter()
}

func (g *Gallery) DownloadPagesIter(ctx context.Context) iter.Seq2[Image, error] {
	return newDownloader(ctx, newPagesDownload(g)).downloadIter()
}

func GetGallery(ctx context.Context, bookId int) (*Gallery, error) {
	url := toUrl(ApiUrl)
	url.Path = path.Join(API_GALLERY, strconv.Itoa(bookId))

	return getAndUnmarshalTo[Gallery](ctx, url.String(), nil)
}

func (g *Gallery) PageFilename(i int) string {
	return strconv.Itoa(i+1) + "." + getFullType(g.Images.Pages[i].T)
}

func (g *Gallery) PageUrl(i int) string {
	return UP.NextImageUrl() + path.Join(
		"/galleries",
		g.MediaId,
		g.PageFilename(i),
	)
}

func (g *Gallery) PageUrlsIter() iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		for i := range g.Images.Pages {
			if !yield(i, g.PageUrl(i)) {
				return
			}
		}
	}
}

func (g *Gallery) PageUrls() []string {
	urls := make([]string, 0, len(g.Images.Pages))
	for i := range g.Images.Pages {
		urls = append(urls, g.PageUrl(i))
	}
	return urls
}

func (g *Gallery) ThumbFilename(i int) string {
	return strconv.Itoa(i+1) + "t" + "." + getFullType(g.Images.Pages[i].T)
}

func (g *Gallery) ThumbUrl(i int) string {
	return UP.NextThumbUrl() + path.Join(
		"/galleries",
		g.MediaId,
		g.ThumbFilename(i),
	)
}

func (g *Gallery) ThumbUrlsIter() iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		for i := range g.Images.Pages {
			if !yield(i, g.ThumbUrl(i)) {
				return
			}
		}
	}
}

func (g *Gallery) ThumbUrls() []string {
	urls := make([]string, 0, len(g.Images.Pages))
	for i := range g.Images.Pages {
		urls = append(urls, g.ThumbUrl(i))
	}
	return urls
}

func (g *Gallery) CoverFilename() string {
	return "cover" + "." + getFullType(g.Images.Cover.T)
}

func (g *Gallery) CoverUrl() string {
	return UP.NextThumbUrl() + path.Join(
		"/galleries",
		g.MediaId,
		g.CoverFilename(),
	)
}

func (g *Gallery) GetRelated(ctx context.Context) (Gallerys, error) {
	type RelatedResp struct {
		Result Gallerys `json:"result"`
	}

	url := toUrl(ApiUrl)
	url.Path = path.Join(API_GALLERY, strconv.Itoa(g.Id), "related")

	r, err := getAndUnmarshalTo[RelatedResp](ctx, url.String(), nil)
	if err != nil {
		return nil, err
	}
	return r.Result, nil
}

type SearchResp struct {
	Result   Gallerys `json:"result"`
	NumPages int      `json:"num_pages"`
	PerPage  int      `json:"per_page"`
}

func Search(ctx context.Context, query string, page int, sort Sort) (*SearchResp, error) {
	url := toUrl(ApiUrl)
	url.Path = API_SEARCH

	q := url.Query()
	q.Add("query", query)
	if page > 0 {
		q.Add("page", strconv.Itoa(page))
	}
	if sort != "" {
		q.Add("sort", sort)
	}
	url.RawQuery = q.Encode()

	return getAndUnmarshalTo[SearchResp](ctx, url.String(), nil)
}

func SearchTagged(ctx context.Context, tag_id int, page int, sort Sort) (*SearchResp, error) {
	url := toUrl(ApiUrl)
	url.Path = API_SEARCH_TAGGED

	q := url.Query()
	q.Add("tag_id", strconv.Itoa(tag_id))
	if page > 0 {
		q.Add("page", strconv.Itoa(page))
	}
	if sort != "" {
		q.Add("sort", sort)
	}
	url.RawQuery = q.Encode()

	return getAndUnmarshalTo[SearchResp](ctx, url.String(), nil)
}
