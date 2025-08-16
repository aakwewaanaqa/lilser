package apis

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"ponito/lilser/apis/helpers"
	"ponito/lilser/apis/helpers/urling"
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
		dirInfo  os.FileInfo
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
		if dirInfo, err = os.Stat(folder); err != nil {
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
		if tmpFile, err = os.CreateTemp(dirInfo.Name(), "tmp"); err != nil {
			state = 500
			return
		}
		defer tmpFile.Close()

		if _, err = io.Copy(tmpFile, r.Body); err != nil {
			state = 500
			return
		}

		err = os.Rename(tmpFile.Name(), filename)
		log.Println(tmpFile.Name(), filename, err)
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

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		for _, entry := range entries {
			var msg string
			msg = fmt.Sprintf("<a href='?filename=%s/%s'>%s</a><br>", filename, entry.Name(), entry.Name())
			_, _ = w.Write([]byte(msg))
		}

		_, _ = w.Write([]byte(fmt.Sprintf(`
<input id='file' type="file" name="file" required/>
<script>
    function putFile() {
        const file = document.querySelector('input[type="file"]').files[0];
        fetch('?filename=%s/' + file.name, {
            method: 'PUT',
            body: file,
        })
        .then(rsp => rsp.json())
        .then(_ => window.location.reload())
        .catch(err => console.log(err))
    }

    document.getElementById('file').onchange = e => {
        putFile();
    }
</script>
`, filename)))
	} else {
		var (
			file *os.File
		)

		if file, err = os.Open(filename); err != nil {
			state = 500
			return
		}

		defer file.Close()

		if _, err = io.Copy(w, file); err != nil {
			state = 500
			return
		}
	}
	return
}

func UseFile() {
	go printWorkingDir()

	http.HandleFunc("PUT /file", helpers.Wrap(putFile))
	http.HandleFunc("GET /file", helpers.Wrap(getFile))
}
