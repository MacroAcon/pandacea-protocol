package limits

import (
	"context"
	"net/http"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

func TokenBucket(r rate.Limit, b int) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(r, b)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if !limiter.Allow() {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			req = req.WithContext(context.WithValue(req.Context(), "req_id", uuid.NewString()))
			next.ServeHTTP(w, req)
		})
	}
}
