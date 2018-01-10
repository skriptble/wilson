package bson

type ReaderIterator struct {
	r    Reader
	pos  uint32
	end  uint32
	elem *Element
	err  error
}

func NewReadIterator(r Reader) (*ReaderIterator, error) {
	itr := new(ReaderIterator)
	if len(r) < 5 {
		return nil, ErrTooSmall
	}
	givenLength := readi32(r[0:4])
	if len(r) < int(givenLength) {
		return nil, ErrInvalidLength
	}

	itr.r = r
	itr.pos = 4
	itr.end = uint32(givenLength)
	itr.elem = &Element{}

	return itr, nil
}

func (itr *ReaderIterator) Next() bool {
	if itr.pos >= itr.end {
		itr.err = ErrInvalidReadOnlyDocument
		return false
	}
	if itr.r[itr.pos] == '\x00' {
		return false
	}
	elemStart := itr.pos
	itr.pos++
	n, err := itr.r.validateKey(itr.pos, itr.end)
	itr.pos += n
	if err != nil {
		itr.err = err
		return false
	}

	itr.elem.value.start = elemStart
	itr.elem.value.offset = itr.pos
	itr.elem.value.data = itr.r

	n, err = itr.elem.value.validate(false)
	itr.pos += n
	if err != nil {
		itr.err = err
		return false
	}
	return true
}

func (itr *ReaderIterator) Element() *Element {
	return itr.elem
}

func (itr *ReaderIterator) Err() error {
	return itr.err
}
