# WebOSD

Proof of concept using a web server to broadcast device measurements* using an [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource).

\* This PoC uses a random number generator polled 10 times a second.

## Running

Start the web server
```$ ./webosd```

Navigate to [http://localhost:8080/osd/](http://localhost:8080/osd/)

### OBS Studio

Use the Browser plugin, with the URL set to [http://localhost:8080/osd/](http://localhost:8080/osd/)

## Settings

### Command line
```
Usage of ./webosd:
  -addr string
    	web server addr host:port (default "localhost:8080")
  -psu string
    	psu driver name (default "fake")
  -version
    	version
```

### Web

[http://localhost:8080/osd/settings](http://localhost:8080/osd/settings) provides a UI to change presentation settings live.
