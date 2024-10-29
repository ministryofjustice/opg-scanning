package parser

type DocumentParser interface {
	ParseDocument(data []byte) (interface{}, error)
	AddError(err string)
	GetErrors() []string
}
