package lp1h_types

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
)

type LP1HDocument struct {
	lpf1_types.LP1FDocument
	XMLName xml.Name `xml:"LP1H" required:"true"`
	OptionA string   `xml:"OptionA"`
	OptionB string   `xml:"OptionB"`
}
