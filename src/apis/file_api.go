package apis

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"ponito/lilser/apis/helpers"
	"ponito/lilser/apis/helpers/urling"
	"sort"
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
		_ = dirListTmpl.Execute(w, struct {
			Path    string
			Entries []dirEntryView
		}{
			Path:    filename,
			Entries: views,
		})
	} else {
		http.ServeFile(w, r, filename)
	}
	return
}

func UseFile() {
	go printWorkingDir()

	http.HandleFunc("PUT /file", helpers.Wrap(putFile))
	http.HandleFunc("GET /file", helpers.Wrap(getFile))
}

var dirListTmpl = template.Must(template.New("dirlist").Parse(`<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Index of {{.Path}}</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, Helvetica, Arial, sans-serif; padding: 16px; }
    ul { list-style: none; padding-left: 0; }
    li { margin: 4px 0; }
    a { text-decoration: none; }
  </style>
</head>
<body>
  <h3>Index of {{.Path}}</h3>
  <ul>
    {{range .Entries}}
      <li><a href="?filename={{$.Path}}/{{.Name}}">{{.Name}}{{if .IsDir}}/{{end}}</a></li>
    {{end}}
  </ul>
  <div>
    <input id="file" type="file" name="file" required />
  </div>
  <script>
    (function(){
      var input = document.getElementById('file');
      input.addEventListener('change', function(){
        var f = input.files[0];
        if(!f) return;
        var p = '{{.Path}}';
        var url = '?filename=' + encodeURIComponent(p) + '/' + encodeURIComponent(f.name);
        fetch(url, { method: 'PUT', body: f })
          .then(function(r){ if(!r.ok) throw new Error(r.statusText); })
          .then(function(){ window.location.reload(); })
          .catch(function(e){ console.log(e); alert('Upload failed: ' + e); });
      });
    })();
  </script>
</body>
</html>`))
