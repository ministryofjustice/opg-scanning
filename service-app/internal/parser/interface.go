package parser

type CommonValidator interface {
	Setup(doc interface{}) error
	Validate() error
	GetValidatorErrorMessages() []string
}

type CommonSanitizer interface {
	Setup(doc interface{}) error
	Sanitize() (interface{}, error)
}
