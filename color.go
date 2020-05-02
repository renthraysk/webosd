package main

import (
	"fmt"
	"strconv"
)

type RGB struct {
	R, G, B uint8
}

func (c *RGB) UnmarshalString(s string) error {
	if len(s) != 7 {
		return fmt.Errorf("invalid hex color length: %q", s)
	}
	if s[0] != '#' {
		return fmt.Errorf("invalid hex color %q", s)
	}
	u, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return fmt.Errorf("invalid hex color %q: %w", s, err)
	}
	c.B = byte(u)
	u >>= 8
	c.G = byte(u)
	u >>= 8
	c.R = byte(u)
	return nil
}

func (c RGB) String() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

type RGBA struct {
	RGB
	A uint8
}

func (c *RGBA) UnmarshalString(s string) error {
	if len(s) < 9 {
		if err := c.RGB.UnmarshalString(s); err != nil {
			return err
		}
		c.A = 255
		return nil
	}
	if s[0] != '#' {
		return fmt.Errorf("invalid hex color %q", s)
	}
	u, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return fmt.Errorf("invalid hex color %q: %w", s, err)
	}
	c.R = byte(u >> 24)
	c.G = byte(u >> 16)
	c.B = byte(u >> 8)
	c.A = byte(u)
	return nil
}

func (c RGBA) String() string {
	return fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
}
