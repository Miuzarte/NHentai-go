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
	req.Header.Set("User-Agent", UserAgent)
	if ApiKey != "" {
		req.Header.Set("Authorization", "Key "+ApiKey)
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
	t := new(T)
	err := json.Unmarshal(body, &t)
	if err != nil {
		return nil, err
	}
	return t, nil
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

func unmarshalToSlice[S ~[]E, E any](body []byte) (S, error) {
	s := make(S, 0)
	err := json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func getAndUnmarshalToSlice[S ~[]E, E any](ctx context.Context, url string, body io.Reader) (S, error) {
	_, b, err := get(ctx, url, body)
	if err != nil {
		return nil, err
	}
	s, err := unmarshalToSlice[S](b)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func toUrl(u string) *url.URL {
	url, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return url
}
