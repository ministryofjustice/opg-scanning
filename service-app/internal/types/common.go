package types

type Parser interface {
	Parse(data []byte) (interface{}, error)
}

type Validator interface {
	Validate() error
}

type Sanitizer interface {
	Sanitize(doc interface{}) (interface{}, error)
}
