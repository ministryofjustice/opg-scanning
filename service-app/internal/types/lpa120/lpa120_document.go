package lpa120_document

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1f_types"
)

type LPA120Document struct {
	XMLName   xml.Name `xml:"LPA120"`
	Page1     Page     `xml:"Page1"`
	Page2     Page     `xml:"Page2"`
	Page3     Page3    `xml:"Page3"`
	Page4     Page4    `xml:"Page4"`
	InfoPages []Page   `xml:"InfoPage,omitempty"`
}

type Page struct {
	types.BasePage
}

type Page3 struct {
	Section1 Section1 `xml:"Section1"`
	Section2 Section2 `xml:"Section2"`
	types.BasePage
}

type Section1 struct {
	FullName                string             `xml:"FullName"`
	Address                 lp1f_types.Address `xml:"Address"`
	CaseReference           string             `xml:"CaseReference"`
	LPAApplicationFee       bool               `xml:"LPAApplicationFee"`
	EPAApplicationFee       bool               `xml:"EPAApplicationFee"`
	RepeatLPAApplicationFee bool               `xml:"RepeatLPAApplicationFee"`
	LPAHealthWelfare        bool               `xml:"LPAHealthWelfare"`
	LPAPropertyFinance      bool               `xml:"LPAPropertyFinance"`
	EPA                     bool               `xml:"EPA"`
}

type Section2 struct {
	Relationship Relationship          `xml:"Relationship"`
	Salutation   lp1f_types.Salutation `xml:"Salutation"`
	FirstName    string                `xml:"FirstName"`
	LastName     string                `xml:"LastName"`
	Address      lp1f_types.Address
	Telephone    string `xml:"Telephone"`
	Mobile       string `xml:"Mobile"`
	Email        string `xml:"Email"`
	FeePaidTo    string `xml:"FeePaidTo"`
}

type Page4 struct {
	Section3 Section3 `xml:"Section3"`
	Section4 Section4 `xml:"Section4"`
	Section5 Section5 `xml:"Section5"`
	types.BasePage
}

type Section3 struct {
	DonorReceivesBenefits lp1f_types.YesOrNo `xml:"DonorReceivesBenefits"`
	DonorAwardedInjuries  lp1f_types.YesOrNo `xml:"DonorAwardedInjuries"`
}

type Section4 struct {
	DonorOver12000               lp1f_types.YesOrNo `xml:"DonorOver12000"`
	DonorReceivesUniversalCredit lp1f_types.YesOrNo `xml:"DonorReceivesUniversalCredit"`
}

type Section5 struct {
	Declaration lp1f_types.Declaration `xml:"Declaration"`
}

type Relationship struct {
	Donor       bool   `xml:"Donor"`
	Attorney    bool   `xml:"Attorney"`
	Other       bool   `xml:"Other"`
	OtherDetail string `xml:"OtherDetail,omitempty"`
}
