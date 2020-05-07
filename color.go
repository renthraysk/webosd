package main

import (
	"fmt"
	"strconv"
)

type RGB struct {
	R, G, B uint8
}

func (c *RGB) hex4(u uint64) error {
	c.B = byte(u) & 0xF
	c.B |= c.B << 4
	u >>= 4
	c.G = byte(u) & 0xF
	c.G |= c.G << 4
	u >>= 4
	c.R = byte(u) & 0xF
	c.R |= c.R << 4
	return nil
}

func (c *RGB) hex8(u uint64) error {
	c.B = byte(u)
	u >>= 8
	c.G = byte(u)
	u >>= 8
	c.R = byte(u)
	return nil
}

func (c *RGB) UnmarshalString(s string) error {
	if len(s) < 2 {
		return fmt.Errorf("invalid color length: %q", s)
	}
	if s[0] == '#' {
		u, err := strconv.ParseUint(s[1:], 16, 32)
		if err != nil {
			return fmt.Errorf("invalid hex color %q: %w", s, err)
		}
		switch len(s) {
		case 4:
			return c.hex4(u)
		case 7:
			return c.hex8(u)
		}
		return fmt.Errorf("invalid hex length color %q", s)
	}
	return fmt.Errorf("invalid color %q", s)
}

func (c RGB) String() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

type RGBA struct {
	RGB
	A uint8
}

func (c *RGBA) UnmarshalString(s string) error {
	if len(s) < 2 {
		return fmt.Errorf("invalid color length: %q", s)
	}
	if s[0] == '#' {
		u, err := strconv.ParseUint(s[1:], 16, 32)
		if err != nil {
			return fmt.Errorf("invalid hex color %q: %w", s, err)
		}
		switch len(s) {
		case 4: // #RGB
			c.A = 255
			return c.RGB.hex4(u)
		case 5: // #RGBA
			c.A = byte(u) & 0x0F
			c.A |= c.A << 4
			return c.RGB.hex4(u >> 4)
		case 7: // #RRGGBB
			c.A = 255
			return c.RGB.hex8(u)
		case 9: // #RRGGBBAA
			c.A = byte(u)
			return c.RGB.hex8(u >> 8)
		}
		return fmt.Errorf("invalid hex length color %q", s)
	}
	return fmt.Errorf("invalid color %q", s)
}

func (c RGBA) String() string {
	return fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
}
