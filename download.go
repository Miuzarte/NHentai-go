package NHentai

import (
	"context"
	"iter"
	"strconv"
	"sync"
)

type ImageType int

const (
	IMAGE_TYPE_UNKNOWN ImageType = iota

	IMAGE_TYPE_WEBP
	IMAGE_TYPE_JPEG
	IMAGE_TYPE_PNG
)

func (it ImageType) String() string {
	switch it {
	case IMAGE_TYPE_UNKNOWN:
		return "unknown"
	case IMAGE_TYPE_WEBP:
		return "webp"
	case IMAGE_TYPE_JPEG:
		return "jpeg"
	case IMAGE_TYPE_PNG:
		return "png"
	default:
		return ""
	}
}

type Image struct {
	Name string // "1.webp" | "1t.webp" | "cover.webp"
	P    int
	Data []byte
	Type ImageType
	// IsFromCache bool // TODO(maybe
}

func (p *Image) String() string {
	return p.Name + ": " + strconv.Itoa(len(p.Data))
}

type download struct {
	img *Image
	url string // 动态的 host, 不与 img 绑定
	err chan error
	// cache *cacheComic // TODO(maybe
}

func (d *download) start(ctx context.Context) {
	resp, body, err := get(ctx, d.url, nil)
	_ = resp
	if err == nil {
		d.img.Data = body
	}
	d.err <- err
}

func newCoversDownload(gs Gallerys) (dls []*download) {
	dls = make([]*download, len(gs))
	for i := range gs {
		dls[i] = &download{
			img: &Image{
				Name: gs[i].CoverFilename(),
				P:    i + 1,
			},
			url: gs[i].CoverUrl(),
			err: make(chan error, 1),
		}
	}
	return dls
}

func newThumbsDownload(g *Gallery) (dls []*download) {
	dls = make([]*download, len(g.Images.Pages))
	for i := range g.Images.Pages {
		dls[i] = &download{
			img: &Image{
				Name: g.ThumbFilename(i),
				P:    i + 1,
			},
			url: g.ThumbUrl(i),
			err: make(chan error, 1),
		}
	}
	return dls
}

func newPagesDownload(g *Gallery) (dls []*download) {
	dls = make([]*download, len(g.Images.Pages))
	for i := range g.Images.Pages {
		dls[i] = &download{
			img: &Image{
				Name: g.PageFilename(i),
				P:    i + 1,
			},
			url: g.PageUrl(i),
			err: make(chan error, 1),
		}
	}
	return dls
}

type downloader struct {
	ctx    context.Context
	cancel context.CancelFunc
	items  []*download
}

func newDownloader(ctx context.Context, dls []*download) *downloader {
	ctx, cancel := context.WithCancel(ctx)
	return &downloader{
		ctx:    ctx,
		cancel: cancel,
		items:  dls,
	}
}

func (dl *downloader) startBackground() {
	go func() {
		limiter := newLimiter()
		defer limiter.close()

		for _, item := range dl.items {
			select {
			case <-dl.ctx.Done():
				return
			case limiter.acquire() <- struct{}{}:
			}

			go func() {
				defer limiter.release()
				item.start(dl.ctx)
			}()
		}
	}()
}

func (dl *downloader) downloadIter() iter.Seq2[Image, error] {
	return func(yield func(Image, error) bool) {
		dl.startBackground()
		defer dl.cancel()
		for _, item := range dl.items {
			if !yield(*item.img, <-item.err) {
				return
			}
		}
	}
}

type limiter struct {
	sem  chan struct{}
	once sync.Once
}

func newLimiter() *limiter {
	return &limiter{
		sem: make(chan struct{}, threads),
	}
}

func (l *limiter) acquire() chan<- struct{} {
	return l.sem
}

func (l *limiter) release() {
	<-l.sem
}

func (l *limiter) close() {
	l.once.Do(func() {
		close(l.sem)
	})
}
