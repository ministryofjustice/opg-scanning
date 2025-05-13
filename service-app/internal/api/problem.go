package api

type problem struct {
	Title            string
	ValidationErrors []string
}

func (p problem) Error() string {
	return p.Title
}
