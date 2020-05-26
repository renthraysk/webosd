package eventsource

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

type Event interface {
	io.WriterTo
}

type Publisher interface {
	Publish(e Event) bool
}

type eventBytes []byte

func NewEvent(event, data string) Event {
	var b bytes.Buffer

	b.WriteString("event: ")
	b.WriteString(event)
	b.WriteString("\ndata: ")
	for i := strings.Index(data, "\n"); i >= 0; i = strings.Index(data, "\n") {
		i++
		b.WriteString(data[:i])
		b.WriteString("data: ")
		data = data[i:]
	}
	b.WriteString(data)
	b.WriteString("\n\n")
	return eventBytes(b.Bytes())
}

func (e eventBytes) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte(e))
	return int64(n), err
}

func NewEventFromError(err error) Event {
	return NewEvent("error", err.Error())
}

type Subscriber chan<- Event

type EventSource struct {
	in            chan Event
	subscribeCh   chan Subscriber
	unsubscribeCh chan Subscriber
}

func New(ctx context.Context) *EventSource {
	es := &EventSource{
		in:            make(chan Event, 8),
		subscribeCh:   make(chan Subscriber, 2),
		unsubscribeCh: make(chan Subscriber, 2),
	}
	go es.run(context.WithCancel(ctx))
	return es
}

func (es *EventSource) run(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	subscribers := make(map[Subscriber]struct{})
	for {
		select {
		case e, ok := <-es.in:
			if !ok {
				return
			}
			for s := range subscribers {
				select {
				case s <- e:
				default:
					delete(subscribers, s)
					close(s)
				}
			}

		case s := <-es.subscribeCh:
			subscribers[s] = struct{}{}

		case s := <-es.unsubscribeCh:
			delete(subscribers, s)
			close(s)

		case <-ctx.Done():
			return
		}
	}
}

// Publish event to all subscribers
func (es *EventSource) Publish(e Event) bool {
	select {
	case es.in <- e:
		return true
	default:
		return false
	}
}

func (es *EventSource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming events not supported", http.StatusBadRequest)
		return
	}

	subscriberCh := make(chan Event, 3)
	es.subscribeCh <- subscriberCh
	defer func() {
		select {
		case es.unsubscribeCh <- subscriberCh:
		default:
		}
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
	for {
		f.Flush()
		select {
		case e, ok := <-subscriberCh:
			if !ok {
				return
			}
			// @TODO Error handling...
			e.WriteTo(w)

		case <-r.Context().Done():
			return
		}
	}
}

// Ticker calls f every d and publishes response
func (es *EventSource) Ticker(ctx context.Context, f func(t time.Time) Event, d time.Duration) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			// @TODO Error handling...
			es.Publish(f(t))

		case <-ctx.Done():
			return
		}
	}
}
