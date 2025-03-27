package lp1f_types

// Address always outputs all four sub-elements.
type Address struct {
	Address1 string `xml:"Address1"`
	Address2 string `xml:"Address2"`
	Address3 string `xml:"Address3"`
	Postcode string `xml:"Postcode"`
}

type Salutation struct {
	Mr        bool   `xml:"Mr"`
	Mrs       bool   `xml:"Mrs"`
	Ms        bool   `xml:"Ms"`
	Miss      bool   `xml:"Miss"`
	Other     bool   `xml:"Other"`
	OtherName string `xml:"OtherName"`
}

type PersonName struct {
	Salutation     Salutation `xml:"Salutation"`
	LastName       string     `xml:"LastName"`
	Forename       string     `xml:"Forename"`
	OtherForenames string     `xml:"OtherForenames"`
}

// Declaration: remove omitempty so Signature and Date always appear.
type Declaration struct {
	Signature bool   `xml:"Signed"`
	Date      string `xml:"Date"`
}

type Notification struct {
	NoticeDate string  `xml:"NoticeDate"`
	LastName   string  `xml:"LastName"`
	FirstName  string  `xml:"FirstName"`
	Address    Address `xml:"Address"`
}

type Appointment struct {
	Jointly             bool `xml:"Jointly"`
	JointlyAndSeverally bool `xml:"JointlyAndSeverally"`
	Alone               bool `xml:"Alone"`
}

type Attorney struct {
	Title            string  `xml:"Title"`
	FirstName        string  `xml:"FirstName"`
	LastName         string  `xml:"LastName"`
	DOB              string  `xml:"DOB"`
	Address          Address `xml:"Address"`
	EmailAddress     string  `xml:"EmailAddress"`
	TrustCorporation bool    `xml:"TrustCorporation"` // not a pointer so it's always output
	Declaration
}

type PeopleToNotify struct {
	Title     string  `xml:"Title"`
	FirstName string  `xml:"FirstName"`
	LastName  string  `xml:"LastName"`
	Address   Address `xml:"Address"`
}

type SkillCertification struct {
	RegisteredProfessional     bool   `xml:"RegisteredProfessional"`
	BarristerSolicitorAdvocate bool   `xml:"BarristerSolicitorAdvocate"`
	SocialWorker               bool   `xml:"SocialWorker"`
	IMCA                       bool   `xml:"IMCA"`
	NoneOfTheAbove             bool   `xml:"NoneOfTheAbove"`
	SkillsAndExpertise         string `xml:"SkillsAndExpertise"`
}

type Witness struct {
	Signature bool    `xml:"Signature"`
	FullName  string  `xml:"FullName"`
	Address   Address `xml:"Address"`
}

type YesOrNo struct {
	Yes bool `xml:"Yes"`
	No  bool `xml:"No"`
}

type Section1 struct {
	Title        string  `xml:"Title"`
	FirstName    string  `xml:"FirstName"`
	LastName     string  `xml:"LastName"`
	OtherNames   string  `xml:"OtherNames"`
	DOB          string  `xml:"DOB"`
	Address      Address `xml:"Address"`
	EmailAddress string  `xml:"EmailAddress"`
}

// Section2 for Page2: uses Attorney1 and Attorney2.
type Section2Page2 struct {
	Attorney1 Attorney `xml:"Attorney1"`
	Attorney2 Attorney `xml:"Attorney2"`
}

// Section2 for Page3: two Attorney elements and MoreAttorneys.
type Section2Page3 struct {
	Attorneys     []Attorney `xml:"Attorney"` // must contain exactly two elements
	MoreAttorneys bool       `xml:"MoreAttorneys"`
}

type Section3 struct {
	AppointedOneAttorney bool `xml:"AppointedOneAttorney"`
	JointlyAndSeverally  bool `xml:"JointlyAndSeverally"`
	Jointly              bool `xml:"Jointly"`
	JointlyForSome       bool `xml:"JointlyForSome"`
}

type Section4 struct {
	Attorney1             Attorney `xml:"Attorney1"`
	Attorney2             Attorney `xml:"Attorney2"`
	MoreReplacements      bool     `xml:"MoreReplacements"`
	ChangeHowAttorneysAct bool     `xml:"ChangeHowAttorneysAct"`
}

