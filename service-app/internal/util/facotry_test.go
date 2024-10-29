package util

import (
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sampleXML = []byte(`
<LP1F>
    <Page1>
        <Section1>
            <Title>Mr.</Title>
            <FirstName>John</FirstName>
            <LastName>Doe</LastName>
            <OtherNames>Johnny</OtherNames>
            <DOB>1980-01-01</DOB>
            <Address>123 Main St, Springfield, USA</Address>
            <EmailAddress>john.doe@example.com</EmailAddress>
        </Section1>
        <BURN>123456789</BURN>
        <PhysicalPage>1</PhysicalPage>
    </Page1>
    <Page2>
        <Section2>
            <Attorney1>
                <Title>Ms.</Title>
                <FirstName>Jane</FirstName>
                <LastName>Smith</LastName>
                <DOB>1985-06-15</DOB>
                <Address>456 Oak St, Springfield, USA</Address>
                <EmailAddress>jane.smith@example.com</EmailAddress>
                <TrustCorporation>true</TrustCorporation>
            </Attorney1>
            <Attorney2>
                <Title>Mr.</Title>
                <FirstName>Bob</FirstName>
                <LastName>Brown</LastName>
                <DOB>1990-04-20</DOB>
                <Address>789 Pine St, Springfield, USA</Address>
                <EmailAddress>bob.brown@example.com</EmailAddress>
            </Attorney2>
        </Section2>
        <BURN>987654321</BURN>
        <PhysicalPage>2</PhysicalPage>
    </Page2>
    <Page3>
        <Section2>
            <Attorney>
                <Title>Dr.</Title>
                <FirstName>Emily</FirstName>
                <LastName>White</LastName>
                <DOB>1995-05-30</DOB>
                <Address>321 Elm St, Springfield, USA</Address>
                <EmailAddress>emily.white@example.com</EmailAddress>
            </Attorney>
            <Attorney>
                <Title>Mr.</Title>
                <FirstName>Chris</FirstName>
                <LastName>Green</LastName>
                <DOB>1988-03-25</DOB>
                <Address>654 Cedar St, Springfield, USA</Address>
                <EmailAddress>chris.green@example.com</EmailAddress>
            </Attorney>
            <MoreAttorneys>false</MoreAttorneys>
        </Section2>
        <BURN>456789123</BURN>
        <PhysicalPage>3</PhysicalPage>
    </Page3>
    <Page4>
        <Section3>
            <AppointedOneAttorney>true</AppointedOneAttorney>
            <JointlyAndSeverally>false</JointlyAndSeverally>
            <Jointly>true</Jointly>
            <JointlyForSome>false</JointlyForSome>
        </Section3>
        <BURN>789123456</BURN>
        <PhysicalPage>4</PhysicalPage>
    </Page4>
</LP1F>
`)

func TestParseDocument_TestParser(t *testing.T) {
	parsedDoc, err := ProcessDocument(sampleXML, "LP1F", "xml")
	require.NoError(t, err)

	// Assert type and cast
	doc, ok := parsedDoc.(*types.LP1FDocument)
	require.True(t, ok, "expected parsedDoc to be of type *types.LP1FDocument")

	assert.Equal(t, "John", doc.Page1.Section1.FirstName)
	assert.Equal(t, "Jane", doc.Page2.Section2.Attorney1.FirstName)
	assert.Equal(t, "Emily", doc.Page3.Section2.Attorney[0].FirstName)
}
