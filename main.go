package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/renthraysk/webosd/device"
	"github.com/renthraysk/webosd/eventsource"
)

var Version string = "xx.xx.xx"
var Build string = "xxxx"

func main() {
	log := log.New(os.Stderr, "", log.LstdFlags)
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	addr := flagset.String("addr", "localhost:8080", "web server addr host:port")
	psu := flagset.String("psu", "fake", "psu driver name")
	version := flagset.Bool("version", false, "version")

	flagset.Parse(os.Args[1:])

	if *version {
		fmt.Fprintf(os.Stdout, "Version: %s Build: %s\n", Version, Build)
		os.Exit(0)
	}

	if _, _, err := net.SplitHostPort(*addr); err != nil {
		log.Fatalf("invalid addr %q: %s", *addr, err)
	}

	p, err := device.New(*psu, "")
	if err != nil {
		log.Printf("failed to create device %q: %s", *psu, err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	osd := New(eventsource.New(ctx))

	// 10 times a second
	go osd.Ticker(ctx, p.Poll, time.Second/10)

	// Web server
	mux := http.NewServeMux()
	osd.SetMux(mux)

	s := http.Server{
		Addr:        *addr,
		Handler:     mux,
		BaseContext: func(net.Listener) context.Context { return ctx },
		ErrorLog:    log,
	}
	errCh := make(chan error, 1)
	go func() { errCh <- s.ListenAndServe() }()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt, os.Kill) // cntrl+c to quit

	for err == nil {
		select {
		case err = <-errCh:
		case s := <-sigCh:
			switch s {
			case os.Interrupt, os.Kill:
				err = errors.New(s.String())
			}
		}
	}
	// Shutdown
	log.Printf("shutting down: %s", err)
	signal.Reset(os.Interrupt, os.Kill)
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			http.Redirect(w, r, "/index.html", http.StatusSeeOther)
			return
		}
		http.ServeFile(w, r, "./static/index.html")
	})
}
