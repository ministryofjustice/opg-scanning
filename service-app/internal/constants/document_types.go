package constants

const (
	DocumentTypeLPA002  = "LPA002"
	DocumentTypeLP0002R = "LPA002R"
	DocumentTypeLPAPA   = "LPA-PA"
	DocumentTypeLPAPW   = "LPA-PW"
	DocumentTypeLPA114  = "LPA114"
	DocumentTypeLPA117  = "LPA117"
	DocumentTypeLPA120  = "LPA120"
	DocumentTypeLP1F    = "LP1F"
	DocumentTypeLP1H    = "LP1H"
	DocumentTypeLP2     = "LP2"
	DocumentTypeLPC     = "LPC"
	DocumentCorresp     = "Correspondence"
)

const (
	DocumentTypeEP2PG = "EP2PG"
	DocumentTypeEPA   = "EPA"
)

const (
	DocumentTypeCOPORD = "COPORD"
)

var (
	SupportedDocumentTypes = []string{
		DocumentTypeLP1F,
		DocumentTypeLP1H,
		DocumentCorresp,
		DocumentTypeLPC,
		DocumentTypeLPA120,
		DocumentTypeLPA002,
		DocumentTypeLPAPA,
		DocumentTypeLPAPW,
		DocumentTypeLPA114,
		DocumentTypeLPA117,
		DocumentTypeEP2PG,
		DocumentTypeLP2,
	}

	LPATypeDocuments = []string{
		DocumentTypeLPA002,
		DocumentTypeLP1F,
		DocumentTypeLP1H,
		DocumentTypeLP2,
	}

	EPATypeDocuments = []string{
		DocumentTypeEP2PG,
		DocumentTypeEPA,
	}

	NewCaseDocuments = []string{
		DocumentTypeEP2PG,
		DocumentTypeLPA002,
		DocumentTypeLP0002R,
		DocumentTypeLP2,
		DocumentTypeLP1F,
		DocumentTypeLP1H,
		DocumentTypeCOPORD,
	}
)
