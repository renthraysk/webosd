package main

import (
	"net/http"
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
}
