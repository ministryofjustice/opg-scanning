package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var Set = `<?xml version="1.0" encoding="UTF-8"?><Set>
  <Header CaseNo="123" Scanner="scanner1" ScanTime="2025-03-06T12:34:56" ScannerOperator="JohnDoe" Schedule="A" FeeNumber="456"/>
  <Body>
    <Document Type="example" Encoding="utf-8" NoPages="1">
      <XML>Some embedded XML here</XML>
      <PDF>Some embedded PDF here</PDF>
    </Document>
  </Body>
  <Test>unexpected element</Test>
</Set>
`

func Test_BaseParserXml(t *testing.T) {
	// Convert Set to Bytes
	set := []byte(Set)
	_, err := BaseParserXml(set)
	// We expect there to be an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected XML elements found in field")
}
