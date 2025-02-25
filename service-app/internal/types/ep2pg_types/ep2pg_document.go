package ep2pg_types

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type EP2PGDocument struct {
	XMLName   xml.Name `xml:"EP2PG"`
	Page1     Page     `xml:"Page1"`
}

type Page struct {
	types.BasePage
}