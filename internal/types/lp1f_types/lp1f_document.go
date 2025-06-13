package lp1f_types

import "encoding/xml"

// For now we assume XML but we could negotiate in the future
type LP1FDocument struct {
	XMLName           xml.Name            `xml:"LP1F" required:"true"`
	Page1             Page1               `xml:"Page1"`
	Page2             Page2               `xml:"Page2"`
	Page3             Page3               `xml:"Page3"`
	Page4             Page4               `xml:"Page4"`
	Page5             Page5               `xml:"Page5"`
	Page6             Page6               `xml:"Page6"`
	Page7             Page7               `xml:"Page7"`
	Page8             Page8               `xml:"Page8"`
	Page9             Page9               `xml:"Page9"`
	Page10            Page10              `xml:"Page10"`
	Page11            Page11              `xml:"Page11"`
	Page12            []Page12            `xml:"Page12"`
	Page16            Page16              `xml:"Page16,omitempty"`
	Page17            Page17              `xml:"Page17"`
	Page18            Page18              `xml:"Page18"`
	Page19            Page19              `xml:"Page19"`
	Page20            []Page20            `xml:"Page20"`
	ContinuationPage1 []ContinuationPage1 `xml:"ContinuationPage1,omitempty"`
	ContinuationPage2 []ContinuationPage2 `xml:"ContinuationPage2,omitempty"`
	ContinuationPage3 []ContinuationPage3 `xml:"ContinuationPage3,omitempty"`
	ContinuationPage4 []ContinuationPage4 `xml:"ContinuationPage4,omitempty"`
	InfoPage          []InfoPage          `xml:"InfoPage,omitempty"`
}
