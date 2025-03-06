package lpa115_types

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1f_types"
)

type LPA115Document struct {
	XMLName xml.Name `xml:"LPA115"`
	Page3   []Page3  `xml:"Page3"`
	Page6   []Page6  `xml:"Page6"`
}

type Page3 struct {
	PartA             PartA3 `xml:"PartA"`
	Signature         bool   `xml:"Signature"`
	Date              string `xml:"Date"`
	ContinuationSheet int    `xml:"ContinuationSheetNo"`
	TotalSheets       string `xml:"TotalSheets"`
	types.BasePage
}

type PartA3 struct {
	FullName string                 `xml:"FullName"`
	OptionA  lp1f_types.Declaration `xml:"OptionA"`
	OptionB  lp1f_types.Declaration `xml:"OptionB"`
}

type Page6 struct {
	PartB PartBPage6 `xml:"PartB"`
	types.BasePage
}

type PartBPage6 struct {
	Salutation        lp1f_types.Salutation `xml:"Salutation"`
	FirstName         string                `xml:"FirstName"`
	LastName          string                `xml:"LastName"`
	Address           lp1f_types.Address    `xml:"Address"`
	Signature         bool                  `xml:"Signature"`
	Date              string                `xml:"Date"`
	ContinuationSheet int                   `xml:"ContinuationSheetNo"`
	TotalSheets       int                   `xml:"TotalSheets"`
}
