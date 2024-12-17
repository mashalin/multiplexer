package urlsservice

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/mashalin/multiplexer/internal/urls/dto"
)

const (
	MaxParallelURLs = 4
	TimeoutPerURL   = time.Second
)

type Service struct {
}

func New() *Service {
	return &Service{}
}

func (s *Service) FetchOne(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (s *Service) Fetch(ctx context.Context, urls []string) ([]dto.ResponseData, error) {
	mu := &sync.Mutex{}
	results := make([]dto.ResponseData, 0, len(urls))
	workers := make(chan struct{}, MaxParallelURLs)
	errOnce := &sync.Once{}
	wg := &sync.WaitGroup{}

	var errFinal error

	for _, url := range urls {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		workers <- struct{}{}
		wg.Add(1)

		go func(url string) {
			defer func() {
				<-workers
				wg.Done()
			}()

			reqCtx, cancel := context.WithTimeout(ctx, TimeoutPerURL)
			defer cancel()

			body, err := s.FetchOne(reqCtx, url)
			if err != nil {
				errOnce.Do(func() {
					errFinal = err
				})
				cancel()
				return
			}

			mu.Lock()
			results = append(results, dto.ResponseData{URL: url, Body: body})
			mu.Unlock()
		}(url)
	}

	wg.Wait()
	close(workers)

	if errFinal != nil {
		return nil, errFinal
	}

	return results, nil
}
