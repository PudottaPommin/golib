package main

import (
	"encoding/xml"
	"io"
)

type SVG struct {
	XMLName    xml.Name    `xml:"svg"`
	Version    string      `xml:"version,attr"`
	Xmlns      string      `xml:"xmlns,attr"`
	XmlnsXlink string      `xml:"xmlns:xlink,attr"`
	XmlnsXml   string      `xml:"xmlns:xml,attr"`
	Symbols    []SVGSymbol `xml:"symbol"`
}

type SVGSymbol struct {
	XMLName xml.Name `xml:"symbol"`
	ViewBox string   `xml:"viewBox,attr"`
	ID      string   `xml:"id,attr"`
	Content string   `xml:",innerxml"`
}

func NewSVG() SVG {
	return SVG{
		Version:    "1.1",
		Xmlns:      "http://www.w3.org/2000/svg",
		XmlnsXlink: "http://www.w3.org/1999/xlink",
		XmlnsXml:   "http://www.w3.org/XML/1998/namespace",
		Symbols:    make([]SVGSymbol, 0),
	}
}

func StreamFilterSVG(r io.Reader, w io.Writer, filter func(SVGSymbol) bool) error {
	dec := xml.NewDecoder(r)
	cw := &countingWriter{w: w}
	if _, err := cw.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
`)); err != nil {
		return err
	}

	enc := xml.NewEncoder(cw)
	enc.Indent("", "  ")
	defer enc.Close()

	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "svg" {
				e := se.Copy()
				e.Attr = make([]xml.Attr, 0)
				for _, a := range se.Attr {
					if a.Name.Local == "xmlns" || a.Name.Space == "xmlns" {
						// we skip this because it would be double encoded
					} else {
						e.Attr = append(e.Attr, a)
					}
				}

				if err = enc.EncodeToken(e); err != nil {
					return err
				}

				continue
			}
			if se.Name.Local == "symbol" {
				var symbol SVGSymbol
				if err = dec.DecodeElement(&symbol, &se); err != nil {
					return err
				}
				if filter(symbol) {
					// Use Encode(symbol) to avoid duplicating attributes from 'se'
					if err = enc.Encode(symbol); err != nil {
						return err
					}
				}
			}

		case xml.EndElement:
			if se.Name.Local == "svg" {
				if err = enc.EncodeToken(se); err != nil {
					return err
				}
			}
		}
	}

	return enc.Flush()
}

type countingWriter struct {
	w     io.Writer
	count int64
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.count += int64(n)
	return n, err
}

func parseRawSVG(id, s string) (symbol SVGSymbol, err error) {
	var iconSVG struct {
		XMLName xml.Name `xml:"svg"`
		Xmlns   string   `xml:"xmlns,attr"`
		ViewBox string   `xml:"viewBox,attr"`
		Fill    string   `xml:"fill,attr"`
		Content string   `xml:",innerxml"`
	}

	if err = xml.Unmarshal([]byte(s), &iconSVG); err != nil {
		return symbol, err
	}
	symbol = SVGSymbol{
		//Fill:    iconSVG.Fill,
		ID:      id,
		ViewBox: iconSVG.ViewBox,
		Content: iconSVG.Content,
	}
	return symbol, nil
}
