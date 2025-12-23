package usage

import nhentai "github.com/Miuzarte/NHentai-go"

func UsageSetCustomHostProvider() {
	myHostProvider := nhentai.HostProvider(MyHostProvider{})
	nhentai.SetCustomHostProvider(myHostProvider)
}

//	type HostProvider interface {
//	    NextImageHost() string
//	    NextThumbHost() string
//	}
//
// 简单实现 仅作示例
type MyHostProvider struct{}

func (hp MyHostProvider) NextImageHost() string {
	return "https://i.nhentai.net"
}

func (hp MyHostProvider) NextThumbHost() string {
	return "https://t.nhentai.net"
}
