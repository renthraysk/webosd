package fake

import (
	"math/rand"
	"time"

	"github.com/renthraysk/webosd/eventsource"
	"github.com/renthraysk/webosd/poller/psu"
)

type dev struct{}

func New() *dev { return &dev{} }

func (dev) Poll(t time.Time) eventsource.Event {
	return psu.NewEvent(11.75+rand.Float64()/3, 1.75+rand.Float64()/3)
}
