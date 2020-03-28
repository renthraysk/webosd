package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/renthraysk/webosd/eventsource"
)

type OSD struct {
	*eventsource.EventSource
	tmpl      *template.Template
	mu        sync.Mutex // Mutex lock for settings below.
	_settings *Settings
}

func New(es *eventsource.EventSource, tmpl *template.Template, settings *Settings) *OSD {
	return &OSD{
		EventSource: es,
		tmpl:        tmpl,
		_settings:   settings,
	}
}

func (o *OSD) copySettings() Settings {
	o.mu.Lock()
	defer o.mu.Unlock()
	return *o._settings
}

func (o *OSD) setSettings(s *Settings) {
	o.mu.Lock()
	defer o.mu.Unlock()
	*o._settings = *s
}

// SetMux sets up handlers for es, the EventSource, index and settings pages.
func (o *OSD) SetMux(mux *http.ServeMux, es, index, settings string) {
	mux.Handle(es, o)

	mux.HandleFunc(index, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := o.copySettings()
		data := struct {
			EventSource string
			*Settings
		}{
			EventSource: es,
			Settings:    &s,
		}
		w.Header().Set("Content-Type", "text/html")
		if err := o.tmpl.ExecuteTemplate(w, "index.gohtml", data); err != nil {
			logPrintf(r, "index template failed: %s", err)
		}
	}))

	mux.HandleFunc(settings, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s := o.copySettings()
			data := struct {
				*Settings
				Fonts []string
			}{
				Settings: &s,
				Fonts: []string{
					"monospace",
					"Bitstream Vera Sans Mono",
					"Consolas",
					"Courier",
					"Roboto Mono"},
			}
			w.Header().Set("Content-Type", "text/html")
			if err := o.tmpl.ExecuteTemplate(w, "settings.gohtml", data); err != nil {
				logPrintf(r, "settings template failed: %s", err)
			}

		case http.MethodPost:
			// Get copy of settings
			s := o.copySettings()
			// modify as needed
			_ = s.BackgroundColor.UnmarshalString(r.PostFormValue("backgroundColor"))
			_ = s.VoltColor.UnmarshalString(r.PostFormValue("voltColor"))
			_ = s.AmpColor.UnmarshalString(r.PostFormValue("ampColor"))
			s.Font = r.PostFormValue("font")
			if fontSize := r.PostFormValue("fontSize"); fontSize != "" {
				if u, err := strconv.ParseUint(fontSize, 10, 64); err == nil {
					s.FontSize = u
				}
			}
			if fontWeight := r.PostFormValue("fontWeight"); fontWeight != "" {
				if u, err := strconv.ParseUint(fontWeight, 10, 64); err == nil {
					s.FontWeight = u
				}
			}
			// Set
			o.setSettings(&s)
			o.Publish(eventsource.FormatEvent("settings", s.String()))
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
		}
	}))
}

func logPrintf(r *http.Request, format string, v ...interface{}) {
	s := r.Context().Value(http.ServerContextKey).(*http.Server)
	if s != nil && s.ErrorLog != nil {
		s.ErrorLog.Printf(format, v...)
		return
	}
	log.Printf(format, v...)
}
