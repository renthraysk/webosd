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

	"github.com/renthraysk/webosd/eventsource"
	"github.com/renthraysk/webosd/poller/psu/fake"
)

var Version string = "xx.xx.xx"
var Build string = "xxxx"

type Color uint32

func (c *Color) UnmarshalString(s string) error {
	if len(s) != 7 || s[0] != '#' {
		return fmt.Errorf("invalid color length %q", s)
	}
	u, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return err
	}
	*c = Color(u)
	return nil
}

func (c Color) String() string {
	return fmt.Sprintf("#%06x", uint32(c))
}

var fonts = []string{
	"monospace",
	"Consolas",
	"Courier",
	"Roboto Mono",
}

type Settings struct {
	BackgroundColor Color
	VoltColor       Color
	AmpColor        Color
	Font            string
	FontSize        uint64
	FontWeight      uint64
	LineHeight      uint64
}

func (s *Settings) String() string {
	v := url.Values{
		"backgroundColor": []string{s.BackgroundColor.String()},
		"voltColor":       []string{s.VoltColor.String()},
		"ampColor":        []string{s.AmpColor.String()},
		"font":            []string{s.Font},
		"fontSize":        []string{strconv.FormatUint(s.FontSize, 10)},
		"fontWeight":      []string{strconv.FormatUint(s.FontWeight, 10)},
		"lineHeight":      []string{strconv.FormatUint(s.LineHeight, 10)},
	}
	return v.Encode()
}

func main() {

	settings := &Settings{
		BackgroundColor: 0x000000,
		VoltColor:       0x008000,
		AmpColor:        0xFFFF00,
		Font:            "monospace",
		FontSize:        70,
		FontWeight:      400,
		LineHeight:      110,
	}

	log := log.New(os.Stderr, "", log.LstdFlags)
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	addr := flagset.String("addr", "localhost:8080", "web server addr host:port")
	backgroundColor := flagset.String("backgroundColor", settings.BackgroundColor.String(), "background color")
	voltColor := flagset.String("voltColor", settings.VoltColor.String(), "volt color")
	ampColor := flagset.String("ampColor", settings.AmpColor.String(), "amp color")
	flagset.StringVar(&settings.Font, "font", settings.Font, "font name")
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

	tmpl, err := template.New("main").ParseFiles("./tmpl/index.gohtml", "./tmpl/settings.gohtml")
	if err != nil {
		log.Fatalf("failed to load templates: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	osd := New(eventsource.New(ctx), tmpl, settings)

	// go routine to sample fake PSU 10 times a second.
	go eventsource.Ticker(ctx, osd, fake.New(), time.Second/10)

	// Web server
	mux := http.NewServeMux()
	osd.SetMux(mux, "/es", "/", "/settings")

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

	fmt.Fprintf(os.Stdout, "OSD %s\n", url.String())
	url.Path = "/settings"
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
