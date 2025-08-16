package apis

import (
	"net/http"
	"ponito/lilser/apis/helpers"
)

var toFile = func(w http.ResponseWriter, r *http.Request) (state int, err error) {
	http.Redirect(w, r, "/file?filename=.", http.StatusFound)
	return 200, nil
}

func UseIndex() {
	http.HandleFunc("/", helpers.Wrap(toFile))
}
