package bson

type Binary struct {
	start uint
	data  []byte
}

func (b *Binary) Subtype() byte {
	return '\x00'
}

func (b *Binary) Generic() []byte {
	return nil
}

func (b *Binary) Function() []byte {
	return nil
}

func (b *Binary) OldGeneric() []byte {
	return nil
}

// TODO(skriptble): Should this return a uuid.UUID?
func (b *Binary) OldUUID() []byte {
	return nil
}

func (b *Binary) UUID() []byte {
	return nil
}

func (b *Binary) MD5() []byte {
	return nil
}

func (b *Binary) UserDefined() []byte {
	return nil
}