type Section5 struct {
	LPARegistered  bool `xml:"LPARegistered"`
	MentalCapacity bool `xml:"MentalCapacity"`
}

type Section6 struct {
	PeopleToNotify []PeopleToNotify `xml:"PeopleToNotify"`
	AppointAnother bool             `xml:"AppointAnother"`
}

type Section7 struct {
	Preferences           bool `xml:"Preferences"`
	PreferencesMoreSpace  bool `xml:"PreferencesMoreSpace"`
	Instructions          bool `xml:"Instructions"`
	InstructionsMoreSpace bool `xml:"InstructionsMoreSpace"`
}

type Section9 struct {
	Donor   Declaration `xml:"Donor"`
	Witness Witness     `xml:"Witness"`
}

type Section10 struct {
	Title     string  `xml:"Title"`
	FirstName string  `xml:"FirstName"`
	LastName  string  `xml:"LastName"`
	Address   Address `xml:"Address"`
	Declaration
}

type Section11 struct {
	Attorney Attorney `xml:"Attorney"`
	Witness  Witness  `xml:"Witness"`
}

type Section12 struct {
	DonorApply    bool       `xml:"DonorApply"`
	AttorneyApply bool       `xml:"AttorneyApply"`
	Attorney      []Attorney `xml:"Attorney"`
}

type Section13 struct {
	TheDonor     bool    `xml:"TheDonor"`
	AnAttorney   bool    `xml:"AnAttorney"`
	Other        bool    `xml:"Other"`
	Title        string  `xml:"Title"`
	FirstName    string  `xml:"FirstName"`
	LastName     string  `xml:"LastName"`
	CompanyName  string  `xml:"CompanyName"`
	Address      Address `xml:"Address"`
	Post         string  `xml:"Post"`
	Phone        string  `xml:"Phone"`
	PhoneNumber  string  `xml:"PhoneNumber"`
	Email        string  `xml:"Email"`
	EmailAddress string  `xml:"EmailAddress"`
	Welsh        bool    `xml:"Welsh"`
}

type Section14 struct {
	Cheque                bool   `xml:"Cheque"`
	Card                  bool   `xml:"Card"`
	PhoneNumber           string `xml:"PhoneNumber"`
	ReducedApplicationFee bool   `xml:"ReducedApplicationFee"`
	RepeatApplication     bool   `xml:"RepeatApplication"`
	CaseNumber            string `xml:"CaseNumber"`
	OnlineLPA             bool   `xml:"OnlineLPA"`
	OnlineLPAID           string `xml:"OnlineLPAID"`
}

type Section15 struct {
	Applicant []Declaration `xml:"Applicant"`
}

// Page structs (each enforces the order: Section, then BURN, then PhysicalPage)

type Page1 struct {
	Section1     Section1 `xml:"Section1"`
	BURN         string   `xml:"BURN"`
	PhysicalPage int      `xml:"PhysicalPage"`
}

type Page2 struct {
	Section2     Section2Page2 `xml:"Section2"`
	BURN         string        `xml:"BURN"`
	PhysicalPage int           `xml:"PhysicalPage"`
}

type Page3 struct {
	Section2     Section2Page3 `xml:"Section2"`
	BURN         string        `xml:"BURN"`
	PhysicalPage int           `xml:"PhysicalPage"`
}

type Page4 struct {
	Section3     Section3 `xml:"Section3"`
	BURN         string   `xml:"BURN"`
	PhysicalPage int      `xml:"PhysicalPage"`
}

type Page5 struct {
	Section4     Section4 `xml:"Section4"`
	BURN         string   `xml:"BURN"`
	PhysicalPage int      `xml:"PhysicalPage"`
}

type Page6 struct {
	Section5     Section5 `xml:"Section5"`
	BURN         string   `xml:"BURN"`
	PhysicalPage int      `xml:"PhysicalPage"`
}

type Page7 struct {
	Section6     Section6 `xml:"Section6"`
	BURN         string   `xml:"BURN"`
	PhysicalPage int      `xml:"PhysicalPage"`
}

