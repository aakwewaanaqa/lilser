package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"ponito/lilser/helper"
)

func printIp() {
	var (
		ip  net.IP
		err error
	)

	ip, err = helper.GetOutboundIp()
	if err != nil {
		return
	}

	log.Println("Your outbound ip is", ip)
}

func main() {
	var (
		port int
		addr string
	)

	flag.IntVar(&port, "p", 8080, "the port to forward with")
	flag.Parse()

	fileServer := http.FileServer(http.Dir("."))
	http.Handle("/", fileServer)

	addr = fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("Serving files in %s", addr)

	go printIp()

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
