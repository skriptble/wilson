package compact

type BSONType byte

func (bt BSONType) String() string {
	switch bt {
	case '\x01':
		return "double"
	case '\x02':
		return "string"
	default:
		return "invalid"
	}
}
