package downloader

import (
	"context"
	"fmt"
	"io"
	"iter"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// RequestEditor 由 NHentai 包在 init 时设置,
// 用于注入 User-Agent 和 ApiKey header
var RequestEditor func(ctx context.Context, req *http.Request) error

// HttpClient 由 NHentai 包在 init 时设置,
// 复用调优过的 http.Client
var HttpClient *http.Client

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

func ExtToImageType(ext string) ImageType {
	switch strings.ToLower(strings.TrimLeft(ext, ".")) {
	case "webp":
		return IMAGE_TYPE_WEBP
	case "jpg", "jpeg":
		return IMAGE_TYPE_JPEG
	case "png":
		return IMAGE_TYPE_PNG
	default:
		return IMAGE_TYPE_UNKNOWN
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

// download 表示一个下载任务
type download struct {
	img *Image
	url string // 动态的 host, 不与 img 绑定
	err chan error
	// cache *cacheComic // TODO(maybe
}

func (d *download) start(ctx context.Context) {
	body, err := get(ctx, d.url)
	if err == nil {
		d.img.Data = body
	}
	d.err <- err
}

// get 下载图片字节,
// 通过 RequestEditor 注入 header,
// 通过 HttpClient 发送
func get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if RequestEditor != nil {
		if err := RequestEditor(ctx, req); err != nil {
			return nil, err
		}
	}
	if HttpClient == nil {
		return nil, fmt.Errorf("downloader.HttpClient is nil")
	}
	resp, err := HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http status code: %s", resp.Status)
	}
	return b, nil
}

// NewDownload 创建一个下载任务(供 api 包构造 []*Download)
type Download = download

func NewDownload(img *Image, url string) *Download {
	return &Download{
		img: img,
		url: url,
		err: make(chan error, 1),
	}
}

// DownloadImage 返回下载结果(供 api 包从 *Download 取结果)
func (d *Download) Image() *Image { return d.img }
func (d *Download) Wait() error   { return <-d.err }

type downloader struct {
	ctx    context.Context
	cancel context.CancelFunc
	items  []*Download
}

func NewDownloader(ctx context.Context, items []*Download) *downloader {
	ctx, cancel := context.WithCancel(ctx)
	return &downloader{
		ctx:    ctx,
		cancel: cancel,
		items:  items,
	}
}

func (dl *downloader) startBackground(threads int) {
	go func() {
		limiter := newLimiter(threads)
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

func (dl *downloader) downloadIter(threads int) iter.Seq2[Image, error] {
	return func(yield func(Image, error) bool) {
		dl.startBackground(threads)
		defer dl.cancel()
		for _, item := range dl.items {
			if !yield(*item.img, <-item.err) {
				return
			}
		}
	}
}

// DownloadIter 启动并发下载并返回迭代器,threads 控制并发数
func DownloadIter(ctx context.Context, items []*Download, threads int) iter.Seq2[Image, error] {
	return NewDownloader(ctx, items).downloadIter(threads)
}

type limiter struct {
	sem  chan struct{}
	once sync.Once
}

func newLimiter(threads int) *limiter {
	return &limiter{
		sem: make(chan struct{}, max(1, threads)),
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
