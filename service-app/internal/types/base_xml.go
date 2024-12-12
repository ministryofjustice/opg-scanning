package types

import "encoding/xml"

type BaseSet struct {
	XMLName xml.Name    `xml:"Set" required:"true"`
	Header  *BaseHeader `xml:"Header" required:"true"`
	Body    BaseBody    `xml:"Body" required:"true"`
}

type BaseHeader struct {
	CaseNo          string `xml:"CaseNo,attr"`
	Scanner         string `xml:"Scanner,attr"`
	ScanTime        string `xml:"ScanTime,attr" required:"true"`
	ScannerOperator string `xml:"ScannerOperator,attr"`
	Schedule        string `xml:"Schedule,attr"`
	FeeNumber       string `xml:"FeeNumber,attr"`
}

type BaseBody struct {
	Documents []BaseDocument `xml:"Document" required:"true"`
}

type BaseDocument struct {
	Type        string `xml:"Type,attr"`
	Encoding    string `xml:"Encoding,attr"`
	NoPages     int    `xml:"NoPages,attr"`
	EmbeddedXML string `xml:"XML" required:"true"`
	EmbeddedPDF string `xml:"PDF" required:"true"`
}

type BasePage struct {
	BURN         string `xml:"BURN"`
	PhysicalPage string `xml:"PhysicalPage"`
}
