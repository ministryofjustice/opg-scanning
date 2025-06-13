package lpa116_types

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1f_types"
)

type LPA116Document struct {
	XMLName  xml.Name `xml:"LPA116"`
	Page1    Page     `xml:"Page1"`
	PartC    PartC    `xml:"PartC"`
	Attorney Attorney `xml:"Attorney"`
}

type Page struct {
	types.BasePage
}

type PartC struct {
	FullName      string             `xml:"FullName"`
	Address       lp1f_types.Address `xml:"Address"`
	CaseReference string             `xml:"CaseReference"`
	Declaration   string             `xml:"Declaration"`
}

type Attorney struct {
	FirstName string `xml:"FirstName"`
	LastName  string `xml:"LastName"`
	Address   lp1f_types.Address
	Telephone string `xml:"Telephone"`
	Email     string `xml:"Email"`
}
