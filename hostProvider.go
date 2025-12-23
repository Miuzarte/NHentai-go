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
	i1 atomic.Uintptr
	i2 atomic.Uintptr
}

func (up *hostProvider) NextImageHost() string {
	return constant.ImageHosts[int(up.i1.Add(1))%len(constant.ImageHosts)]
}

func (up *hostProvider) NextThumbHost() string {
	return constant.ThumbHosts[int(up.i2.Add(1))%len(constant.ThumbHosts)]
}
