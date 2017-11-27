package bson

type Array struct {
	*Document
}

func (a *Array) Validate(recursive bool) (uint32, error) {
	return 0, nil
}

func (a *Array) Walk(recursive bool, f func(prefix []string, e *ReaderElement)) {
}

func (a *Array) Lookup(index int) (*ReaderElement, error) {
	return nil, nil
}

// TODO(skriptble): Should this also renumber all the idexes after this so they
// are immediately writeable after this operation or should we delay the update
// of the underlying arrays until we actually need to write?
func (a *Array) Delete(index int) *Array {
	return a
}

func (a *Array) Update(m Modifier, index int) *Array {
	return a
}
