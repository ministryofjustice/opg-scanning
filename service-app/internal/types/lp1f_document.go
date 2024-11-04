package types

import "encoding/xml"

// For now we force XML but we could negotiate in the future
type LP1FDocument struct {
	XMLName xml.Name `xml:"LP1F" required:"true"`
	Page1   Page1    `xml:"Page1" required:"true"`
	Page2   Page2    `xml:"Page2"`
	Page3   Page3    `xml:"Page3"`
	Page4   Page4    `xml:"Page4"`
}

// GetAttorneys provides a unified way to access attorney data.
func (doc *LP1FDocument) GetAttorneys() []Attorney {
	attorneys := []Attorney{}
	if doc.Page2.Section2.Attorney1 != (Attorney{}) {
		attorneys = append(attorneys, doc.Page2.Section2.Attorney1)
	}
	if doc.Page2.Section2.Attorney2 != (Attorney{}) {
		attorneys = append(attorneys, doc.Page2.Section2.Attorney2)
	}
	return attorneys
}

// Define Page1, Page2, Attorney, etc.
