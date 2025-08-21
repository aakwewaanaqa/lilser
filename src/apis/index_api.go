package apis

import (
	"net/http"
)

func UseIndex(url string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusFound)
	})
}
