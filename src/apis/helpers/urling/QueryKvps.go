package urling

import (
	"fmt"
	"net/url"
	"strings"
)

type QueryKvps struct {
	Map map[string][]string
}

func NewKvps(url *url.URL, mode QueryMode) (kvps QueryKvps, err error) {
	kvps = QueryKvps{Map: make(map[string][]string)}
	for k, v := range url.Query() {

		if mode == Normal {
			kvps.Map[k] = v
			continue
		}

		for _, v := range v {
			if mode&TrimSpace == TrimSpace {
				v = strings.TrimSpace(v)
			}

			if mode&NoEmpty == NoEmpty && v == "" {
				continue
			}

			if mode&ErrorEmpty == ErrorEmpty && v == "" {
				err = fmt.Errorf("empty query value for key %s", k)
				return
			}

			kvps.Map[k] = append(kvps.Map[k], v)
		}
	}
	return
}

func (kvps *QueryKvps) ErrIfNo(key string) (err error) {
	if _, ok := kvps.Map[key]; !ok {
		err = fmt.Errorf("missing query key %s", key)
		return
	}

	return
}

func (kvps *QueryKvps) ErrIfValContains(key, substr string) (err error) {
	var (
		vals []string
		ok   bool
		v    string
	)

	if vals, ok = kvps.Map[key]; !ok {
		err = fmt.Errorf("missing query key %s", key)
		return
	}

	for _, v = range vals {
		if strings.Contains(v, substr) {
			err = fmt.Errorf("query key %s contains %s", key, substr)
			return
		}
	}

	return
}

func (kvps *QueryKvps) Get(key string) (vals []string) {
	vals, _ = kvps.Map[key]
	return
}
