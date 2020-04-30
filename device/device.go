package device

import (
	"fmt"

	"github.com/renthraysk/webosd/device/psu"
)

func New(name, addr string) (psu.PSU, error) {
	switch name {
	case "fake":
		return psu.Fake(), nil
	}
	return nil, fmt.Errorf("unknown %q", name)
}
