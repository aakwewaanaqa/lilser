package apis

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"ponito/lilser/helper"
	"sort"
	"strings"
	"sync"
	"time"
)

func getUserName() (username string, err error) {
	username = os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}

	if username == "" {
		username = os.Getenv("LOGNAME")
	}

	username = "default"
	err = fmt.Errorf("username not found in environment")
	return
}

var probeSelf = func(w http.ResponseWriter, r *http.Request) {
	var (
		username string
		err      error
	)

	if username, err = getUserName(); err != nil {
		log.Print(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte(username)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

//go:embed templates/probe_page.gohtml
var probeHtml string

var probeTmpl = template.Must(template.New("probePage").Parse(probeHtml))

var probeOthers = func(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	selfIP, _ := helper.GetOutboundIp()
	self := selfIP.String()
	base := ""
	if ip4 := selfIP.To4(); ip4 != nil {
		parts := strings.Split(ip4.String(), ".")
		if len(parts) == 4 {
			base = fmt.Sprintf("%s.%s.%s", parts[0], parts[1], parts[2])
		}
	}

	found := make([]string, 0, 16)
	if base != "" {
		client := &http.Client{Timeout: 350 * time.Millisecond}
		sem := make(chan struct{}, 64)
		var wg sync.WaitGroup
		var mu sync.Mutex
		for i := 1; i <= 254; i++ {
			host := fmt.Sprintf("%s.%d", base, i)
			if host == self {
				continue
			}
			wg.Add(1)
			sem <- struct{}{}
			go func(ip string) {
				defer wg.Done()
				defer func() { <-sem }()
				url := fmt.Sprintf("http://%s:1347/probe", ip)
				req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
				resp, err := client.Do(req)
				if err != nil {
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					mu.Lock()
					found = append(found, ip)
					mu.Unlock()
				}
			}(host)
		}
		wg.Wait()
		sort.Slice(found, func(i, j int) bool { return found[i] < found[j] })
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = probeTmpl.Execute(w, struct {
		IPs    []string
		SelfIP string
	}{
		IPs:    found,
		SelfIP: self,
	})
}

func UseProbe() (index string) {
	http.HandleFunc("GET /probe", probeSelf)
	http.HandleFunc("GET /probe/others", probeOthers)
	return "/probe/others"
}
