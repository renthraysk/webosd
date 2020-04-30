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
	path      string
	tmpl      *template.Template
	mu        sync.Mutex // Mutex lock for settings below.
	_settings *Settings
}

func New(es *eventsource.EventSource, path string, tmpl *template.Template, settings *Settings) *OSD {
	return &OSD{
		EventSource: es,
		path:        path,
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

func (o *OSD) settings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data := struct {
			Settings
			EventSource string
			Fonts       []string
		}{
			Settings:    o.copySettings(),
			EventSource: o.path,
			Fonts:       fonts,
		}
		w.Header().Set("Content-Type", "text/html")
		if err := o.tmpl.ExecuteTemplate(w, "settings.gohtml", data); err != nil {
			logPrintf(r, "settings template failed: %s", err)
		}

	case http.MethodPost:
		// Get copy of settings
		s := o.copySettings()
		// modify as needed

		if backgroundColor := r.PostFormValue("backgroundColor"); backgroundColor != "" {
			if err := s.BackgroundColor.UnmarshalString(backgroundColor); err != nil {
				log.Printf("failed to parse backgroundColor %q: %v", backgroundColor, err)
			}
		}

		if alpha := r.PostFormValue("backgroundAlpha"); alpha != "" {
			if a, err := strconv.ParseUint(alpha, 10, 32); err == nil {
				if a >= 0xFF {
					s.BackgroundColor.A = 0xFF
				} else if a <= 0 {
					s.BackgroundColor.A = 0
				} else {
					s.BackgroundColor.A = byte(a)
				}
			}
		}
		if voltColor := r.PostFormValue("voltColor"); voltColor != "" {
			if err := s.VoltColor.UnmarshalString(voltColor); err != nil {
				log.Printf("failed to parse voltColor %q: %v", voltColor, err)
			}
		}
		if ampColor := r.PostFormValue("ampColor"); ampColor != "" {
			if err := s.AmpColor.UnmarshalString(r.PostFormValue("ampColor")); err != nil {
				log.Printf("failed to parse ampColor %q: %v", ampColor, err)
			}
		}
		// Validate font, string parameter so have to prevent
		if font := r.PostFormValue("font"); font != "" {
			for _, name := range fonts {
				if name == font {
					s.Font = name
				}
			}
		}
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

		if lineHeight := r.PostFormValue("lineHeight"); lineHeight != "" {
			if u, err := strconv.ParseUint(lineHeight, 10, 64); err == nil {
				s.LineHeight = u
			}
		}

		// Set
		o.setSettings(&s)
		o.Publish(eventsource.NewEvent("settings", s.String()))
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}

// SetMux sets up handlers for es, the EventSource, index and settings pages.
func (o *OSD) SetMux(mux *http.ServeMux, index, settings string) {
	mux.Handle(o.path, o)
	mux.HandleFunc(index, func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			EventSource string
			Settings
		}{
			EventSource: o.path,
			Settings:    o.copySettings(),
		}
		w.Header().Set("Content-Type", "text/html")
		if err := o.tmpl.ExecuteTemplate(w, "index.gohtml", data); err != nil {
			logPrintf(r, "index template failed: %s", err)
		}
	})
	mux.HandleFunc(settings, o.settings)
}

func logPrintf(r *http.Request, format string, v ...interface{}) {
	s := r.Context().Value(http.ServerContextKey).(*http.Server)
	if s != nil && s.ErrorLog != nil {
		s.ErrorLog.Printf(format, v...)
		return
	}
	log.Printf(format, v...)
}
