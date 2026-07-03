package api

import (
	"context"
	"iter"
	"path"
	"strconv"
	"strings"

	"github.com/Miuzarte/NHentai-go/downloader"
	"github.com/Miuzarte/NHentai-go/internal/utils"
)

var (
	NextImageHostFn func() string
	NextThumbHostFn func() string
)

// DownloadThreads 控制图片下载并发数
var DownloadThreads int

type (
	Tags             []TagResponse
	GalleryListItems []GalleryListItem
	TagSet           struct {
		Namespace string
		Tags      []string
	}
)

func (t TagResponse) String() string {
	return t.Type + ":" + t.Name
}

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

func (t Tags) Strings() []string {
	ss := make([]string, 0, len(t))
	for _, tag := range t {
		ss = append(ss, tag.String())
	}
	return ss
}

func (g *GalleryDetailResponse) pages() []PageInfo {
	if g == nil || g.Pages == nil {
		return nil
	}
	return *g.Pages
}

func (g *GalleryDetailResponse) PageFilename(i int) string {
	return path.Base(g.pages()[i].Path)
}

func (g *GalleryDetailResponse) PagePath(i int) string {
	return "/" + g.pages()[i].Path
}

func (g *GalleryDetailResponse) PageUrl(i int) string {
	p := g.PagePath(i)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return NextImageHostFn() + p
}

func (g *GalleryDetailResponse) PageUrls() []string {
	pages := g.pages()
	urls := make([]string, 0, len(pages))
	for i := range pages {
		urls = append(urls, g.PageUrl(i))
	}
	return urls
}

func (g *GalleryDetailResponse) PageUrlsIter() iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		pages := g.pages()
		for i := range pages {
			if !yield(i, g.PageUrl(i)) {
				return
			}
		}
	}
}

func (g *GalleryDetailResponse) ThumbFilename(i int) string {
	return path.Base(g.pages()[i].Thumbnail)
}

func (g *GalleryDetailResponse) ThumbPath(i int) string {
	return "/" + g.pages()[i].Thumbnail
}

func (g *GalleryDetailResponse) ThumbUrl(i int) string {
	p := g.ThumbPath(i)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return NextThumbHostFn() + p
}

func (g *GalleryDetailResponse) ThumbUrls() []string {
	pages := g.pages()
	urls := make([]string, 0, len(pages))
	for i := range pages {
		urls = append(urls, g.ThumbUrl(i))
	}
	return urls
}

func (g *GalleryDetailResponse) ThumbUrlsIter() iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		pages := g.pages()
		for i := range pages {
			if !yield(i, g.ThumbUrl(i)) {
				return
			}
		}
	}
}

func (g *GalleryDetailResponse) CoverFilename() string {
	return path.Base(g.Cover.Path)
}

func (g *GalleryDetailResponse) CoverPath() string {
	return "/" + g.Cover.Path
}

func (g *GalleryDetailResponse) CoverUrl() string {
	p := g.CoverPath()
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return NextThumbHostFn() + p
}

// DownloadThumbsIter downloads gallery thumbs using iterator
func (g *GalleryDetailResponse) DownloadThumbsIter(ctx context.Context) iter.Seq2[downloader.Image, error] {
	pages := g.pages()
	items := make([]*downloader.Download, len(pages))
	for i := range pages {
		items[i] = downloader.NewDownload(
			&downloader.Image{
				Name: g.ThumbFilename(i),
				P:    i + 1,
			},
			g.ThumbUrl(i),
		)
	}
	return downloader.DownloadIter(ctx, items, DownloadThreads)
}

// DownloadPagesIter downloads gallery images using iterator
func (g *GalleryDetailResponse) DownloadPagesIter(ctx context.Context) iter.Seq2[downloader.Image, error] {
	pages := g.pages()
	items := make([]*downloader.Download, len(pages))
	for i := range pages {
		items[i] = downloader.NewDownload(
			&downloader.Image{
				Name: g.PageFilename(i),
				P:    i + 1,
			},
			g.PageUrl(i),
		)
	}
	return downloader.DownloadIter(ctx, items, DownloadThreads)
}

func (g *GalleryListItem) ThumbFilename() string {
	return path.Base(g.Thumbnail)
}

func (g *GalleryListItem) ThumbPath() string {
	return "/" + g.Thumbnail
}

func (g *GalleryListItem) ThumbUrl() string {
	p := g.ThumbPath()
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return NextThumbHostFn() + p
}

// DownloadCoversIter downloads search result covers using iterator
func (gs GalleryListItems) DownloadCoversIter(ctx context.Context) iter.Seq2[downloader.Image, error] {
	items := make([]*downloader.Download, len(gs))
	for i := range gs {
		items[i] = downloader.NewDownload(
			&downloader.Image{
				Name: strconv.Itoa(gs[i].Id) + path.Ext(gs[i].Thumbnail),
				P:    i + 1,
			},
			gs[i].ThumbUrl(),
		)
	}
	return downloader.DownloadIter(ctx, items, DownloadThreads)
}
