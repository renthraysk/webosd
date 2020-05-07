package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/renthraysk/webosd/eventsource"
)

type OSD struct {
	*eventsource.EventSource
}

func New(es *eventsource.EventSource) *OSD {
	return &OSD{
		EventSource: es,
	}
}

// SetMux sets up handlers for es, the EventSource, index and settings pages.
func (o *OSD) SetMux(mux *http.ServeMux) {
	mux.Handle("/es", o)
	mux.HandleFunc("/osd/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./static/osd/index.html") })
	mux.HandleFunc("/osd/graph", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./static/osd/graph.html") })

	mux.Handle("/osd/js/", http.StripPrefix("/osd/js/", http.FileServer(http.Dir("./static/osd/js/"))))

	mux.HandleFunc("/osd/css/root.css",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/css")
			w.Header().Set("Cache-Control", "no-cache")
			http.ServeFile(w, r, "./static/osd/css/root.css")
		})

	mux.HandleFunc("/osd/settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			http.ServeFile(w, r, "static/osd/settings.html")

		case http.MethodPost:
			var s Settings

			r.ParseForm()
			s.Set(r.PostForm)

			if err := writeFile("./static/osd/css", "root.css", &s, 0666); err == nil {
				o.Publish(eventsource.NewEvent("reload", ""))
			}
			http.Redirect(w, r, "/osd/settings", http.StatusSeeOther)
		}
	})
}

func logPrintf(r *http.Request, format string, v ...interface{}) {
	s := r.Context().Value(http.ServerContextKey).(*http.Server)
	if s != nil && s.ErrorLog != nil {
		s.ErrorLog.Printf(format, v...)
		return
	}
	log.Printf(format, v...)
}

func writeFile(path, name string, wt io.WriterTo, perm os.FileMode) error {
	f, err := ioutil.TempFile(path, name)
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	if _, err := wt.WriteTo(f); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), filepath.Join(path, name))
}
