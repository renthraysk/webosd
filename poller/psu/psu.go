package psu

import (
	"fmt"
	"io"
	"junk/webosd/eventsource"
)

type Event struct {
	volts float64
	amps  float64
}

func NewEvent(volts, amps float64) Event {
	return Event{volts: volts, amps: amps}
}

func (e Event) WriteTo(w io.Writer) (int64, error) {
	n, err := fmt.Fprintf(w, "event: volts\ndata: %.3f\n\nevent: amps\ndata: %.3f\n\n", e.volts, e.amps)
	return int64(n), err
}

var _ eventsource.Event = (*Event)(nil)
