package bson

type Iterator struct {
}

func NewIterator(d *Document) (*Iterator, error) {
	return nil, nil
}

func (itr *Iterator) Next() bool {
	return false
}

func (itr *Iterator) Close() error {
	return nil
}

func (itr *Iterator) Element() *Element {
	return nil
}

func (itr *Iterator) Err() error {
	return nil
}
