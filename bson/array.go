package bson

// Array represents an array in BSON. The methods of this type are more
// expensive than those on Document because they require potentially updateing
// multiple keys to ensure the array stays valid at all times.
type Array struct {
	*Document
}

func (a *Array) Validate(recursive bool) (uint32, error) {
	return 0, nil
}

func (a *Array) Lookup(index int) (*Element, error) {
	return nil, nil
}

func (a *Array) ElementAt(index uint) (*Element, error) {
	return nil, nil
}

func (a *Array) Append(elems ...*Element) *Array {
	return nil
}

func (a *Array) Prepend(elems ...*Element) *Array {
	return nil
}

func (a *Array) Replace(elems ...*Element) *Array {
	return nil
}

func (a *Array) Combine(doc interface{}) error {
	return nil
}

// TODO(skriptble): Should this also renumber all the idexes after this so they
// are immediately writeable after this operation or should we delay the update
// of the underlying arrays until we actually need to write?
func (a *Array) Delete(index int) *Array {
	return a
}
