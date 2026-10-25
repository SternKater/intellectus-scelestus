package sample

import (
	"context"
	"log"
	"time"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func sample_context() {

	parentCtx := context.Background()
	
// 1. Cancellation
	ctx, cancel := context.WithCancel(parentCtx)

// 2. deadline
	ctx, cancel = context.WithDeadline(parentCtx, time.Now().Add(5 * time.Second))

// 3. timeout
	ctx, cancel = context.WithTimeout(parentCtx, 2 * time.Second)

// 4. Key-Value
	key := "key"
	value := "value"
	ctx = context.WithValue(parentCtx, key, value)

// avoid compiler's error
	log.Println(ctx)	
	cancel()

}

func sample_query_db(ctx context.Context) ([]string, error) {
// sample cahnnel
	ch := make(chan []string, 1)

    go func() {
        time.Sleep(500 * time.Millisecond) // heavy query
        ch <- []string{"User1", "User2"}
    }()

    select {
    case <-ctx.Done():
        // cancel / timeout
        return nil, ctx.Err() // context.Canceled / context.DeadlineExceeded
    case res := <-ch:
        return res, nil
    }
}

func main() {
    server := &http.Server{Addr: ":8080"}

    // Ctrl+C, kill
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("Error due start: %v", err)
        }
    }()

    <-stop // awaiting for stop from OS
    log.Println("Shutdown detected...")

    // Graceful shutdown context
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Shutdown stop receiving new requests and wait for stopping exist requests for 5 sec
    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Fatalf("Error graceful shutdown: %v", err)
    }
    log.Println("Server has stopped successfully.")
}
