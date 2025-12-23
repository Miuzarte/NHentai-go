package NHentai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func get(ctx context.Context, url string, body io.Reader) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, body)
	if err != nil {
		return nil, nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, b, err
	}
	if resp.StatusCode != http.StatusOK {
		return resp, b, fmt.Errorf("unexpected http status code: %s", resp.Status)
	}
	return resp, b, nil
}

func unmarshalTo[T any](body []byte) (*T, error) {
	output := new(T)
	err := json.Unmarshal(body, &output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func getAndUnmarshalTo[T any](ctx context.Context, url string, body io.Reader) (*T, error) {
	_, b, err := get(ctx, url, body)
	if err != nil {
		return nil, err
	}
	t, err := unmarshalTo[T](b)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func toUrl(u string) *url.URL {
	url, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return url
}

func getFullType(t string) string {
	if t == "" {
		return ""
	}
	switch t[0] {
	case 'w':
		return "webp"
	case 'j':
		return "jpg"
	case 'p':
		return "png"
	default:
		panic(fmt.Errorf("[FIXME] unknown t: %s", t))
	}
}
