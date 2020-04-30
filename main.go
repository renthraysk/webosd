package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/renthraysk/webosd/device"
	"github.com/renthraysk/webosd/eventsource"
)

var Version string = "xx.xx.xx"
var Build string = "xxxx"

var fonts = []string{
	"monospace",
	"Consolas",
	"Courier",
	"Roboto Mono",
}

type Settings struct {
	BackgroundColor RGBA
	Padding         uint64
	VoltColor       RGB
	AmpColor        RGB
	FontFamily      string
	FontSize        uint64
	FontWeight      uint64
	LineHeight      uint64
}

func (s *Settings) String() string {
	v := url.Values{
		"backgroundColor": []string{s.BackgroundColor.String()},
		"padding":         []string{strconv.FormatUint(s.Padding, 10) + "px"},
		"voltColor":       []string{s.VoltColor.String()},
		"ampColor":        []string{s.AmpColor.String()},
		"fontFamily":      []string{s.FontFamily},
		"fontSize":        []string{strconv.FormatUint(s.FontSize, 10) + "px"},
		"fontWeight":      []string{strconv.FormatUint(s.FontWeight, 10)},
		"lineHeight":      []string{strconv.FormatUint(s.LineHeight, 10) + "%"},
	}
	return v.Encode()
}

func main() {

	settings := &Settings{
		BackgroundColor: RGBA{RGB: RGB{R: 0, G: 0, B: 0}, A: 255},
		VoltColor:       RGB{G: 0x80},
		AmpColor:        RGB{R: 0xFF, G: 0xFF},
		FontFamily:      "monospace",
		FontSize:        70,
		FontWeight:      400,
		LineHeight:      110,
	}

	log := log.New(os.Stderr, "", log.LstdFlags)
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	addr := flagset.String("addr", "localhost:8080", "web server addr host:port")
	backgroundColor := flagset.String("backgroundColor", settings.BackgroundColor.String(), "RGBA background color")
	flagset.Uint64Var(&settings.Padding, "padding", settings.Padding, "padding")
	voltColor := flagset.String("voltColor", settings.VoltColor.String(), "RGB volt color")
	ampColor := flagset.String("ampColor", settings.AmpColor.String(), "RGB amp color")
	psu := flagset.String("psu", "fake", "psu driver name")
	flagset.StringVar(&settings.FontFamily, "font", settings.FontFamily, "font name")
	flagset.Uint64Var(&settings.FontSize, "fontsize", settings.FontSize, "font size")
	flagset.Uint64Var(&settings.FontWeight, "fontweight", settings.FontWeight, "font weight")
	flagset.Uint64Var(&settings.LineHeight, "lineheight", settings.LineHeight, "line height")
	version := flagset.Bool("version", false, "version")

	flagset.Parse(os.Args[1:])

	if *version {
		fmt.Fprintf(os.Stdout, "Version: %s Build: %s\n", Version, Build)
		os.Exit(0)
	}

	if _, _, err := net.SplitHostPort(*addr); err != nil {
		log.Fatalf("invalid addr %q: %s", *addr, err)
	}

	if settings.FontWeight < 100 {
		settings.FontWeight = 100
	} else if settings.FontWeight > 900 {
		settings.FontWeight = 900
	}

	if err := settings.BackgroundColor.UnmarshalString(*backgroundColor); err != nil {
		log.Fatalf("invalid background color: %s", err)
	}
	if err := settings.VoltColor.UnmarshalString(*voltColor); err != nil {
		log.Fatalf("invalid volt color: %s", err)
	}
	if err := settings.AmpColor.UnmarshalString(*ampColor); err != nil {
		log.Fatalf("invalid amp color: %s", err)
	}

	osdTmpl := template.Must(template.New("main").ParseFiles("./tmpl/osd/index.gohtml", "./tmpl/osd/settings.gohtml"))

	p, err := device.New(*psu, "")
	if err != nil {
		log.Printf("failed to create device %q: %s", *psu, err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	osd := New(eventsource.New(ctx), "/es", osdTmpl, settings)

	// 10 times a second
	go eventsource.Ticker(ctx, osd, p.Poll, time.Second/10)

	// Web server
	mux := http.NewServeMux()

	mux.Handle("/", index())

	osd.SetMux(mux, "/osd", "/osd/settings")

	s := http.Server{
		Addr:        *addr,
		Handler:     mux,
		BaseContext: func(net.Listener) context.Context { return ctx },
		ErrorLog:    log,
	}
	errCh := make(chan error, 1)
	go func() { errCh <- s.ListenAndServe() }()

	url := url.URL{
		Scheme: "http",
		Host:   *addr,
		Path:   "/",
	}

	fmt.Fprintf(os.Stdout, "Index %s\n", url.String())
	url.Path = "/osd"
	fmt.Fprintf(os.Stdout, "OSD %s\n", url.String())
	url.Path = "/osd/settings"
	fmt.Fprintf(os.Stdout, "OSD Settings %s\n", url.String())

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt) // cntrl+c to quit
	select {
	case err = <-errCh:
	case s := <-sigCh:
		err = errors.New(s.String())
	}
	// Shutdown
	log.Printf("shutting down: %s", err)
	signal.Reset(os.Interrupt)
	cancel()
	close(sigCh)
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err2 := s.Shutdown(ctx); err2 != nil {
			log.Printf("failed to shutdown webserver: %s", err2)
		}
	}
	close(errCh)
	if err != nil {
		os.Exit(1)
	}
}

func index() http.Handler {
	tmpl := template.Must(template.ParseFiles("./tmpl/index.gohtml"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
			return
		}

		tmpl.ExecuteTemplate(w, "index.gohtml", nil)
	})
}
