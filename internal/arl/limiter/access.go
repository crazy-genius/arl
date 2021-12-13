package limiter

import (
	"arl/internal/arl/api"
	"fmt"
	"net/http"
)

type JsonOverHTTP struct {
	router *http.ServeMux
	srv    Service
}

func NewJsonOverHttp(srv Service) *JsonOverHTTP {
	mux := http.NewServeMux()

	joh := &JsonOverHTTP{
		srv:    srv,
		router: mux,
	}

	mux.Handle("/", api.EnforceJSONMiddleware(http.HandlerFunc(joh.handle)))

	return joh
}

func (j *JsonOverHTTP) handle(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if len(key) == 0 {
		http.Error(w, "Empty key are not allowed", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		rateLimitHandler(key, r, w, j.srv)
		return
	case http.MethodPost:
		updateLimitHandler(key, r, w, j.srv)
		return
	default:
		http.Error(w, fmt.Sprintf("Method %s not allowed", r.Method), http.StatusMethodNotAllowed)
		return
	}
}

func rateLimitHandler(key string, r *http.Request, w http.ResponseWriter, srv Service) {
	var cnt uint8
	cnt, err := srv.Count(r.Context(), key, Hour)
	if err != nil {
		cnt = 0
	}

	if cnt > 10 {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("{\"message\": \"hour limit exceeded\"}")); err != nil {
			//TODO: write extra log
		}
		return
	}

	cnt, err = srv.Count(r.Context(), key, Second)
	if err != nil {
		cnt = 0
	}

	if cnt > 10 {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("{\"message\": \"seconds limit exceeded\"}")); err != nil {
			//TODO: write extra log
		}
		return
	}

	w.WriteHeader(http.StatusNotFound)
	if _, err := w.Write([]byte("{\"message\": \"key not found\"}")); err != nil {
		//TODO: write extra log
	}
}

func updateLimitHandler(key string, r *http.Request, w http.ResponseWriter, srv Service) {
	err := srv.Inc(r.Context(), key)
	if err != nil {
		http.Error(w, "Could not update storage", http.StatusInternalServerError)
		//TODO: write extra to log
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("{\"message\": \"ok\"}")); err != nil {
		// todo write log
	}
}

func (j *JsonOverHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	j.router.ServeHTTP(w, r)
}
