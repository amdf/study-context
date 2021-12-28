package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

type key int

const requestIDKey key = 0

func newContextWithRequestID(ctx context.Context, req *http.Request) context.Context {
	fmt.Println(req.URL)
	return context.WithValue(ctx, requestIDKey, req.Header.Get("X-Request-ID"))
}

func requestIDFromContext(ctx context.Context) string {
	return ctx.Value(requestIDKey).(string)
}

type ContextHandler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request)
}

type ContextHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

func (h ContextHandlerFunc) ServeHTTPContext(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	h(ctx, rw, req)
}

func middleware(h ContextHandler) ContextHandler {
	return ContextHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		ctx = newContextWithRequestID(ctx, req)
		h.ServeHTTPContext(ctx, rw, req)
	})
}

type ContextAdapter struct {
	ctx     context.Context
	handler ContextHandler
}

func (ca *ContextAdapter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ca.handler.ServeHTTPContext(ca.ctx, rw, req)
}

func startHandler(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	reqID := requestIDFromContext(ctx)
	fmt.Fprintf(rw, "Start handler. Request ID %s\n", reqID)
}

func stopHandler(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	reqID := requestIDFromContext(ctx)
	fmt.Fprintf(rw, "Stop handler. Request ID %s\n", reqID)
}

func main() {
	addr := ":8000"
	start := &ContextAdapter{
		ctx:     context.Background(),
		handler: middleware(ContextHandlerFunc(startHandler)),
	}
	stop := &ContextAdapter{
		ctx:     context.Background(),
		handler: middleware(ContextHandlerFunc(stopHandler)),
	}
	mux := http.NewServeMux()

	mux.Handle("/start", start)
	mux.Handle("/stop", stop)

	server := http.Server{Addr: addr, Handler: mux}
	fmt.Println("starting server at", addr)
	log.Fatal(server.ListenAndServe())
}

// func main() {
// http.HandleFunc("/start", startHandler)
// http.HandleFunc("/stop", stopHandler)
// http.HandleFunc("/limited", limitedHandler)

// log.Fatal(http.ListenAndServe("localhost:8000", nil))
// }

// func startHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "URL.Path = %q\n", r.URL.Path)
// }

// func stopHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "URL.Path = %q\n", r.URL.Path)
// }

// func limitedHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "URL.Path = %q\n", r.URL.Path)
// }
