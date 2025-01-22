package lp1h_types

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/types/lp1f_types"
)

type LP1HDocument struct {
	lp1f_types.LP1FDocument
	XMLName xml.Name `xml:"LP1H" required:"true"`
	Page6   Page6    `xml:"Page6"`
	Page10  Page10   `xml:"Page10"`
}

type Page10 struct {
	Section9 Section9 `xml:"Section9"`
}

type Page6 struct {
	Section5 Section5 `xml:"Section5"`
}

type Section9 struct {
	Donor   SignatureAndDOB    `xml:"Donor"`
	Witness lp1f_types.Witness `xml:"Witness"`
}

type Section5 struct {
	OptionA SignatureAndDOB    `xml:"OptionA"`
	OptionB SignatureAndDOB    `xml:"OptionB"`
	Witness lp1f_types.Witness `xml:"Witness"`
}

type SignatureAndDOB struct {
	Signature string `xml:"Signature"`
	DOB       string `xml:"DOB"`
}
