package apis

import (
	"net/http"
)

var toFile = func(w http.ResponseWriter, r *http.Request) (state int, err error) {
	http.Redirect(w, r, "/file?filename=.", http.StatusFound)
	return 200, nil
}

func UseIndex(url string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusFound)
	})
}
