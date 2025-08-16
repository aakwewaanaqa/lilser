package helpers

import "net/http"

func Wrap(step func(w http.ResponseWriter, r *http.Request) (state int, err error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := step(w, r)
		if err != nil {
			w.WriteHeader(state)
			_, err = w.Write([]byte(err.Error()))
			return
		}

		if state == 0 {
			state = 200
		}

		w.WriteHeader(state)
		return
	}
}
