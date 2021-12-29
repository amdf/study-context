package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type contextKey int

const userKey = contextKey(1)

var cnc map[string]context.CancelFunc

func someFunc(w io.Writer, user string, ctx context.Context) {
	t := time.NewTimer(10 * time.Second)

	select {
	case <-t.C:
		fmt.Println(user, "finished...")
	case <-ctx.Done():
		fmt.Println(user, "canceled...")
	}
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userKey).(string)

	someFunc(w, user, r.Context())
}

func stopHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userKey).(string)
	cancel, ok := cnc[user]
	if ok {
		cancel()
	}
}

func addUser(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(w, r.Clone(context.WithValue(r.Context(), userKey, r.FormValue("user"))))
	}
}

func addTimeout(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(userKey).(string)
		if "" != user {
			fmt.Println(user, "add timeout for user ", user)
			ctx, cf := context.WithTimeout(r.Context(), time.Second)
			cnc[user] = cf
			h(w, r.Clone(ctx))
		} else {
			http.Error(w, "User unknown!", http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	cnc = make(map[string]context.CancelFunc)
	fmt.Println("start")
	http.HandleFunc("/start", addUser(addTimeout(startHandler)))
	http.HandleFunc("/stop", addUser(stopHandler))
	log.Fatal(http.ListenAndServe(":8000", nil))
}
