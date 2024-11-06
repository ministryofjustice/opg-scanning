package lpf1_parser

import (
	"encoding/xml"
	"fmt"

	lp1f_types "github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
)

type LP1FParser struct {
	Doc *lp1f_types.LP1FDocument
}

func (lp *LP1FParser) ParseDocument(data []byte) (interface{}, error) {
	if lp.Doc == nil {
		return nil, fmt.Errorf("document is not populated")
	}

	// Assuming the input data is XML, we parse it into the LP1FDocument struct
	if err := xml.Unmarshal(data, lp.Doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %v", err)
	}

	// TODO: Perform necessary data transformations

	return lp.Doc, nil
}
