package constants

const (
	DocumentTypeLPA002 = "LPA002"
	DocumentTypeLP1F   = "LP1F"
	DocumentTypeLP1H   = "LP1H"
	DocumentTypeLP2    = "LP2"
)

const (
	DocumentTypeEP2PG = "EP2PG"
	DocumentTypeEPA   = "EPA"
)

const (
	DocumentTypeCOPORD = "COPORD"
)

var (
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

	CourtOrderDocuments = []string{
		DocumentTypeCOPORD,
	}

	StandaloneInstruments = []string{
		DocumentTypeLP1F,
		DocumentTypeLP1H,
	}

	ExemptApplications = []string{
		"LPA002R",
	}
)
