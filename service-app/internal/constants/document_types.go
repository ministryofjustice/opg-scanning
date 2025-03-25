package constants

const (
	DocumentTypeLPA002  = "LPA002"
	DocumentTypeLP0002R = "LPA002R"
	DocumentTypeLPAPA   = "LPA-PA"
	DocumentTypeLPAPW   = "LPA-PW"
	DocumentTypeLPA114  = "LPA114"
	DocumentTypeLPA115  = "LPA115"
	DocumentTypeLPA116  = "LPA116"
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
	DocumentTypeCOPORD     = "COPORD"
	DocumentTypeDEPREPORTS = "DEPREPORTS"
	DocumentTypeDEPCORRES  = "DEPCORRES"
	DocumentTypeFINDOCS    = "FINDOCS"
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
		DocumentTypeLPA115,
		DocumentTypeLPA116,
		DocumentTypeLPA117,
		DocumentTypeEP2PG,
		DocumentTypeLP2,
		DocumentTypeEPA,
		DocumentTypeDEPREPORTS,
		DocumentTypeDEPCORRES,
		DocumentTypeFINDOCS,
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

	NewCaseNumberDocuments = []string{
		DocumentTypeEP2PG,
		DocumentTypeLP2,
		DocumentTypeLP1F,
		DocumentTypeLP1H,
	}

	// these documents should be sent to Sirius to be extracted
	SiriusExtractionDocuments = []string{
		DocumentTypeEP2PG,
		DocumentTypeLP1F,
		DocumentTypeLP1H,
		DocumentTypeLP2,
		DocumentTypeLPC,
	}
)
