package ingestion

import (
	"errors"
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

var Instruments = []string{"EPA", "LPA-PA", "LPA-PW", "LPA114", "LPA117", "LP1F", "LP1H"}
var Applications = []string{"EP2PG", "LPA002", "LPA002R", "LP2"}
var StandaloneInstruments = []string{"LP1F", "LP1H"}
var ExemptApplications = []string{"LPA002R"}

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateSet(parsedSet *types.Set) error {
	if parsedSet == nil {
		return errors.New("parsedSet is nil")
	}

	// Validate Header fields
	if parsedSet.Header == nil {
		return errors.New("missing required Header element")
	}
	if parsedSet.Header.CaseNo == "" {
		return errors.New("missing CaseNo in Header")
	}
	if parsedSet.Header.ScanTime == "" {
		return errors.New("missing ScanTime in Header")
	}

	// Validate combinations of instruments and applications
	instrumentsDiscovered := v.getEmbeddedDocumentTypes(parsedSet, Instruments)
	applicationsDiscovered := v.getEmbeddedDocumentTypes(parsedSet, Applications)

	if parsedSet.Header.CaseNo != "" {
		// Validate document combinations if CaseNo exists
		if err := v.validateDocCombosWithCaseNo(instrumentsDiscovered, applicationsDiscovered); err != nil {
			return err
		}
	} else {
		// Validate document instruments if no CaseNo exists
		if err := v.validateInstrumentCountWithoutCaseNo(instrumentsDiscovered); err != nil {
			return err
		}
		if err := v.validateInstrumentApplications(instrumentsDiscovered, applicationsDiscovered); err != nil {
			return err
		}
	}

	// Validate Body and Documents
	if len(parsedSet.Body.Documents) == 0 {
		return errors.New("no Document elements found in Body")
	}

	for _, doc := range parsedSet.Body.Documents {
		if doc.Type == "" {
			return fmt.Errorf("document Type attribute is missing")
		}
		if doc.NoPages <= 0 {
			return fmt.Errorf("document NoPages attribute is missing or invalid")
		}

		// TODO: Check if validate Embedded XML is needed at this stage
	}

	return nil
}

func (v *Validator) validateDocCombosWithCaseNo(instruments []string, applications []string) error {
	if len(instruments) == 0 && len(applications) == 0 {
		return nil
	}

	if len(applications) == 1 && v.isExemptApplication(applications[0]) {
		return nil
	}

	fullList := append(instruments, applications...)
	return fmt.Errorf("document(s) %s cannot be used if you have set a CaseNo in the Header", fullList)
}

func (v *Validator) validateInstrumentCountWithoutCaseNo(instruments []string) error {
	if len(instruments) == 0 {
		return fmt.Errorf("no instrument found. Valid instruments are %s", Instruments)
	}
	if len(instruments) > 1 {
		return fmt.Errorf("too many instruments found. You may only supply one instrument. Set contained %s", instruments)
	}
	return nil
}

func (v *Validator) validateInstrumentApplications(instruments []string, applications []string) error {
	if len(instruments) == 0 {
		return nil
	}
	if v.isStandaloneInstrument(instruments[0]) && len(applications) > 0 {
		return fmt.Errorf("instrument %s must not be accompanied by an application (%s)", instruments[0], applications)
	}
	if !v.isStandaloneInstrument(instruments[0]) && len(applications) != 1 {
		return fmt.Errorf("instrument %s must be accompanied by one application. Found applications: %s", instruments[0], applications)
	}
	return nil
}

func (v *Validator) getEmbeddedDocumentTypes(parsedSet *types.Set, validTypes []string) []string {
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

func (v *Validator) isExemptApplication(application string) bool {
	for _, exempt := range ExemptApplications {
		if application == exempt {
			return true
		}
	}
	return false
}

func (v *Validator) isStandaloneInstrument(instrument string) bool {
	for _, standalone := range StandaloneInstruments {
		if instrument == standalone {
			return true
		}
	}
	return false
}
