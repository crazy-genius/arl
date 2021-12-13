package main

import (
	"arl/internal/arl/api"
	"arl/internal/arl/limiter"
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for true {
			select {
			case sig := <-sigs:
				log.Printf("system call:%+v", sig)
				cancel()
			}
		}
	}()

	lru := limiter.NewInMemoryStorage()
	db := limiter.NewRedisStorage(redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}))
	limitSrv := limiter.NewService(lru, db)
	acc := limiter.NewJsonOverHttp(limitSrv)

	mux := http.NewServeMux()
	mux.Handle("/", acc)

	if err := api.Serve(mux, ctx); err != nil {
		log.Printf("failed to serve:+%v\n", err)
	}
}
