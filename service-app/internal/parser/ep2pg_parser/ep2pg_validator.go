package ep2pg_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/ep2pg_types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type Validator struct {
	doc           *ep2pg_types.EP2PGDocument
	baseValidator *parser.BaseValidator
}

func NewValidator() *Validator {
	return &Validator{
		doc: &ep2pg_types.EP2PGDocument{},
	}
}

func (v *Validator) Setup(doc any) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	v.doc = doc.(*ep2pg_types.EP2PGDocument)
	v.baseValidator = parser.NewBaseValidator(v.doc)

	return nil
}

func (v *Validator) Validate() []string {
	if v.doc.Page1.Part1.DOB != "" {
		if _, err := util.ParseDate(v.doc.Page1.Part1.DOB, "02012006"); err != nil {
			v.baseValidator.AddValidatorErrorMessage("Failed to parse date of birth for Donor: " + err.Error())
		}
	}

	if v.doc.Page4.Part4.DOB != "" {
		if _, err := util.ParseDate(v.doc.Page4.Part4.DOB, "02012006"); err != nil {
			v.baseValidator.AddValidatorErrorMessage("Failed to parse date of birth for Attorney: " + err.Error())
		}
	}

	return v.baseValidator.GetValidatorErrorMessages()
}
