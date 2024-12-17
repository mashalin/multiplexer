package rest

import (
	"net/http"
)

const MaxConcurrentReq = 100

var semaphore = make(chan struct{}, MaxConcurrentReq)

func LimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case semaphore <- struct{}{}:
			defer func() { <-semaphore }()
			next(w, r)
		default:
			w.WriteHeader(http.StatusTooManyRequests)
		}
	}
}
