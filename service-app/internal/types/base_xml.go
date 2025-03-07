package types

import "encoding/xml"

// AnyElement can capture unexpected XML elements.
type AnyElement struct {
	XMLName  xml.Name
	InnerXML string `xml:",innerxml"`
}

type BaseSet struct {
	XMLName xml.Name    `xml:"Set"`
	Header  *BaseHeader `xml:"Header"`
	Body    BaseBody    `xml:"Body"`
	// Catch all for unexpected elements at the Set level
	ExtraElements []AnyElement `xml:",any"`
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
	Type        string `xml:"Type,attr"`
	Encoding    string `xml:"Encoding,attr"`
	NoPages     int    `xml:"NoPages,attr"`
	EmbeddedXML string `xml:"XML"`
	EmbeddedPDF string `xml:"PDF"`
}

type BasePage struct {
	BURN         string `xml:"BURN"`
	PhysicalPage string `xml:"PhysicalPage"`
}
