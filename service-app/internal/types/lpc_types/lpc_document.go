package lpc_types

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1f_types"
)

type LPCDocument struct {
	XMLName xml.Name `xml:"LPC"`
	Page1   []Page1  `xml:"Page1,omitempty"`
	Page2   []Page2  `xml:"Page2,omitempty"`
	Page3   []Page3  `xml:"Page3,omitempty"`
	Page4   []Page4  `xml:"Page4,omitempty"`
}

type Page1 struct {
	XMLName            xml.Name           `xml:"Page1"`
	ContinuationSheet1 ContinuationSheet1 `xml:"ContinuationSheet1"`
	types.BasePage
}

type ContinuationSheet1 struct {
	Attorneys []AttorneyContinuation1     `xml:"Attorney"`
	Donor     lp1f_types.AuthorisedPerson `xml:"Donor"`
}

type AttorneyContinuation1 struct {
	Attorney            bool   `xml:"Attorney"`
	ReplacementAttorney bool   `xml:"ReplacementAttorney"`
	PersonToNotify      bool   `xml:"PersonToNotify"`
	Title               string `xml:"Title"`
	FirstName           string `xml:"FirstName"`
	LastName            string `xml:"LastName"`
	DOB                 string `xml:"DOB"`
	Address             lp1f_types.Address
	Email               string        `xml:"Email"`
	Relationship        *Relationship `xml:"Relationship,omitempty"`
}

type Page2 struct {
	XMLName            xml.Name           `xml:"Page2"`
	ContinuationSheet2 ContinuationSheet2 `xml:"ContinuationSheet2"`
	types.BasePage
}

type ContinuationSheet2 struct {
	AdditionalInformation lp1f_types.AdditionalInformation `xml:"AdditionalInformation"`
	Donor                 lp1f_types.AuthorisedPerson      `xml:"Donor"`
}

type Page3 struct {
	XMLName            xml.Name           `xml:"Page3"`
	ContinuationSheet3 ContinuationSheet3 `xml:"ContinuationSheet3"`
	types.BasePage
}

type ContinuationSheet3 struct {
	Donor     DonorNameOnly               `xml:"Donor"`
	Signatory lp1f_types.AuthorisedPerson `xml:"Signatory"`
	Witnesses []WitnessContinuation3      `xml:"Witnesses"`
}

type DonorNameOnly struct {
	FullName string `xml:"FullName"`
}

type WitnessContinuation3 struct {
	Signature bool               `xml:"Signature"`
	FullName  string             `xml:"FullName"`
	Address   lp1f_types.Address `xml:"Address"`
}

type Page4 struct {
	XMLName            xml.Name           `xml:"Page4"`
	ContinuationSheet4 ContinuationSheet4 `xml:"ContinuationSheet4"`
	types.BasePage
}

type ContinuationSheet4 struct {
	CompanyRegistration string                        `xml:"CompanyRegistration"`
	AuthorisedPerson    []lp1f_types.AuthorisedPerson `xml:"AuthorisedPerson"`
}

type Relationship struct {
	CivilPartnerSpouse bool   `xml:"CivilPartnerSpouse"`
	Child              bool   `xml:"Child"`
	Solicitor          bool   `xml:"Solicitor"`
	Other              bool   `xml:"Other"`
	OtherProfessional  bool   `xml:"OtherProfessional"`
	OtherName          string `xml:"OtherName,omitempty"`
}
