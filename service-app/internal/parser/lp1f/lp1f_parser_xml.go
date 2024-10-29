package lp1f

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

func ParseLP1FXml(data []byte) (*types.LP1FDocument, error) {
	var doc types.LP1FDocument
	err := xml.Unmarshal(data, &doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}
