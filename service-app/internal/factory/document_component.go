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
	"github.com/ministryofjustice/opg-scanning/internal/parser/lpa115_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/lpa116_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/lpa120_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/lpc_parser"
)

// component defines a registry entry for a document type.
type component struct {
	parser    func([]byte) (interface{}, error)
	validator parser.CommonValidator
}

// Stores the mapping of document types to their respective components.
var componentRegistry = map[string]component{
	"LP1H": {
		parser:    lp1h_parser.Parse,
		validator: lp1h_parser.NewValidator(),
	},
	"LP1F": {
		parser:    lp1f_parser.Parse,
		validator: lp1f_parser.NewValidator(),
	},
	"Correspondence": {
		parser:    corresp_parser.Parse,
		validator: corresp_parser.NewValidator(),
	},
	"LPC": {
		parser:    lpc_parser.Parse,
		validator: lpc_parser.NewValidator(),
	},
	"LPA115": {
		parser: lpa115_parser.Parse,
	},
	"LPA116": {
		parser: lpa116_parser.Parse,
	},
	"LPA120": {
		parser: lpa120_parser.Parse,
	},
	"LP2": {
		parser:    lp2_parser.Parse,
		validator: lp2_parser.NewValidator(),
	},
}

// Returns the component for the specified document type.
func getComponent(docType string) (component, error) {
	if component, exists := componentRegistry[docType]; exists {
		return component, nil
	}

	if slices.Contains(constants.SupportedDocumentTypes, docType) {
		return component{
			parser: generic_parser.Parse,
		}, nil
	}

	return component{}, fmt.Errorf("unsupported docType: %s", docType)
}
