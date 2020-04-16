GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOOPTS=-trimpath

BINARY=webosd
BUILD=$(shell git rev-parse HEAD)
VERSION=$(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
DIRTY=$(shell test -n "`git status --porcelain`" && echo "-dirty" || echo "")
LDFLAGS=-X main.Version=${VERSION} -X main.Build=${BUILD}$(DIRTY)

webosd: main.go osd.go $(wildcard eventsource/*.go) $(wildcard poller/*.go)

.PHONY: build
build: webosd
	$(GOBUILD) $(GOOPTS) -ldflags "$(LDFLAGS)" -o $<

.PHONY: build_static
build_static: webosd
	CGO_ENABLED=0 $(GOBUILD) $(GOOPTS) -ldflags="$(LDFLAGS)" -o $<

.dist/linux_amd64/webosd:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(GOOPTS) -ldflags "$(LDFLAGS)" -o $@

.dist/linux_arm64/webosd:
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(GOOPTS) -ldflags "$(LDFLAGS)" -o $@ 
	
.dist/windows_amd64/webosd.exe:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(GOOPTS) -ldflags "$(LDFLAGS)" -o $@

.dist/darwin_amd64/webosd:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(GOOPTS) -ldflags "$(LDFLAGS)" -o $@ 
	
webosd-$(VERSION).%.tar.gz: .dist/%/webosd tmpl/index.gohtml tmpl/settings.gohtml
	tar -zc --transform="s,^$(dir $<),webosd/,;s,^tmpl/,webosd/tmpl/," -f $@ $^

webosd-$(VERSION).windows_amd64.tar.gz: .dist/windows_amd64/webosd.exe tmpl/index.gohtml tmpl/settings.gohtml
	tar -zc --transform="s,^.dist/windows_amd64/,webosd/,;s,^tmpl/,webosd/tmpl/," -f $@ $^

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f webosd
	rm -rf .dist/
	rm -f webosd*.*.tar.gz

release: webosd-$(VERSION).linux_amd64.tar.gz \
		 webosd-$(VERSION).linux_arm64.tar.gz \
		 webosd-$(VERSION).windows_amd64.tar.gz \
		 webosd-$(VERSION).darwin_amd64.tar.gz
