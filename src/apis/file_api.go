package apis

import (
	_ "embed"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"ponito/lilser/apis/helpers"
	"ponito/lilser/apis/helpers/urling"
	"sort"
	"strings"
)

var (
	//go:embed templates/dir_page.gohtml
	dirPage string
	//go:embed templates/editor_page.gohtml
	editorPage string
	dirTmpl    = template.Must(template.New("dirPage").Parse(dirPage))
	editorTmpl = template.Must(template.New("editorPage").Parse(editorPage))
)

func printWorkingDir() {
	var (
		wd  string
		err error
	)

	if wd, err = os.Getwd(); err != nil {
		return
	}

	if wd, err = filepath.Abs(wd); err != nil {
		return
	}

	log.Printf("Serving files at: %s\n", wd)
}

var putFile = func(w http.ResponseWriter, r *http.Request) (state int, err error) {
	const (
		FileName = "filename"
	)

	var (
		mode     = urling.TrimSpace | urling.ErrorEmpty
		queries  urling.QueryKvps
		folder   string
		filename string
		tmpFile  *os.File
	)

	// query parsing
	{
		if queries, err = urling.NewKvps(r.URL, mode); err != nil {
			state = 400
			return
		}

		if err = queries.ErrIfNo(FileName); err != nil {
			state = 400
			return
		}

		if err = queries.ErrIfValContains(FileName, ".."); err != nil {
			state = 400
			return
		}
	}

	// file stat checking
	{
		filename = queries.Get(FileName)[0]
		folder = filepath.Dir(filename)
		if _, err = os.Stat(folder); err != nil {
			if os.IsNotExist(err) {
				if err = os.MkdirAll(folder, 0755); err != nil {
					state = 500
					return
				}
			} else if os.IsPermission(err) {
				state = 403
				return
			}
		}

		if _, err = os.Stat(filename); err != nil {
			if os.IsPermission(err) {
				state = 403
				return
			}
		}
	}

	// atomic file replacing
	{
		if tmpFile, err = os.CreateTemp(folder, "tmp-*"); err != nil {
			state = 500
			return
		}

		if _, err = io.Copy(tmpFile, r.Body); err != nil {
			_ = tmpFile.Close()
			state = 500
			return
		}

		if err = tmpFile.Close(); err != nil {
			state = 500
			return
		}

		if err = os.Rename(tmpFile.Name(), filename); err != nil {
			state = 500
			return
		}
	}
	return
}

var getFile = func(w http.ResponseWriter, r *http.Request) (state int, err error) {
	const (
		FileName = "filename"
	)

	var (
		mode     = urling.TrimSpace | urling.ErrorEmpty
		queries  urling.QueryKvps
		filename string
		fileinfo os.FileInfo
	)

	if queries, err = urling.NewKvps(r.URL, mode); err != nil {
		state = 400
		return
	}

	if err = queries.ErrIfNo(FileName); err != nil {
		state = 400
		return
	}

	if err = queries.ErrIfValContains(FileName, ".."); err != nil {
		state = 400
		return
	}

	filename = queries.Get(FileName)[0]
	if fileinfo, err = os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			state = 404
			return
		} else if os.IsPermission(err) {
			state = 403
			return
		}
	}

	if fileinfo.IsDir() {
		var (
			entries []os.DirEntry
		)

		if entries, err = os.ReadDir(filename); err != nil {
			state = 500
			return
		}

		// build view model and sort entries by name
		type dirEntryView struct {
			Name  string
			IsDir bool
		}
		views := make([]dirEntryView, 0, len(entries))
		for _, e := range entries {
			views = append(views, dirEntryView{Name: e.Name(), IsDir: e.IsDir()})
		}
		sort.SliceStable(views, func(i, j int) bool { return views[i].Name < views[j].Name })

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = dirTmpl.Execute(w, struct {
			Path    string
			Entries []dirEntryView
		}{
			Path:    filename,
			Entries: views,
		})
	} else {
		// When browser navigates (Accept: text/html), and the file looks like text/code,
		// serve the editor page. Otherwise, serve the raw file.
		if strings.Contains(r.Header.Get("Accept"), "text/html") && isTextLike(filename) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = editorTmpl.Execute(w, nil)
			return
		}
		http.ServeFile(w, r, filename)
	}
	return
}

// minimal text/code extension set used to decide when to show the editor UI
var textExts = map[string]struct{}{
	".txt": {}, ".md": {}, ".markdown": {}, ".log": {},
	".json": {}, ".jsonc": {}, ".yaml": {}, ".yml": {}, ".toml": {}, ".ini": {}, ".conf": {}, ".config": {},
	".csv": {}, ".tsv": {}, ".sql": {},
	".js": {}, ".mjs": {}, ".cjs": {}, ".ts": {}, ".tsx": {}, ".jsx": {}, ".css": {}, ".scss": {}, ".less": {}, ".html": {}, ".htm": {}, ".xml": {},
	".go": {}, ".py": {}, ".rb": {}, ".rs": {}, ".java": {}, ".kt": {}, ".kts": {}, ".c": {}, ".h": {}, ".cpp": {}, ".cc": {}, ".hpp": {}, ".cs": {},
	".sh": {}, ".bash": {}, ".zsh": {}, ".fish": {}, ".ps1": {}, ".bat": {},
}

func isTextLike(name string) bool {
	base := filepath.Base(name)
	if base == "Makefile" || base == "Dockerfile" { // common text files without extensions
		return true
	}
	ext := strings.ToLower(filepath.Ext(base))
	if ext == "" {
		return false
	}
	_, ok := textExts[ext]
	return ok
}

func UseFile() (index string) {
	go printWorkingDir()

	http.HandleFunc("PUT /file", helpers.Wrap(putFile))
	http.HandleFunc("GET /file", helpers.Wrap(getFile))
	index = "/file?filename=."
	return
}
