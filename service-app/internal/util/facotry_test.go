package util

import (
	"encoding/base64"
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Small XML sample to test parsing logic
const sampleXML = `
<LP1F>
    <Page1>
        <Section1>
            <Title>Mr.</Title>
            <FirstName>John</FirstName>
            <LastName>Doe</LastName>
        </Section1>
        <BURN>123456789</BURN>
    </Page1>
</LP1F>
`

func TestProcessDocument(t *testing.T) {
	encodedXML := base64.StdEncoding.EncodeToString([]byte(sampleXML))

	doc := &types.Document{
		Type:        "LP1F",
		Encoding:    "base64",
		NoPages:     1,
		EmbeddedXML: encodedXML,
	}

	parsedDoc, err := ProcessDocument(doc, "LP1F", "xml")
	require.NoError(t, err, "ProcessDocument returned an error")

	lp1fDoc, ok := parsedDoc.(*types.LP1FDocument)
	require.True(t, ok, "expected parsedDoc to be of type *types.LP1FDocument")

	assert.Equal(t, "John", lp1fDoc.Page1.Section1.FirstName, "FirstName mismatch")
	assert.Equal(t, "Doe", lp1fDoc.Page1.Section1.LastName, "LastName mismatch")
	assert.Equal(t, "123456789", lp1fDoc.Page1.BURN, "BURN mismatch")
}
