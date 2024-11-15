package lpf1_types

// Provides a unified way to access attorney data.
func (doc *LP1FDocument) GetAttorneys() []Attorney {
	attorneys := []Attorney{}
	if doc.Page2.Section2.Attorney1 != (Attorney{}) {
		attorneys = append(attorneys, doc.Page2.Section2.Attorney1)
	}
	if doc.Page2.Section2.Attorney2 != (Attorney{}) {
		attorneys = append(attorneys, doc.Page2.Section2.Attorney2)
	}
	return attorneys
}
