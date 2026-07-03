package NHentai

import (
	"sync/atomic"

	"github.com/Miuzarte/NHentai-go/internal/constant"
)

type HostProvider interface {
	NextImageHost() string
	NextThumbHost() string
	// MarkUnavailableImageHost(string)
	// MarkUnavailableThumbHost(string)
}

type hostProvider struct {
	imageCounter atomic.Uintptr
	thumbCounter atomic.Uintptr
}

func (up *hostProvider) NextImageHost() string {
	return constant.ImageHosts[int(up.imageCounter.Add(1))%len(constant.ImageHosts)]
}

func (up *hostProvider) NextThumbHost() string {
	return constant.ThumbHosts[int(up.thumbCounter.Add(1))%len(constant.ThumbHosts)]
}
