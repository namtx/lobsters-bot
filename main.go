package main

import (
	"net/http"

	handler "github.com/namtx/rssbot/api"
)

func main() {
	http.HandleFunc("/", handler.Handler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
