package api

import (
	"context"
	"log"
	"mime"
	"net/http"
	"os"
	"time"
)

// Serve start http server
func Serve(ctx context.Context, mux *http.ServeMux) (err error) {

	var port string
	if port = os.Getenv("port"); port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err = srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("listen:%s\n", err)
		}
	}()

	log.Printf("server started http://localhost:%s", port)

	<-ctx.Done()

	log.Printf("server stopped")

	shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(shutDownCtx); err != nil {
		log.Fatalf("server Shutdown Failed:%s", err)
	}

	log.Printf("server exited properly")

	if err == http.ErrServerClosed {
		err = nil
	}

	return err
}

// EnforceJSONMiddleware adds proper JSON headers
func EnforceJSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")

		mt, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			http.Error(w, "Malformed Content-Type header", http.StatusBadRequest)
			return
		}

		if mt != "application/json" {
			http.Error(w, "Content-Type header must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		w.Header().Add("Content-type", "application/json")

		next.ServeHTTP(w, r)
	})
}
