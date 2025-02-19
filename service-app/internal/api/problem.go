package api

type Problem struct {
	Title            string
	ValidationErrors []string
}

func (p Problem) Error() string {
	return p.Title
}
