package types

type Page1 struct {
	Section1     Section1 `xml:"Section1"`
	BURN         string   `xml:"BURN,omitempty"`
	PhysicalPage string   `xml:"PhysicalPage,omitempty"`
}

type Section1 struct {
	Title        string `xml:"Title"`
	FirstName    string `xml:"FirstName"`
	LastName     string `xml:"LastName"`
	OtherNames   string `xml:"OtherNames"`
	DOB          string `xml:"DOB"`
	Address      string `xml:"Address"`
	EmailAddress string `xml:"EmailAddress"`
}

type Page2 struct {
	Section2     Section2 `xml:"Section2"`
	BURN         string   `xml:"BURN,omitempty"`
	PhysicalPage string   `xml:"PhysicalPage,omitempty"`
}

type Section2 struct {
	Attorney1 Attorney `xml:"Attorney1"`
	Attorney2 Attorney `xml:"Attorney2"`
}

type Attorney struct {
	Title            string `xml:"Title"`
	FirstName        string `xml:"FirstName"`
	LastName         string `xml:"LastName"`
	DOB              string `xml:"DOB"`
	Address          string `xml:"Address"`
	EmailAddress     string `xml:"EmailAddress"`
	TrustCorporation *bool  `xml:"TrustCorporation,omitempty"` // Optional boolean field
}

type Page3 struct {
	Section2     Section2B `xml:"Section2"`
	BURN         string    `xml:"BURN,omitempty"`
	PhysicalPage string    `xml:"PhysicalPage,omitempty"`
}

type Section2B struct {
	Attorney      []Attorney `xml:"Attorney"` // Fixed-size array for exactly 2 attorneys based on XSD
	MoreAttorneys *bool      `xml:"MoreAttorneys,omitempty"`
}

type Page4 struct {
	Section3     Section3 `xml:"Section3"`
	BURN         string   `xml:"BURN,omitempty"`
	PhysicalPage string   `xml:"PhysicalPage,omitempty"`
}

type Section3 struct {
	AppointedOneAttorney bool `xml:"AppointedOneAttorney"`
	JointlyAndSeverally  bool `xml:"JointlyAndSeverally"`
	Jointly              bool `xml:"Jointly"`
	JointlyForSome       bool `xml:"JointlyForSome"`
}
