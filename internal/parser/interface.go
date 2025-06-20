package parser

type CommonValidator interface {
	Setup(doc any) error
	Validate() []string
}
