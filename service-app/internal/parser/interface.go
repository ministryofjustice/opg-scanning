package parser

type CommonValidator interface {
	Setup(doc any) error
	Validate() error
	GetValidatorErrorMessages() []string
}

type CommonSanitizer interface {
	Setup(doc any) error
	Sanitize() (any, error)
}
