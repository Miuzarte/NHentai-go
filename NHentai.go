package NHentai

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Miuzarte/NHentai-go/api"
	"github.com/Miuzarte/NHentai-go/downloader"
)

const API_URL = "https://nhentai.net"

var apiUrl = API_URL

func SetApiUrl(url string) {
	if url != "" {
		apiUrl = url
	}
}

const DEFAULT_USER_AGENT = "NHentai-go/0.0.0 (https://github.com/Miuzarte/NHentai-go)"

// userAgent 设置请求的 User-Agent,
// 默认值符合 nhentai API 文档要求
var userAgent = DEFAULT_USER_AGENT

// SetUserAgent 设置请求的 User-Agent,
// 符合 nhentai API 文档要求
// 建议格式: `AppName/version (contact or project URL)`
func SetUserAgent(ua string) {
	userAgent = ua
}

// apiKey 用于 Authorization: Key <apiKey>,
// 留空则不发送该 header
var apiKey string

// SetApiKey 设置 API Key,用于 Authorization: Key <apiKey> header
// 留空则不发送该 header
func SetApiKey(key string) {
	apiKey = key
}

var threads = 4 // 下载并发数

// SetThreads 设置下载并发数
func SetThreads(n int) {
	threads = max(1, n)
	api.DownloadThreads = threads
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

var defaultHostProvider = &hostProvider{}

// 负载均衡到所有 cdn
var hp HostProvider = defaultHostProvider

// SetCustomHostProvider 自定义 cdn 主机名轮询
func SetCustomHostProvider(hostProvider HostProvider) {
	if hostProvider == nil {
		hp = defaultHostProvider
		return
	}
	hp = hostProvider
}

// Sort is a handy alias of [api.SearchGalleriesApiV2SearchGetParamsSort]
type Sort = string

const (
	// SORT_DATE is an alias of [api.SearchGalleriesApiV2SearchGetParamsSortDate]
	SORT_DATE Sort = "date" // "Recent"

	// SORT_POPULAR is an alias of [api.SearchGalleriesApiV2SearchGetParamsSortPopular]
	SORT_POPULAR Sort = "popular" // "Popular: all time"

	// SORT_POPULAR_MONTH is an alias of [api.SearchGalleriesApiV2SearchGetParamsSortPopularMonth]
	SORT_POPULAR_MONTH Sort = "popular-month" // "Popular: month"

	// SORT_POPULAR_TODAY is an alias of [api.SearchGalleriesApiV2SearchGetParamsSortPopularToday]
	SORT_POPULAR_TODAY Sort = "popular-today" // "Popular: today"

	// SORT_POPULAR_WEEK is an alias of [api.SearchGalleriesApiV2SearchGetParamsSortPopularWeek]
	SORT_POPULAR_WEEK Sort = "popular-week" // "Popular: week"
)

var apiClient *api.ClientWithResponses

// requestEditor 注入 User-Agent 与 ApiKey header
func requestEditor(ctx context.Context, req *http.Request) error {
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Key "+apiKey)
	}
	return nil
}

// initApiClient 创建全局 apiClient;init 时调用,也供 ApiUrl 变更后重建
func initApiClient() error {
	c, err := api.NewClientWithResponses(apiUrl,
		api.WithHTTPClient(&httpClient),
		api.WithRequestEditorFn(requestEditor),
	)
	if err != nil {
		return err
	}
	apiClient = c
	return nil
}

// ReinitClient 重建 apiClient,在修改 ApiUrl 后调用
func ReinitClient() error {
	return initApiClient()
}

func init() {
	// 注入 host provider 给 api 包
	// 闭包捕获 hp 变量, SetCustomHostProvider 更改 hp 后自动生效
	api.NextImageHostFn = func() string { return hp.NextImageHost() }
	api.NextThumbHostFn = func() string { return hp.NextThumbHost() }
	api.DownloadThreads = threads
	downloader.RequestEditor = requestEditor
	downloader.HttpClient = &httpClient
	if err := initApiClient(); err != nil {
		panic(err)
	}
}

// GetGallery gets gallery info by book id
func GetGallery(ctx context.Context, bookId int) (*api.GalleryDetailResponse, error) {
	resp, err := apiClient.GetGalleryApiV2GalleriesGalleryIdGetWithResponse(ctx, bookId, nil)
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status())
	}
	return resp.JSON200, nil
}

// Search searches
func Search(ctx context.Context, query string, page int, sort Sort) (*api.PaginatedResponseGalleryListItem, error) {
	var sortParam *api.SearchGalleriesApiV2SearchGetParamsSort
	if sort != "" {
		s := api.SearchGalleriesApiV2SearchGetParamsSort(sort)
		sortParam = &s
	}
	var pageParam *int
	if page > 0 {
		pageParam = &page
	}
	params := api.SearchGalleriesApiV2SearchGetParams{
		Query: query,
		Sort:  sortParam,
		Page:  pageParam,
	}
	resp, err := apiClient.SearchGalleriesApiV2SearchGetWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status())
	}
	return resp.JSON200, nil
}

// SearchTagged searches galleries by tag id
func SearchTagged(ctx context.Context, tagId int, page int, sort Sort) (*api.PaginatedResponseGalleryListItem, error) {
	var sortParam *api.GetGalleriesByTagApiV2GalleriesTaggedGetParamsSort
	if sort != "" {
		s := api.GetGalleriesByTagApiV2GalleriesTaggedGetParamsSort(sort)
		sortParam = &s
	}
	var pageParam *int
	if page > 0 {
		pageParam = &page
	}
	params := api.GetGalleriesByTagApiV2GalleriesTaggedGetParams{
		TagId: tagId,
		Sort:  sortParam,
		Page:  pageParam,
	}
	resp, err := apiClient.GetGalleriesByTagApiV2GalleriesTaggedGetWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status())
	}
	return resp.JSON200, nil
}

// LookupTags looks up tags by ids
func LookupTags(ctx context.Context, ids ...int) (api.Tags, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	strIds := make([]string, len(ids))
	for i, id := range ids {
		strIds[i] = strconv.Itoa(id)
	}
	params := api.GetTagsByIdsApiV2TagsIdsGetParams{
		Ids: strings.Join(strIds, ","),
	}
	resp, err := apiClient.GetTagsByIdsApiV2TagsIdsGetWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status())
	}
	return *resp.JSON200, nil
}

// GetRelated queries "More Like This" by gallery id
func GetRelated(ctx context.Context, galleryId int) (*api.RelatedGalleriesResponse, error) {
	resp, err := apiClient.GetRelatedGalleriesApiV2GalleriesGalleryIdRelatedGetWithResponse(ctx, galleryId)
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status())
	}
	return resp.JSON200, nil
}
