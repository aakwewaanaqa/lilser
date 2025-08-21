package apis

import "net/http"

var probe = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func UseProbe() {
	http.HandleFunc("/probe/lilser", probe)
}
