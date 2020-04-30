# WebOSD

Proof of concept using a web server to broadcast device measurements* using an [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource).

\* This PoC uses a random number generator polled 10 times a second.

## Running

Start the web server
```$ ./webosd```

Navigate to [http://localhost:8080/](http://localhost:8080/osd)

### OBS Studio

Use the Browser plugin, with the URL set to [http://localhost:8080/](http://localhost:8080/osd)

## Settings

### Command line
```
Usage of ./webosd:
  -addr string
    	web server addr host:port (default "localhost:8080")
  -ampColor string
    	RGB amp color (default "#ffff00")
  -backgroundColor string
    	RGBA background color (default "#000000ff")
  -font string
    	font name (default "monospace")
  -fontsize uint
    	font size (default 70)
  -fontweight uint
    	font weight (default 400)
  -lineheight uint
    	line height (default 110)
  -padding uint
    	padding
  -psu string
    	psu driver name (default "fake")
  -version
    	version
  -voltColor string
    	RGB volt color (default "#008000")

```

### Web

[http://localhost:8080/osd/settings](http://localhost:8080/osd/settings) provides a UI to change presentation settings live.
