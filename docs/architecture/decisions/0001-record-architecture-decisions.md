# Avoid Go Struct-Based Re-Encoding for XML with Complex XSDs to support Sanitization

## Status

Accepted

## Context

We explored using Go structs with the `encoding/xml` package to sanitize and re-encode XML documents that must conform to a complex XSD. The XML schema enforces strict requirements, including:

- Specific element ordering
- Mandatory presence of all elements (even if empty)
- Slight variations in expected field names/types across similar sections

Initial attempts to encode XML from Go structs revealed several challenges:

- Go's XML encoder outputs elements in the order fields are defined in the struct, making ordering difficult to enforce
- Use of `omitempty` caused required elements to be omitted, breaking validation
- Minor schema differences required duplication of nearly identical types
- Any mismatch during re-encoding could lead to validation failures

## Decision

We will **not** proceed with struct-based XML re-encoding in Go.

Instead, sanitized XML will be handled and re-encoded using a more XML-native language/tool (e.g., PHP), which provides better support for:

- Schema validation
- Element ordering
- Handling of required empty elements
- Manipulation of XML trees without rigid struct mapping

A fallback mechanism will remain in place: if sanitized XML cannot be re-encoded successfully, the original XML will be passed through the pipeline.

## Rationale

Using Go for strict XML/XSD compliance is brittle and introduces unnecessary complexity. Languages like PHP (Sirius API) offer more flexibility and better tooling for dynamic XML transformation.

This approach:

- Reduces fragility caused by strict struct ordering and mapping
- Avoids duplication of types for minor schema differences
- Improves long-term maintainability
- Enables more robust schema validation workflows

## Consequences

- Go code remains simpler, avoiding complex struct hierarchies
- XML handling remains in a more suitable language
- Decision is documented for future reference to avoid re-evaluating this trade-off unnecessarily
