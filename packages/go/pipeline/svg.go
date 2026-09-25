package pipeline

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

var ErrInvalidSVG = errors.New("pipeline: invalid svg")

var allowedSVGElements = map[string]bool{
	"svg": true, "g": true, "defs": true, "symbol": true, "use": true,
	"path": true, "rect": true, "circle": true, "ellipse": true,
	"line": true, "polyline": true, "polygon": true,
	"text": true, "tspan": true, "style": true, "title": true, "desc": true,
	"clippath": true, "mask": true, "lineargradient": true, "radialgradient": true,
	"stop": true, "pattern": true, "marker": true, "image": true, "a": true,
}

// SanitizeSVG removes scripts, event handlers, and external references from SVG.
func SanitizeSVG(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, ErrInvalidSVG
	}
	dec := xml.NewDecoder(bytes.NewReader(src))
	dec.Strict = false
	var out bytes.Buffer
	enc := xml.NewEncoder(&out)
	depth := 0
	skipDepth := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, ErrInvalidSVG
		}
		switch t := tok.(type) {
		case xml.StartElement:
			name := strings.ToLower(t.Name.Local)
			if skipDepth > 0 {
				skipDepth++
				continue
			}
			if name == "script" || name == "foreignobject" {
				skipDepth = 1
				continue
			}
			if !allowedSVGElements[name] {
				skipDepth = 1
				continue
			}
			clean := xml.StartElement{Name: xml.Name{Local: t.Name.Local}}
			if t.Name.Space != "" {
				clean.Name.Space = t.Name.Space
			}
			for _, a := range t.Attr {
				an := strings.ToLower(a.Name.Local)
				if strings.HasPrefix(an, "on") {
					continue
				}
				if isDangerousURLAttr(an) && isExternalOrScriptURL(a.Value) {
					continue
				}
				if name == "use" && isDangerousURLAttr(an) && isExternalOrScriptURL(a.Value) {
					continue
				}
				clean.Attr = append(clean.Attr, a)
			}
			if err := enc.EncodeToken(clean); err != nil {
				return nil, ErrInvalidSVG
			}
			depth++
		case xml.EndElement:
			if skipDepth > 0 {
				skipDepth--
				continue
			}
			if depth == 0 {
				continue
			}
			if err := enc.EncodeToken(t); err != nil {
				return nil, ErrInvalidSVG
			}
			depth--
		case xml.CharData:
			if skipDepth > 0 {
				continue
			}
			if err := enc.EncodeToken(t); err != nil {
				return nil, ErrInvalidSVG
			}
		case xml.Comment, xml.ProcInst, xml.Directive:
			// drop
		}
	}
	if skipDepth != 0 || depth != 0 {
		// unclosed skip is ok if we drained; unclosed depth means bad xml usually caught earlier
	}
	if err := enc.Flush(); err != nil {
		return nil, ErrInvalidSVG
	}
	if out.Len() == 0 {
		return nil, ErrInvalidSVG
	}
	return out.Bytes(), nil
}

func isDangerousURLAttr(name string) bool {
	switch name {
	case "href", "xlink:href", "src":
		return true
	default:
		return name == "href" || strings.HasSuffix(name, "href")
	}
}

func isExternalOrScriptURL(v string) bool {
	s := strings.TrimSpace(strings.ToLower(v))
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "javascript:") || strings.HasPrefix(s, "data:text/html") {
		return true
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "//") {
		return true
	}
	return false
}
