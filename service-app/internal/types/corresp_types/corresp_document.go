package corresp_types

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type Correspondence struct {
	XMLName    xml.Name         `xml:"Correspondence"`
	SubType    string           `xml:"SubType"`
	CaseNumber string           `xml:"CaseNumber"`
	Pages      []types.BasePage `xml:"Page"`
}
