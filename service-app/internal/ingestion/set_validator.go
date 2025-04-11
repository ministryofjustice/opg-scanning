package ingestion

import (
	"errors"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateSet(parsedSet *types.BaseSet) error {
	if parsedSet == nil {
		return errors.New("parsedSet is nil")
	}

	// Validate Header fields
	if parsedSet.Header == nil {
		return errors.New("missing required Header element")
	}

	if parsedSet.Header.Schedule == "" {
		return errors.New("missing required Schedule attribute on Header")
	}

	// Validate Body and Documents
	if len(parsedSet.Body.Documents) == 0 {
		return errors.New("no Document elements found in Body")
	}

	for _, doc := range parsedSet.Body.Documents {
		if doc.Type == "" {
			return errors.New("document Type attribute is missing")
		}
		if doc.NoPages <= 0 {
			return errors.New("document NoPages attribute is missing or invalid")
		}
	}

	// Validate combinations of instruments and applications
	newCaseDocuments := v.getEmbeddedDocumentTypes(parsedSet, constants.NewCaseNumberDocuments)

	if len(newCaseDocuments) > 1 {
		return errors.New("set cannot contain multiple cases which would create a case")
	}

	if len(newCaseDocuments) > 0 {
		// Sets that create new cases must not have a case number
		if parsedSet.Header.CaseNo != "" {
			return errors.New("must not supply a case number when creating a new case")
		}
	} else {
		// Sets that don't create new cases must have a case number
		if parsedSet.Header.CaseNo == "" {
			return errors.New("must supply a case number when not creating a new case")
		}
	}

	return nil
}

func (v *Validator) getEmbeddedDocumentTypes(parsedSet *types.BaseSet, validTypes []string) []string {
	typesDiscovered := []string{}
	for _, doc := range parsedSet.Body.Documents {
		for _, validType := range validTypes {
			if doc.Type == validType {
				typesDiscovered = append(typesDiscovered, doc.Type)
				break
			}
		}
	}
	return typesDiscovered
}
