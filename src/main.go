package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"ponito/lilser/apis"
	"ponito/lilser/helper"
)

func printIp(port *int) {
	var (
		ip  net.IP
		err error
	)

	ip, err = helper.GetOutboundIp()
	if err != nil {
		return
	}

	log.Printf("Serving ip at %s:%d", ip, *port)
	log.Printf("Please visit http://%s:%d for files", ip, *port)
}

func main() {
	var (
		bin            = false
		port           = 80
		addr           = fmt.Sprintf("0.0.0.0:%d", port)
		useRevProxyApi = ""
		useFileApi     = false
		err            error
	)

	flag.StringVar(&useRevProxyApi, "rev-proxy", useRevProxyApi, "to do reverse proxy")
	flag.BoolVar(&useFileApi, "file", false, "use shared files")

	flag.BoolVar(&bin, "b", false, "run in bin directory")
	flag.IntVar(&port, "p", port, "the port to forward with")
	flag.Parse()

	if bin {
		log.Println("Running in bin directory")
		if err = os.Chdir("bin"); err != nil {
			panic(err)
		}
	}

	if useFileApi {
		var index = "/file?filename=."
		apis.UseIndex(index)
		apis.UseFile()
	}

	apis.UseProbe()

	if useRevProxyApi != "" {
		apis.UseReverseProxy(useRevProxyApi, port)
	}

	go printIp(&port)

	addr = fmt.Sprintf("0.0.0.0:%d", port)
	if err = http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
