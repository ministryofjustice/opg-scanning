package lpa115_types

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1f_types"
)

type LPA115Document struct {
	XMLName  xml.Name `xml:"LPA115"`
	Page1    Page     `xml:"Page1"`
	PartA    PartA    `xml:"PartA"`
	Attorney Attorney `xml:"Attorney"`
}

type Page struct {
	types.BasePage
}

type PartA struct {
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
