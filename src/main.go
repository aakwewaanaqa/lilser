package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := "8080"
	flag.StringVar(&port, "p", port, "port to listen on")
	flag.Parse()

	fileServer := http.FileServer(http.Dir("."))
	http.Handle("/", fileServer)

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Serving files in %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