type Page8 struct {
	Section7     Section7 `xml:"Section7"`
	BURN         string   `xml:"BURN"`
	PhysicalPage int      `xml:"PhysicalPage"`
}

type Page9 struct {
	Section8     string `xml:"Section8"`
	BURN         string `xml:"BURN"`
	PhysicalPage int    `xml:"PhysicalPage"`
}

type Page10 struct {
	Section9     Section9 `xml:"Section9"`
	BURN         string   `xml:"BURN"`
	PhysicalPage int      `xml:"PhysicalPage"`
}

type Page11 struct {
	Section10    Section10 `xml:"Section10"`
	BURN         string    `xml:"BURN"`
	PhysicalPage int       `xml:"PhysicalPage"`
}

type Page12 struct {
	Section11    Section11 `xml:"Section11"`
	BURN         string    `xml:"BURN"`
	PhysicalPage int       `xml:"PhysicalPage"`
}

type Page16 struct {
	Section16    string `xml:"Section16"`
	BURN         string `xml:"BURN"`
	PhysicalPage int    `xml:"PhysicalPage"`
}

type Page17 struct {
	Section12    Section12 `xml:"Section12"`
	BURN         string    `xml:"BURN"`
	PhysicalPage int       `xml:"PhysicalPage"`
}

type Page18 struct {
	Section13    Section13 `xml:"Section13"`
	BURN         string    `xml:"BURN"`
	PhysicalPage int       `xml:"PhysicalPage"`
}

type Page19 struct {
	Section14    Section14 `xml:"Section14"`
	BURN         string    `xml:"BURN"`
	PhysicalPage int       `xml:"PhysicalPage"`
}

type Page20 struct {
	Section15    Section15 `xml:"Section15"`
	BURN         string    `xml:"BURN"`
	PhysicalPage int       `xml:"PhysicalPage"`
}

// Continuation pages and sheets

type ContinuationSheet1 struct {
	Attorney []Attorney  `xml:"Attorney"`
	Donor    Declaration `xml:"Donor"`
}

type ContinuationSheet2 struct {
	AdditionalInformation AdditionalInformation `xml:"AdditionalInformation"`
	Donor                 Declaration           `xml:"Donor"`
}

type ContinuationSheet3 struct {
	Donor     PersonName       `xml:"Donor"`
	Signatory AuthorisedPerson `xml:"Signatory"`
	Witnesses []Witness        `xml:"Witnesses"`
}

type ContinuationSheet4 struct {
	CompanyRegistration string             `xml:"CompanyRegistration"`
	AuthorisedPerson    []AuthorisedPerson `xml:"AuthorisedPerson"`
}

type ContinuationPage1 struct {
	ContinuationSheet1 ContinuationSheet1 `xml:"ContinuationSheet1"`
	BURN               string             `xml:"BURN"`
	PhysicalPage       int                `xml:"PhysicalPage"`
}

type ContinuationPage2 struct {
	ContinuationSheet2 ContinuationSheet2 `xml:"ContinuationSheet2"`
	BURN               string             `xml:"BURN"`
	PhysicalPage       int                `xml:"PhysicalPage"`
}

type ContinuationPage3 struct {
	ContinuationSheet3 ContinuationSheet3 `xml:"ContinuationSheet3"`
	BURN               string             `xml:"BURN"`
	PhysicalPage       int                `xml:"PhysicalPage"`
}

type ContinuationPage4 struct {
	ContinuationSheet4 ContinuationSheet4 `xml:"ContinuationSheet4"`
	BURN               string             `xml:"BURN"`
	PhysicalPage       int                `xml:"PhysicalPage"`
}

type AuthorisedPerson struct {
	FullName string `xml:"FullName"`
	Declaration
}

type InfoPage struct {
	BURN         string `xml:"BURN"`
	PhysicalPage int    `xml:"PhysicalPage"`
}

type AdditionalInformation struct {
	Notes                bool `xml:"Notes"`
	Instructions         bool `xml:"Instructions"`
	Preferences          bool `xml:"Preferences"`
	ReplacementAttorneys bool `xml:"ReplacementAttorneys"`
	Jointly              bool `xml:"Jointly"`
}
