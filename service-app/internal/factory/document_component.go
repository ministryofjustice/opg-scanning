package factory

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/corresp_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/generic_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/lp1f_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/lp1h_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/lpc_parser"
)

// Component defines a registry entry for a document type.
type Component struct {
	Parser    func([]byte) (interface{}, error)
	Validator parser.CommonValidator
	Sanitizer parser.CommonSanitizer
}

func GetComponent(docType string) (Component, error) {
	switch docType {
	case "LP1H":
		return Component{
			Parser: func(data []byte) (interface{}, error) {
				return lp1h_parser.Parse(data)
			},
			Validator: lp1h_parser.NewValidator(),
			Sanitizer: lp1h_parser.NewSanitizer(),
		}, nil
	case "LP1F":
		return Component{
			Parser: func(data []byte) (interface{}, error) {
				return lp1f_parser.Parse(data)
			},
			Validator: lp1f_parser.NewValidator(),
			Sanitizer: lp1f_parser.NewSanitizer(),
		}, nil
	case "Correspondence":
		return Component{
			Parser: func(data []byte) (interface{}, error) {
				return corresp_parser.Parse(data)
			},
			Validator: corresp_parser.NewValidator(),
			Sanitizer: corresp_parser.NewSanitizer(),
		}, nil
	case "LPC":
		return Component{
			Parser: func(data []byte) (interface{}, error) {
				return lpc_parser.Parse(data)
			},
			Validator: lpc_parser.NewValidator(),
			Sanitizer: lpc_parser.NewSanitizer(),
		}, nil
	case constants.DocumentTypeLPA002, constants.DocumentTypeLPAPA, constants.DocumentTypeLPAPW, constants.DocumentTypeLPA114, constants.DocumentTypeLPA117:
		return Component{
			Parser: func(data []byte) (interface{}, error) {
				return generic_parser.Parse(data)
			},
			Validator: generic_parser.NewValidator(),
			Sanitizer: generic_parser.NewSanitizer(),
		}, nil
	default:
		return Component{}, fmt.Errorf("unsupported docType: %s", docType)
	}
}
