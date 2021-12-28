package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

type key int

const requestIDKey key = 0

func newContextWithID(ctx context.Context, req *http.Request) context.Context {
	ids, ok := req.URL.Query()["id"]
	var id string
	if ok {
		id = ids[0]
	}
	return context.WithValue(ctx, requestIDKey, id)
}

func getIDFromContext(ctx context.Context) string {
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
		ctx = newContextWithID(ctx, req)
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
	id := getIDFromContext(ctx)
	if id == "" {
		fmt.Fprintf(rw, "Start handler. Empty ID!\n")
	} else {
		fmt.Fprintf(rw, "Start handler. Request ID %s\n", id)
	}
}

func stopHandler(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	id := getIDFromContext(ctx)
	fmt.Fprintf(rw, "Stop handler. Request ID %s\n", id)
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
