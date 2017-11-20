package bson

type CodeWithScope struct {
	start uint32 // includes the int32 length
	data  []byte
}

func (cws *CodeWithScope) Javascript() string {
	return ""
}

func (cws *CodeWithScope) Scope() *Document {
	return nil
}
