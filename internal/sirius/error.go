package sirius

import "fmt"

type Error struct {
	StatusCode       int
	ValidationErrors map[string]map[string]string
}

func (e Error) Error() string {
	return fmt.Sprintf("received %d response from Sirius", e.StatusCode)
}
