package parser

type CommonValidator interface {
	Setup(doc any) error
	Validate() []string
}

type CommonSanitizer interface {
	Setup(doc any) error
}
