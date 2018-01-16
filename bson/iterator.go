package bson

type Iterator struct {
	d     *Document
	index int
	elem  *Element
	err   error
}

func newIterator(d *Document) *Iterator {
	return &Iterator{d: d}
}

func (itr *Iterator) Next() bool {
	if itr.index >= len(itr.d.elems) {
		return false
	}

	e := itr.d.elems[itr.index]

	_, err := e.Validate()
	if err != nil {
		itr.err = err
		return false
	}

	itr.elem = e
	itr.index++

	return true
}

func (itr *Iterator) Element() *Element {
	return itr.elem
}

func (itr *Iterator) Err() error {
	return itr.err
}
