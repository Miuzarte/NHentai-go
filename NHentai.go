package NHentai

import (
	"context"
	"iter"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/Miuzarte/NHentai-go/internal/utils"
)

const API_URL = "https://nhentai.net"

var ApiUrl = API_URL

var defaultHostProvider = &hostProvider{}

// 负载均衡到所有 cdn
var hp HostProvider = defaultHostProvider

const (
	API_SEARCH        = "/api/v2/search"
	API_SEARCH_TAGGED = "/api/v2/galleries/tagged" // [TODO] figure out v2 api path
	API_GALLERY       = "/api/v2/galleries"        // /{id}
	API_TAG_LOOKUP    = "/api/v2/tags/ids"         // ?ids=1,2,3

	// API_TAG_TYPE = "/api/v2/tags" // /{type}?sort=name
)

type Sort = string

const (
	SORT_DATE          Sort = "date"          // "Recent"
	SORT_POPULAR       Sort = "popular"       // "Popular: all time"
	SORT_POPULAR_WEEK  Sort = "popular-week"  // "Popular: week"
	SORT_POPULAR_TODAY Sort = "popular-today" // "Popular: today"
)

var threads = 4 // 下载并发数

// SetThreads 设置下载并发数
func SetThreads(n int) {
	threads = max(1, n)
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

// SetCustomHostProvider 自定义 cdn 主机名轮询
func SetCustomHostProvider(hostProvider HostProvider) {
	if hostProvider == nil {
		hp = defaultHostProvider
		return
	}
	hp = hostProvider
}

// [TODO] 尝试兼容 [EHentai.TranslateMulti]
// 处理 "tag" 域
type Tag struct {
	Id    int    `json:"id"`
	Type  string `json:"type"` // "tag" | "language" | "category" | "parody" | "artist" | "group" ... // [TODO] TagSet type
	Name  string `json:"name"`
	Slug  string `json:"slug"` // == {.Name} ?
	Url   string `json:"url"`  // "/{.Type}/{.Name}"
	Count int    `json:"count"`

	Description *string `json:"description"` // only in [API_TAG_LOOKUP]

	// IsCommunity *any `json:"is_community"`
	// PendingDescribeId *any `json:"pending_describe_id"`
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
	s := utils.Set[string]{}
	return s.Clean(namespaces)
}

func (t Tags) Set() (ts []TagSet) {
	setPos := map[string]int{}
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

func LookupTags(ctx context.Context, ids ...int) (Tags, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	url := toUrl(ApiUrl)
	url.Path = API_TAG_LOOKUP

	q := url.Query()
	for _, id := range ids {
		q.Add("ids", strconv.Itoa(id))
	}
	url.RawQuery = q.Encode()

	return getAndUnmarshalToSlice[Tags](ctx, url.String(), nil)
}

type ImageInfo struct {
	T string `json:"t"` // type // "w"
	W int    `json:"w"`
	H int    `json:"h"`
}

type ImageInfoBase struct {
	Path   string `json:"path"` // "galleries/{MediaId}/cover.webp" | "galleries/3138775/thumb.webp"
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Page struct {
	Number          int    // 1
	Path            string // "galleries/{MediaId}/1.webp"
	Width           int
	Height          int
	Thumbnail       string // "galleries/{MediaId}/1t.webp"
	ThumbnailWidth  int
	ThumbnailHeight int
}

type Gallery struct {
	Id      int    `json:"id"`
	MediaId string `json:"media_id"`
	Title   struct {
		English  string `json:"english"`
		Japanese string `json:"japanese"`
		Pretty   string `json:"pretty"`
	} `json:"title"`
	Cover        ImageInfoBase `json:"cover"`
	Thumbnail    ImageInfoBase `json:"thumbnail"`
	Scanlator    string        `json:"scanlator"`   // scan + translator
	UploadDate   int           `json:"upload_date"` // ts in second
	Tags         Tags          `json:"tags"`
	NumPages     int           `json:"num_pages"`
	NumFavorites int           `json:"num_favorites"`
	Pages        []Page        `json:"pages"`
}

/*
type Gallerys []*Gallery

// DownloadCoversIter downloads search result covers using iterator
func (gs Gallerys) DownloadCoversIter(ctx context.Context) iter.Seq2[Image, error] {
	return newDownloader(ctx, newCoversDownload(gs)).downloadIter()
}
*/

// DownloadThumbsIter downloads gallery thumbs using iterator
func (g *Gallery) DownloadThumbsIter(ctx context.Context) iter.Seq2[Image, error] {
	return newDownloader(ctx, g.newThumbsDownload()).downloadIter()
}

// DownloadPagesIter downloads gallery images using iterator
func (g *Gallery) DownloadPagesIter(ctx context.Context) iter.Seq2[Image, error] {
	return newDownloader(ctx, g.newPagesDownload()).downloadIter()
}

// GetGallery gets gallery info by book id
func GetGallery(ctx context.Context, bookId int) (*Gallery, error) {
	url := toUrl(ApiUrl)
	url.Path = path.Join(API_GALLERY, strconv.Itoa(bookId))

	return getAndUnmarshalTo[Gallery](ctx, url.String(), nil)
}

// count from 0
func (g *Gallery) PageFilename(i int) string {
	// return strconv.Itoa(i+1) + "." + getFullType(g.Images.Pages[i].T)
	return path.Base(g.Pages[i].Path)
}

func (g *Gallery) PagePath(i int) string {
	// return path.Join("/galleries", g.MediaId, g.PageFilename(i))
	// 保持原先行为, 带上前导斜杠
	return "/" + g.Pages[i].Path
}

// PageUrl does not provide a fixed "host"
//
// Consider using [Gallery.PagePath]
func (g *Gallery) PageUrl(i int) string {
	path := g.PagePath(i)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return hp.NextImageHost() + path
}

// PageUrlsIter does not provide a fixed "host"
//
// Consider using [Gallery.PagePath]
func (g *Gallery) PageUrlsIter() iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		for i := range g.Pages {
			if !yield(i, g.PageUrl(i)) {
				return
			}
		}
	}
}

// PageUrls does not provide a fixed "host"
//
// Consider using [Gallery.PagePath]
func (g *Gallery) PageUrls() []string {
	urls := make([]string, 0, len(g.Pages))
	for i := range g.Pages {
		urls = append(urls, g.PageUrl(i))
	}
	return urls
}

func (g *Gallery) ThumbFilename(i int) string {
	// return strconv.Itoa(i+1) + "t" + "." + getFullType(g.Images.Pages[i].T)
	return path.Base(g.Pages[i].Thumbnail)
}

func (g *Gallery) ThumbPath(i int) string {
	// return path.Join("/galleries", g.MediaId, g.ThumbFilename(i))
	return "/" + g.Pages[i].Thumbnail
}

// ThumbUrl does not provide a fixed "host"
//
// Consider using [Gallery.ThumbPath]
func (g *Gallery) ThumbUrl(i int) string {
	path := g.ThumbPath(i)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return hp.NextThumbHost() + path
}

// ThumbUrlsIter does not provide a fixed "host"
//
// Consider using [Gallery.ThumbPath]
func (g *Gallery) ThumbUrlsIter() iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		for i := range g.Pages {
			if !yield(i, g.ThumbUrl(i)) {
				return
			}
		}
	}
}

