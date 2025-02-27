package ep2pg_types

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1f_types"
)

type EP2PGDocument struct {
	XMLName  xml.Name   `xml:"EP2PG"`
	Page1    Page1      `xml:"Page1"`
	Page2    Page2      `xml:"Page2"`
	Page3    Page3      `xml:"Page3"`
	Page4    Page4      `xml:"Page4"`
	Page5    Page5      `xml:"Page5"`
	Page6    Page6      `xml:"Page6"`
	Page7    Page7      `xml:"Page7"`
	InfoPage []types.BasePage
}

type Page1 struct {
	Part1        Part1  `xml:"Part1"`
	types.BasePage
}

type Part1 struct {
	Name        lp1f_types.PersonName `xml:"Name"`
	CompanyName string                `xml:"CompanyName"`
	Address     lp1f_types.Address    `xml:"Address"`
	DOB         string                `xml:"DOB"`
}

type Page2 struct {
	Part2        Part2  `xml:"Part2"`
	types.BasePage
}

type Part2 struct {
	Attorney Attorney `xml:"Attorney"`
}

type Attorney struct {
	Name         lp1f_types.PersonName `xml:"Name"`
	CompanyName  string                `xml:"CompanyName"`
	Address      lp1f_types.Address    `xml:"Address"`
	DXDetails    DXDetails             `xml:"DXDetails"`
	DOB          string                `xml:"DOB"`
	Telephone    string                `xml:"Telephone"`
	Email        string                `xml:"Email"`
	Occupation   string                `xml:"Occupation"`
	Relationship EP2PGRelationship     `xml:"Relationship"`
}

type DXDetails struct {
	DXNumber   string `xml:"DXNumber"`
	DXExchange string `xml:"DXExchange"`
}

type EP2PGRelationship struct {
	CivilPartnerSpouse bool   `xml:"CivilPartnerSpouse"`
	Child              bool   `xml:"Child"`
	OtherRelation      bool   `xml:"OtherRelation"`
	NoRelation         bool   `xml:"NoRelation"`
	Solicitor          bool   `xml:"Solicitor"`
	OtherProfessional  bool   `xml:"OtherProfessional"`
	OtherName          string `xml:"OtherName"`
}

type Page3 struct {
	Part3        Part3  `xml:"Part3"`
	Part4        Part4  `xml:"Part4"`
	types.BasePage
}

type Part3 struct {
	Attorney Attorney `xml:"Attorney"`
}

type Part4 struct {
	Salutation lp1f_types.Salutation `xml:"Salutation"`
	LastName   string                `xml:"LastName"`
	Forename   string                `xml:"Forename"`
}

type Page4 struct {
	Part4        Page4Part4 `xml:"Part4"`
	Part5        Part5      `xml:"Part5"`
	types.BasePage
}

type Page4Part4 struct {
	OtherForenames string              `xml:"OtherForenames"`
	CompanyName    string              `xml:"CompanyName"`
	Address        lp1f_types.Address  `xml:"Address"`
	DXDetails      DXDetails           `xml:"DXDetails"`
	DOB            string              `xml:"DOB"`
	Telephone      string              `xml:"Telephone"`
	Email          string              `xml:"Email"`
	Occupation     string              `xml:"Occupation"`
	Relationship   EP2PGRelationship   `xml:"Relationship"`
}

type Part5 struct {
	Date    string               `xml:"Date"`
	YesorNo lp1f_types.YesOrNo   `xml:"YesorNo"`
	Details string               `xml:"Details"`
}

type Page5 struct {
	Part6        Part6  `xml:"Part6"`
	Part7        Part7  `xml:"Part7"`
	types.BasePage
}

type Part6 struct {
	Date     string               `xml:"Date"`
	Address  lp1f_types.Address   `xml:"Address"`
	FullName string               `xml:"FullName"`
}

type Part7 struct {
	NoneEntitled bool       `xml:"NoneEntitled"`
	Relative     []Relative `xml:"Relative"`
}

type Relative struct {
	FullName    string `xml:"FullName"`
	Relationship string `xml:"Relationship"`
	Address1    string `xml:"Address1"`
	Address2    string `xml:"Address2"`
	Address3    string `xml:"Address3"`
	Date        string `xml:"Date"`
}

type Page6 struct {
	Part8        Part8  `xml:"Part8"`
	Part9        Part9  `xml:"Part9"`
	Part10       Part10 `xml:"Part10"`
	BURN         string `xml:"BURN"`
	PhysicalPage int    `xml:"PhysicalPage"`
}

type Part8 struct {
	YesorNo  lp1f_types.YesOrNo `xml:"YesorNo"`
	Relative []Relative         `xml:"Relative"`
}

type Part9 struct {
	ChequeFee          ChequeFee          `xml:"ChequeFee"`
	ExemptionRemission ExemptionRemission `xml:"ExemptionRemission"`
}

type ChequeFee struct {
	YesorNo lp1f_types.YesOrNo `xml:"YesorNo"`
}

type ExemptionRemission struct {
	YesorNo lp1f_types.YesOrNo `xml:"YesorNo"`
}

type Part10 struct {
	Declaration lp1f_types.Declaration `xml:"Declaration"`
}

type Page7 struct {
	Part11       Part11 `xml:"Part11"`
	Part12       Part12 `xml:"Part12"`
	BURN         string `xml:"BURN"`
	PhysicalPage int    `xml:"PhysicalPage"`
}

type Part11 struct {
	Name             lp1f_types.PersonName `xml:"Name"`
	CompanyName      string                `xml:"CompanyName"`
	CompanyReference string                `xml:"CompanyReference"`
	Address          lp1f_types.Address    `xml:"Address"`
	DXDetails        DXDetails             `xml:"DXDetails"`
	Telephone        string                `xml:"Telephone"`
	Email            string                `xml:"Email"`
}

type Part12 struct {
	AdditionalInfo string `xml:"AdditionalInfo"`
}