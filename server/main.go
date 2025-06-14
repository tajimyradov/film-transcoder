package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("hls"))
	http.Handle("/", fs)

	log.Println("Serving at http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
