package psu

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"time"

	"github.com/renthraysk/webosd/eventsource"
)

type Event struct {
	time  time.Time
	volts float64
	amps  float64
}

func NewEvent(time time.Time, volts, amps float64) eventsource.Event {
	return Event{time: time, volts: volts, amps: amps}
}

func (e Event) WriteTo(w io.Writer) (int64, error) {
	n, err := fmt.Fprintf(w, "event: psu\ndata: {\"time\": %d, \"volts\": %f, \"amps\": %f}\n\n", e.time.UnixNano()/int64(time.Millisecond), e.volts, e.amps)
	return int64(n), err
}

var _ eventsource.Event = (*Event)(nil)

type PSU interface {
	Poll(t time.Time) eventsource.Event
}

type fake struct {
}

func Fake() PSU { return &fake{} }

func (fake) Poll(t time.Time) eventsource.Event {
	return NewEvent(t, 11.75+rand.Float64(), 1.75+rand.Float64()/2)
}

type sin struct {
	x float64
}

func Sin() PSU { return &sin{} }

func (s *sin) Poll(t time.Time) eventsource.Event {
	s.x += 1.0 / 20
	return NewEvent(t, 11+math.Sin(s.x)*4, 2+math.Cos(s.x))
}
