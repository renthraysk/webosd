package psu

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/rand"
	"time"

	"github.com/renthraysk/webosd/conn"
	"github.com/renthraysk/webosd/eventsource"
)

type Event struct {
	volts float64
	amps  float64
}

func NewEvent(volts, amps float64) eventsource.Event {
	return Event{volts: volts, amps: amps}
}

func (e Event) WriteTo(w io.Writer) (int64, error) {
	n, err := fmt.Fprintf(w, "event: volts\ndata: %.3f\n\nevent: amps\ndata: %.3f\n\n", e.volts, e.amps)
	return int64(n), err
}

var _ eventsource.Event = (*Event)(nil)

type Commands interface {
	QueryVoltage(b *bytes.Buffer) error
	QueryCurrent(b *bytes.Buffer) error
}

type PSU interface {
	Poll(t time.Time) eventsource.Event
}

type psu struct {
	conn     conn.Conn
	commands Commands
}

func New(conn conn.Conn, commands Commands) PSU {
	return &psu{conn: conn, commands: commands}
}

func (p *psu) Poll(t time.Time) eventsource.Event {
	if _, err := p.conn.WriteCommand(p.commands.QueryVoltage); err != nil {
		return eventsource.NewEventFromError(err)
	}
	if _, err := p.conn.WriteCommand(p.commands.QueryCurrent); err != nil {
		return eventsource.NewEventFromError(err)
	}
	return NewEvent(1.0, 1.0)
}

type fake struct {
}

func Fake() PSU { return &fake{} }

func (fake) Poll(t time.Time) eventsource.Event {
	return NewEvent(11.75+rand.Float64(), 1.75+rand.Float64()/2)
}

type sin struct {
	x float64
}

func Sin() PSU { return &sin{} }

func (s *sin) Poll(t time.Time) eventsource.Event {
	s.x += 1.0 / 20
	return NewEvent(11+math.Sin(s.x)*4, 2+math.Cos(s.x))
}
