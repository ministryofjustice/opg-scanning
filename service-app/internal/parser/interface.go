package parser

type CommonValidator interface {
	Setup(doc interface{}) error
	Validate() error
}

type CommonSanitizer interface {
	Setup(doc interface{}) error
	Sanitize() (interface{}, error)
}
