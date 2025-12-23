# NHentai-go

NHentai access for go.

## 用法

### 开始

```bash
go get github.com/Miuzarte/NHentai-go
```

```go
package main
import nhentai "github.com/Miuzarte/NHentai-go"
```

### 设置下载并发数

```go
// 默认为 4
nhentai.SetThreads(4)
```

### 设置是否使用系统环境变量中的代理

```go
// 默认为 true
nhentai.SetUseEnvProxy(true)
```

### 镜像站替代

未经测试, 只是给了条路

```go
nhentai.ApiUrl = "https://nhentai.xxx"
```

### 自定义 cdn 主机名轮询

```go
// type HostProvider interface {
//     NextImageHost() string
//     NextThumbHost() string
// }
myHostProvider := nhentai.HostProvider(MyHostProvider{})
nhentai.SetCustomHostProvider(myHostProvider)
```

### 搜索 NHentai

```go
const keyword = "耳で恋した同僚〜オナサポ音声オタク女が同僚の声に反応してイキまくり〜"

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

page := 0
sort := "" // [nhentai.SORT_POPULAR] | [nhentai.SORT_DATE]
search, err := nhentai.Search(ctx, keyword, page, sort)
if err != nil {
    log.Fatalln(err)
}

results := search.Result

for _, gallery := range results {
    log.Println(gallery.Title)
    // all results are a complete gallery,
    // no more requests needed.
    // W for NHentai
    for image, err := range gallery.DownloadThumbsIter(ctx) {
        if err != nil {
            log.Fatalln(err)
        }
        log.Println(image.String())
    }
}

// download covers
for image, err := range results.DownloadCoversIter(ctx) {
    if err != nil {
        log.Fatalln(err)
    }
    log.Println(image.String())
}
```

### 下载画廊

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

const gId = 540994
gallery, err := nhentai.GetGallery(ctx, gId)
if err != nil {
    log.Fatalln(err)
}

for image, err := range gallery.DownloadPagesIter(ctx) {
    if err != nil {
        log.Println(err)
        break
    }
    log.Println(image.String())
}

// download thumbs
for image, err := range gallery.DownloadThumbsIter(ctx) {
    _, _ = image, err
}
```
