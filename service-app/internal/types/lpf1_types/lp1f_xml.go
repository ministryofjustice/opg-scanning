package lpf1_types

type BasePage struct {
	BURN         string `xml:"BURN"`
	PhysicalPage string `xml:"PhysicalPage"`
}

type Address struct {
	Address1 string `xml:"Address1"`
	Address2 string `xml:"Address2,omitempty"`
	Address3 string `xml:"Address3,omitempty"`
	Postcode string `xml:"Postcode"`
}

type Salutation struct {
	Mr        bool   `xml:"Mr"`
	Mrs       bool   `xml:"Mrs"`
	Ms        bool   `xml:"Ms"`
	Miss      bool   `xml:"Miss"`
	Other     bool   `xml:"Other"`
	OtherName string `xml:"OtherName,omitempty"`
}

type Signatory struct {
	FullName string `xml:"FullName"`
	Declaration
}

type PersonName struct {
	Salutation     Salutation `xml:"Salutation"`
	LastName       string     `xml:"LastName"`
	Forename       string     `xml:"Forename"`
	OtherForenames string     `xml:"OtherForenames,omitempty"`
}

type Declaration struct {
	Signature bool   `xml:"Signature,omitempty"`
	Date      string `xml:"Date,omitempty"`
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
	DOB              string  `xml:"DOB,omitempty"`
	Address          Address `xml:"Address,omitempty"`
	EmailAddress     string  `xml:"EmailAddress,omitempty"`
	TrustCorporation *bool   `xml:"TrustCorporation,omitempty"`
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
	SkillsAndExpertise         string `xml:"SkillsAndExpertise,omitempty"`
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
	OtherNames   string  `xml:"OtherNames,omitempty"`
	DOB          string  `xml:"DOB"`
	Address      Address `xml:"Address"`
	EmailAddress string  `xml:"EmailAddress"`
}

type Section2 struct {
	Attorney1 Attorney `xml:"Attorney1"`
	Attorney2 Attorney `xml:"Attorney2"`
}

type Section3 struct {
	AppointedOneAttorney bool `xml:"AppointedOneAttorney"`
	JointlyAndSeverally  bool `xml:"JointlyAndSeverally"`
	Jointly              bool `xml:"Jointly"`
	JointlyForSome       bool `xml:"JointlyForSome"`
}

type Section4 struct {
	Attorney1             Attorney `xml:"Attorney1"`
	Attorney2             Attorney `xml:"Attorney2,omitempty"`
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
	OnlineLPA             bool   `xml:"OnlineLPA,omitempty"`
	OnlineLPAID           string `xml:"OnlineLPAID,omitempty"`
}

type Section15 struct {
	Applicant []Declaration `xml:"Applicant"`
}

type Page1 struct {
	BasePage
	Section1 Section1 `xml:"Section1"`
}

type Page2 struct {
	BasePage
	Section2 Section2 `xml:"Section2"`
}

type Page3 struct {
	BasePage
	Section2 Section3 `xml:"Section3"`
}

type Page4 struct {
	BasePage
	Section3 Section3 `xml:"Section3"`
}

type Page5 struct {
	BasePage
	Section4 Section4 `xml:"Section4"`
}

type Page6 struct {
	BasePage
	Section5 Section5 `xml:"Section5"`
}

type Page7 struct {
	BasePage
	Section6 Section6 `xml:"Section6"`
}

type Page8 struct {
	BasePage
	Section7 Section7 `xml:"Section7"`
}

type Page9 struct {
	BasePage
	Section8 string `xml:"Section8"`
}

type Page10 struct {
	BasePage
	Section9 Section9 `xml:"Section9"`
}

type Page11 struct {
	BasePage
	Section10 Section10 `xml:"Section10"`
}

type Page12 struct {
	BasePage
	Section11 Section11 `xml:"Section11"`
}

type Page16 struct {
	BasePage
	Section16 string `xml:"Section16"`
}

type Page17 struct {
	BasePage
	Section12 Section12 `xml:"Section12"`
}

type Page18 struct {
	BasePage
	Section13 Section13 `xml:"Section13"`
}

type Page19 struct {
	BasePage
	Section14 Section14 `xml:"Section14"`
}

type Page20 struct {
	BasePage
	Section15 Section15 `xml:"Section15"`
}

type ContinuationPage1 struct {
	BasePage
	ContinuationSheet1 ContinuationSheet1 `xml:"ContinuationSheet1"`
}

type ContinuationPage2 struct {
	BasePage
	ContinuationSheet2 ContinuationSheet2 `xml:"ContinuationSheet2"`
}

type ContinuationPage3 struct {
	BasePage
	ContinuationSheet3 ContinuationSheet3 `xml:"ContinuationSheet3"`
}

type ContinuationPage4 struct {
	BasePage
	ContinuationSheet4 ContinuationSheet4 `xml:"ContinuationSheet4"`
}

type ContinuationSheet1 struct {
	Attorney []Attorney  `xml:"Attorney"`
	Donor    Declaration `xml:"Donor"`
}

type ContinuationSheet2 struct {
	AdditionalInformation AdditionalInformation `xml:"AdditionalInformation"`
	Donor                 Declaration           `xml:"Donor"`
}

type ContinuationSheet3 struct {
	Donor     PersonName `xml:"Donor"`
	Signatory Signatory  `xml:"Signatory"`
	Witnesses []Witness  `xml:"Witnesses"`
}

type ContinuationSheet4 struct {
	CompanyRegistration string             `xml:"CompanyRegistration"`
	AuthorisedPerson    []AuthorisedPerson `xml:"AuthorisedPerson"`
}

type AuthorisedPerson struct {
	FullName string `xml:"FullName"`
	Declaration
}

type InfoPage struct {
	BasePage
}

type AdditionalInformation struct {
	Notes                bool `xml:"Notes"`
	Instructions         bool `xml:"Instructions"`
	Preferences          bool `xml:"Preferences"`
	ReplacementAttorneys bool `xml:"ReplacementAttorneys"`
	Jointly              bool `xml:"Jointly"`
}
