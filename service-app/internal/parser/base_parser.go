package parser

type BaseParser struct {
    Errors []string
}

func (bp *BaseParser) AddError(err string) {
    bp.Errors = append(bp.Errors, err)
}
