package NHentai

import (
	"sync/atomic"

	"github.com/Miuzarte/NHentai-go/internal/constant"
)

type UrlProviderType interface {
	NextImageUrl() string
	NextThumbUrl() string
	// MarkUnavailableImageUrl(string)
	// MarkUnavailableThumbUrl(string)
}

type urlProvider struct {
	i1 atomic.Uintptr
	i2 atomic.Uintptr
}

func (up *urlProvider) NextImageUrl() string {
	return constant.ImageUrls[int(up.i1.Add(1))%len(constant.ImageUrls)]
}

func (up *urlProvider) NextThumbUrl() string {
	return constant.ThumbUrls[int(up.i2.Add(1))%len(constant.ThumbUrls)]
}
