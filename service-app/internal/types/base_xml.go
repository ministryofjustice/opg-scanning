package types

import "encoding/xml"

type Set struct {
	XMLName xml.Name `xml:"Set"`
	Header  *Header  `xml:"Header"`
	Body    Body     `xml:"Body"`
}

type Header struct {
	CaseNo          string `xml:"CaseNo,attr"`
	Scanner         string `xml:"Scanner,attr"`
	ScanTime        string `xml:"ScanTime,attr"`
	ScannerOperator string `xml:"ScannerOperator,attr"`
	Schedule        string `xml:"Schedule,attr"`
	FeeNumber       string `xml:"FeeNumber,attr"`
}

type Body struct {
	Documents []Document `xml:"Document"`
}

type Document struct {
	Type          string `xml:"Type,attr"`
	Encoding      string `xml:"Encoding,attr"`
	NoPages       int    `xml:"NoPages,attr"`
	EmbeddedXML   string `xml:"XML"`
	EmbeddedImage string `xml:"Image"`
}
