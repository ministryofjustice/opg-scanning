package types

import "encoding/xml"

type BaseSet struct {
	XMLName xml.Name    `xml:"Set"`
	Header  *BaseHeader `xml:"Header"`
	Body    BaseBody    `xml:"Body"`
}

type BaseHeader struct {
	CaseNo          string `xml:"CaseNo,attr"`
	Scanner         string `xml:"Scanner,attr"`
	ScanTime        string `xml:"ScanTime,attr"`
	ScannerOperator string `xml:"ScannerOperator,attr"`
	Schedule        string `xml:"Schedule,attr"`
	FeeNumber       string `xml:"FeeNumber,attr"`
}

type BaseBody struct {
	Documents []BaseDocument `xml:"Document"`
}

type BaseDocument struct {
	Type          string `xml:"Type,attr"`
	Encoding      string `xml:"Encoding,attr"`
	NoPages       int    `xml:"NoPages,attr"`
	EmbeddedXML   string `xml:"XML"`
	EmbeddedImage string `xml:"Image"`
}

type BasePage struct {
	BURN         string `xml:"BURN"`
	PhysicalPage string `xml:"PhysicalPage"`
}
