package psu

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"time"

	"github.com/renthraysk/webosd/eventsource"
)

type Event struct {
	time      time.Time
	dcVoltage float64
	dcCurrent float64
}

func NewEvent(time time.Time, dcVoltage, dcCurrent float64) eventsource.Event {
	return Event{time: time, dcVoltage: dcVoltage, dcCurrent: dcCurrent}
}

func (e Event) WriteTo(w io.Writer) (int64, error) {
	n, err := fmt.Fprintf(w, "event: psu\ndata: {\"time\": %d, \"voltage\": %f, \"current\": %f}\n\n", e.time.UnixNano()/int64(time.Millisecond), e.dcVoltage, e.dcCurrent)
	return int64(n), err
}

var _ eventsource.Event = (*Event)(nil)

type PSU interface {
	Poll(ctx context.Context, t time.Time) (eventsource.Event, error)
}

type fake struct {
}

func Fake() PSU { return &fake{} }

func (fake) Poll(ctx context.Context, t time.Time) (eventsource.Event, error) {
	return NewEvent(t, 11.75+rand.Float64(), 1.75+rand.Float64()/2), nil
}

type sin struct {
	x float64
}

func Sin() PSU { return &sin{} }

func (s *sin) Poll(ctx context.Context, t time.Time) (eventsource.Event, error) {
	s.x += 1.0 / 20
	return NewEvent(t, 11+math.Sin(s.x)*4, 2+math.Cos(s.x)), nil
}

/*

type SCPI struct {
	conn conn.Conn
}

func New(conn conn.Conn) *SCPI {
	return &SCPI{conn: conn}
}

func (s *SCPI) Poll(ctx context.Context, t time.Time) (eventsource.Event, error) {
	var v, c scpi.Float64
	if err := s.conn.Query(ctx, "MEAS:VOLT?;CURR?", &v, &c); err != nil {
		return nil, err
	}
	return NewEvent(t, v.Float64(), c.Float64()), nil
}

var _ PSU = (*SCPI)(nil)
*/
