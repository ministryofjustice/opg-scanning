package factory

import (
	"fmt"
	"slices"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/corresp_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/generic_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/lp1f_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/lp1h_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/lp2_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/lpa120_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/lpc_parser"
)

// Component defines a registry entry for a document type.
type Component struct {
	Parser    func([]byte) (interface{}, error)
	Validator parser.CommonValidator
	Sanitizer parser.CommonSanitizer
}

// Stores the mapping of document types to their respective components.
var componentRegistry = map[string]Component{
	"LP1H": {
		Parser:    lp1h_parser.Parse,
		Validator: lp1h_parser.NewValidator(),
		Sanitizer: lp1h_parser.NewSanitizer(),
	},
	"LP1F": {
		Parser:    lp1f_parser.Parse,
		Validator: lp1f_parser.NewValidator(),
		Sanitizer: lp1f_parser.NewSanitizer(),
	},
	"Correspondence": {
		Parser:    corresp_parser.Parse,
		Validator: corresp_parser.NewValidator(),
		Sanitizer: corresp_parser.NewSanitizer(),
	},
	"LPC": {
		Parser:    lpc_parser.Parse,
		Validator: lpc_parser.NewValidator(),
		Sanitizer: lpc_parser.NewSanitizer(),
	},
	"LPA120": {	
		Parser: lpa120_parser.Parse,
	},
	"LP2": {
		Parser: lp2_parser.Parse,
		Validator: lp2_parser.NewValidator(),
	},
}

// Returns the component for the specified document type.
func GetComponent(docType string) (Component, error) {
	if component, exists := componentRegistry[docType]; exists {
		return component, nil
	}

	if slices.Contains(constants.SupportedDocumentTypes, docType) {
		return Component{
			Parser: generic_parser.Parse,
		}, nil
	}

	return Component{}, fmt.Errorf("unsupported docType: %s", docType)
}
