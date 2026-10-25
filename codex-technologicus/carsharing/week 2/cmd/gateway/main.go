package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

// lua-script for Fixed Window
// Indicies start from 1! KEYS[1], ARGV[1]...
const fixedWindowLua = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = redis.call("INCR", key)

if current == 1 then
    redis.call("EXPIRE", key, window)
end

if current > limit then
    return 0 -- limit exceeded
end

return 1 -- do what you want
`

type RateLimiterMiddleware struct {
	rdb       *redis.Client
	scriptSHA string
	limit     int
	windowMs  int
}

func NewRateLimiter(rdb *redis.Client, limit int, windowMs int) (*RateLimiterMiddleware, error) {
	// compile into SHA-1 hash
	sha, err := rdb.ScriptLoad(context.Background(), fixedWindowLua).Result()
	if err != nil {
		return nil, fmt.Errorf("Lua-not loaded(for some reason): %w", err)
	}
	log.Printf("Lua-script totally hashed into Redis. SHA: %s", sha)
	return &RateLimiterMiddleware{
		rdb:       rdb,
		scriptSHA: sha,
		limit:     limit,
		windowMs:  windowMs,
	}, nil
}

// Middleware
func (m *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UnixNano() / int64(time.Millisecond)
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr // ops
		}
		key := "rate:fixed:" + host
		// w.WriteHeader(http.StatusOK)
		// w.Write([]byte(key))

		// 200ms timeout. otherwise we just skipping this step
		redisCtx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()

		res, err := m.rdb.EvalSha(redisCtx, m.scriptSHA, []string{key}, now, m.limit, m.windowMs).Result()
		if err != nil {
			log.Printf("Internal Error (Redis EvalSha): %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// 0 - limit exseeded
		if res.(int64) == 0 {
			w.Header().Set("Retry-After", strconv.Itoa(m.windowMs/1000))
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("429 Too Many Requests. Fixed window blocked the req.\n"))
			return // Skip the chain
		}

		// do next
		next.ServeHTTP(w, r)
	})
}

// simulate next step
func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hi there!\n"))
}

func main() {
	// 1. connect
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// check connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("No Redis entuty found: %v", err)
	}

	// 2. RateLimiter(3 due 10s)
	limiter, err := NewRateLimiter(rdb, 3, 1000)
	if err != nil {
		log.Fatalf("Error on init: %v", err)
	}

	// 3. crteate interceptors chain
	mux := http.NewServeMux()

	mux.Handle("/", limiter.Handler(http.HandlerFunc(helloHandler)))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// 4. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Server started on http://localhost:8080")
		log.Println("Limit: 3 reqs in 10 sec(FixedWindow)")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop // Ctrl+C
	log.Println("Interrupt signal has received...")

	// 5 sec timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error due graceful shutdown: %v", err)
	}
	log.Println("Alles Gut.")
}
