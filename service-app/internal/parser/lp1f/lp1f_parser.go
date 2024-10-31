package lp1f

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type LP1FParser struct {
	Doc *types.LP1FDocument
}

func (lp *LP1FParser) ParseDocument(data []byte) (interface{}, error) {
	if lp.Doc == nil {
		return nil, fmt.Errorf("document is not populated")
	}

	if err := lp.ParseAttorneys(); err != nil {
		return nil, err
	}

	return lp.Doc, nil
}

func (lp *LP1FParser) ParseAttorneys() error {
	attorneys := lp.Doc.GetAttorneys()
	for _, attorney := range attorneys {
		if err := lp.processAttorney(attorney); err != nil {
			return fmt.Errorf("error processing attorney: %v", err)
		}
	}
	return nil
}

func (lp *LP1FParser) processAttorney(attorney types.Attorney) error {
	fmt.Printf("Processing Attorney: %s %s\n", attorney.FirstName, attorney.LastName)
	return nil
}
