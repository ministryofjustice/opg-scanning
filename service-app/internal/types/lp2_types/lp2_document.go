package lp2_types

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1f_types"
)

type LP2Document struct {
	XMLName  xml.Name   `xml:"LP2"`
	Page1    Page1      `xml:"Page1"`
	Page2    Page2      `xml:"Page2"`
	Page3    Page3      `xml:"Page3"`
	Page4    Page4      `xml:"Page4"`
	Page5    Page5      `xml:"Page5"`
	Page6    []Page6    `xml:"Page6"`            
	InfoPage []InfoPage `xml:"InfoPage,omitempty"`
}

type Page1 struct {
	Section1  Section1    `xml:"Section1"`
	types.BasePage  
}

type Section1 struct {
	Title                    string `xml:"Title"`
	FirstName                string `xml:"FirstName"`
	LastName                 string `xml:"LastName"`
	PropertyFinancialAffairs bool   `xml:"PropertyFinancialAffairs"`
	HealthWelfare            bool   `xml:"HealthWelfare"`
}

type Page2 struct {
	Section2  Section2    `xml:"Section2"`
	types.BasePage
}

type Section2 struct {
	DonorRegisteration    bool       `xml:"DonorRegisteration"`
	AttorneyRegisteration bool       `xml:"AttorneyRegisteration"`
	Attorney              []Attorney `xml:"Attorney"`
}

type Attorney struct {
	Title     string `xml:"Title"`
	FirstName string `xml:"FirstName"`
	LastName  string `xml:"LastName"`
	DOB       string `xml:"DOB"`
}

type Page3 struct {
	Section3  Section3    `xml:"Section3"`
	types.BasePage
}

type Section3 struct {
	TheDonor     bool               `xml:"TheDonor"`
	AnAttorney   bool               `xml:"AnAttorney"`
	Other        bool               `xml:"Other"`
	Title        string             `xml:"Title"`
	FirstName    string             `xml:"FirstName"`
	LastName     string             `xml:"LastName"`
	Company      string             `xml:"Company"`
	Address      lp1f_types.Address `xml:"Address"`
	Post         bool               `xml:"Post"`
	Phone        bool               `xml:"Phone"`
	PhoneNumber  string             `xml:"PhoneNumber"`
	Email        bool               `xml:"Email"`
	EmailAddress string             `xml:"EmailAddress"`
	Welsh        bool               `xml:"Welsh"`
}

type Page4 struct {
	Section4  Section4    `xml:"Section4"`
	types.BasePage
}

type Section4 struct {
	Cheque      bool   `xml:"Cheque"`
	Card        bool   `xml:"Card"`
	PhoneNumber string `xml:"PhoneNumber"`
	ReducedFee  bool   `xml:"ReducedFee"`
}

type Page5 struct {
	Section5  Section5    `xml:"Section5"`
	types.BasePage
}

type Section5 struct {
	Attorney []lp1f_types.Declaration `xml:"Attorney"`
}

type Page6 struct {
	Section6  Section6    `xml:"Section6"`
	types.BasePage
}

type Section6 struct {
	Addresses []AddressEntry `xml:"Addresses"`
}

type AddressEntry struct {
	Title        string             `xml:"Title"`
	FirstName    string             `xml:"FirstName"`
	LastName     string             `xml:"LastName"`
	Address      lp1f_types.Address `xml:"Address"`
	EmailAddress string             `xml:"EmailAddress"`
}

type InfoPage struct {
	types.BasePage
}