// ThumbUrls does not provide a fixed "host"
//
// Consider using [Gallery.ThumbPath]
func (g *Gallery) ThumbUrls() []string {
	urls := make([]string, 0, len(g.Pages))
	for i := range g.Pages {
		urls = append(urls, g.ThumbUrl(i))
	}
	return urls
}

func (g *Gallery) CoverFilename() string {
	// return "cover" + "." + getFullType(g.Images.Cover.T)
	return path.Base(g.Cover.Path)
}

func (g *Gallery) CoverPath() string {
	// return path.Join("/galleries", g.MediaId, g.CoverFilename())
	return "/" + g.Cover.Path
}

// CoverUrl does not provide a fixed "host"
//
// Consider using [Gallery.CoverPath]
func (g *Gallery) CoverUrl() string {
	path := g.CoverPath()
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return hp.NextThumbHost() + path
}

// GetRelated querys "More Like This"
func (g *Gallery) GetRelated(ctx context.Context) (GallerySearchResults, error) {
	type response struct {
		Result GallerySearchResults `json:"result"`
	}

	url := toUrl(ApiUrl)
	url.Path = path.Join(API_GALLERY, strconv.Itoa(g.Id), "related")

	r, err := getAndUnmarshalTo[response](ctx, url.String(), nil)
	if err != nil {
		return nil, err
	}
	return r.Result, nil
}

type GallerySearchResult struct {
	Id              int    `json:"id"`
	MediaId         string `json:"media_id"`
	EnglishTitle    string `json:"english_title"`
	JapaneseTitle   string `json:"japanese_title"`
	Thumbnail       string `json:"thumbnail"` // "galleries/3138775/thumb.webp"
	ThumbnailWidth  int    `json:"thumbnail_width"`
	ThumbnailHeight int    `json:"thumbnail_height"`
	NumPages        int    `json:"num_pages"`
	NumFavorites    int    `json:"num_favorites"`
	TagIds          []int  `json:"tag_ids"`
	Blacklisted     bool   `json:"blacklisted"`
}

func (g *GallerySearchResult) ThumbFilename() string {
	return path.Base(g.Thumbnail)
}

func (g *GallerySearchResult) ThumbPath() string {
	return "/" + g.Thumbnail
}

func (g *GallerySearchResult) ThumbUrl() string {
	path := g.ThumbPath()
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return hp.NextThumbHost() + path
}

func (g *GallerySearchResult) LookupTags(ctx context.Context) (Tags, error) {
	return LookupTags(ctx, g.TagIds...)
}

type GallerySearchResults []GallerySearchResult

// DownloadCoversIter downloads search result covers using iterator
func (gs GallerySearchResults) DownloadCoversIter(ctx context.Context) iter.Seq2[Image, error] {
	return newDownloader(ctx, gs.newThumbsDownload()).downloadIter()
}

type SearchResp struct {
	Result   GallerySearchResults `json:"result"`
	NumPages int                  `json:"num_pages"`
	PerPage  int                  `json:"per_page"`
	Total    int                  `json:"total"`
}

// Search searches
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

// Deprecated: [TODO] figure out v2 api path
func SearchTagged(ctx context.Context, tagId int, page int, sort Sort) (*SearchResp, error) {
	url := toUrl(ApiUrl)
	url.Path = API_SEARCH_TAGGED

	q := url.Query()
	q.Add("tag_id", strconv.Itoa(tagId))
	if page > 0 {
		q.Add("page", strconv.Itoa(page))
	}
	if sort != "" {
		q.Add("sort", sort)
	}
	url.RawQuery = q.Encode()

	return getAndUnmarshalTo[SearchResp](ctx, url.String(), nil)
}
