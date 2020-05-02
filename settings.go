package main

import (
	"bytes"
	"io"
	"net/url"
	"strconv"
)

type Settings struct {
	backgroundColor RGBA
	padding         uint64
	borderRadius    uint64
	voltColor       RGB
	ampColor        RGB
	fontFamily      string
	fontSize        uint64
	fontWeight      uint64
	lineHeight      uint64
	textStrokeWidth uint64
	textStrokeColor RGB
}

func (s *Settings) WriteTo(w io.Writer) (int64, error) {
	var b bytes.Buffer

	b.WriteString(":root {\n")
	b.WriteString("--background-color: " + s.backgroundColor.String() + ";\n")
	b.WriteString("--font-family: " + s.fontFamily + ";\n")
	b.WriteString("--font-size: " + strconv.FormatUint(s.fontSize, 10) + "px;\n")
	b.WriteString("--font-weight: " + strconv.FormatUint(s.fontWeight, 10) + ";\n")
	b.WriteString("--line-height: " + strconv.FormatUint(s.lineHeight, 10) + "%;\n")
	b.WriteString("--volt-color: " + s.voltColor.String() + ";\n")
	b.WriteString("--amp-color: " + s.ampColor.String() + ";\n")
	b.WriteString("--padding: " + strconv.FormatUint(s.padding, 10) + "px;\n")
	b.WriteString("--border-radius: " + strconv.FormatUint(s.borderRadius, 10) + "px;\n")
	b.WriteString("--box-shadow: " + "10px 10px 20px #000000A0" + ";\n")
	b.WriteString("--text-stroke-width: " + strconv.FormatUint(s.textStrokeWidth, 10) + "px;\n")
	b.WriteString("--text-stroke-color: " + s.textStrokeColor.String() + ";\n")
	b.WriteString("}\n")
	return b.WriteTo(w)
}

func (s *Settings) Set(v url.Values) {
	if backgroundColor := v.Get("backgroundColor"); backgroundColor != "" {
		s.backgroundColor.UnmarshalString(backgroundColor)
	}
	if alpha := v.Get("backgroundAlpha"); alpha != "" {
		if a, err := strconv.ParseUint(alpha, 10, 32); err == nil {
			if a >= 0xFF {
				s.backgroundColor.A = 0xFF
			} else if a <= 0 {
				s.backgroundColor.A = 0
			} else {
				s.backgroundColor.A = byte(a)
			}
		}
	}

	if padding := v.Get("padding"); padding != "" {
		if p, err := strconv.ParseUint(padding, 10, 64); err == nil {
			s.padding = p
		}
	}
	if borderRadius := v.Get("borderRadius"); borderRadius != "" {
		if b, err := strconv.ParseUint(borderRadius, 10, 64); err == nil {
			s.borderRadius = b
		}
	}

	if voltColor := v.Get("voltColor"); voltColor != "" {
		s.voltColor.UnmarshalString(voltColor)
	}
	if ampColor := v.Get("ampColor"); ampColor != "" {
		s.ampColor.UnmarshalString(ampColor)
	}
	// Validate font, string parameter so have to prevent
	if fontFamily := v.Get("font"); fontFamily != "" {
		s.fontFamily = fontFamily
	}
	if fontSize := v.Get("fontSize"); fontSize != "" {
		if u, err := strconv.ParseUint(fontSize, 10, 64); err == nil {
			s.fontSize = u
		}
	}
	if fontWeight := v.Get("fontWeight"); fontWeight != "" {
		if u, err := strconv.ParseUint(fontWeight, 10, 64); err == nil {
			s.fontWeight = u
		}
	}
	if lineHeight := v.Get("lineHeight"); lineHeight != "" {
		if u, err := strconv.ParseUint(lineHeight, 10, 64); err == nil {
			s.lineHeight = u
		}
	}

	if textStrokeWidth := v.Get("textStrokeWidth"); textStrokeWidth != "" {
		if u, err := strconv.ParseUint(textStrokeWidth, 10, 64); err == nil {
			s.textStrokeWidth = u
		}
	}
	if textStrokeColor := v.Get("textStrokeColor"); textStrokeColor != "" {
		s.textStrokeColor.UnmarshalString(textStrokeColor)
	}
}
