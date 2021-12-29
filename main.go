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

	for {
		select {
		case <-t.C:
			fmt.Println(user, "finished...")
			return
		case <-ctx.Done():
			fmt.Println(user, "canceled...")
			return
		default:
			fmt.Print(".")
			time.Sleep(200 * time.Millisecond)
		}
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
		fmt.Println(user, "stopping...")
		cancel()
	} else {
		fmt.Println(user, "not found (stopping)...")
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
			ctx, cf := context.WithTimeout(r.Context(), 2*time.Second)
			cnc[user] = cf
			h(w, r.Clone(ctx))
		} else {
			http.Error(w, "User unknown!", http.StatusInternalServerError)
			return
		}
	}
}

func addCancel(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(userKey).(string)
		if "" != user {
			fmt.Println(user, "add cancel for user ", user)
			ctx, cf := context.WithCancel(r.Context())
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
	http.HandleFunc("/limit", addUser(addTimeout(startHandler)))
	http.HandleFunc("/start", addUser(addCancel(startHandler)))
	http.HandleFunc("/stop", addUser(stopHandler))
	log.Fatal(http.ListenAndServe(":8000", nil))
}
