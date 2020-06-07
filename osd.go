package main

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"

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

type HTMLDir struct {
	d http.Dir
}

func (d *HTMLDir) Open(name string) (http.File, error) {
	if ext := filepath.Ext(name); ext == "" {
		if f, err := d.d.Open(name + ".html"); err == nil {
			return f, nil
		}
	}
	return d.d.Open(name)
}

// SetMux sets up handlers for es, the EventSource, index and settings pages.
func (o *OSD) SetMux(mux *http.ServeMux) {
	mux.Handle("/es", o)
	mux.Handle("/osd/", http.StripPrefix("/osd/", http.FileServer(&HTMLDir{http.Dir("./static/osd/")})))

	mux.HandleFunc("/footer", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:

			s := struct {
				Command  string `json:"command"`
				Text     string `json:"text"`
				Duration uint32 `json:"duration"`
			}{
				Command: r.PostFormValue("command"),
				Text:    r.PostFormValue("text"),
			}

			if d := r.PostFormValue("duration"); d != "" {
				if u, err := strconv.ParseUint(d, 10, 32); err == nil {
					s.Duration = uint32(u)
				}
			}

			if b, err := json.Marshal(&s); err == nil {
				o.Publish(eventsource.NewEventBytes("footer", b))

				if rf := r.Referer(); rf != "" {
					http.Redirect(w, r, rf, http.StatusSeeOther)
				}
			}
		}
	})
}
