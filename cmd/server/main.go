package main

import (
	"log"
	"net/http"
)

func main() {
	port := ":8000"
	http.Handle("/", http.FileServer(http.Dir(".")))

	log.Printf("Serving Oinakos WASM on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
