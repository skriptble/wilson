package bson

// node is a compact representation of an element within a BSON document.
// The first 4 bytes are where the element starts in an underlying []byte. The
// last 4 bytes are where the value for that element begins.
//
// The type of the element can be accessed as `data[n[0]]`. The key of the
// element can be accessed as `data[n[0]+1:n[1]-1]`. This will account for the
// null byte at the end of the c-style string. The value can be accessed as
// `data[n[1]:]`. Since there is no value end byte, an unvalidated document
// could result in parsing errors.
type node [2]uint32

const (
	TypeDouble           BSONType = 0x01
	TypeString           BSONType = 0x02
	TypeEmbeddedDocument BSONType = 0x03
	TypeArray            BSONType = 0x04
	TypeBinary           BSONType = 0x05
	TypeUndefined        BSONType = 0x06
	TypeObjectID         BSONType = 0x07
	TypeBoolean          BSONType = 0x08
	TypeDateTime         BSONType = 0x09
	TypeNull             BSONType = 0x0A
	TypeRegex            BSONType = 0x0B
	TypeDBPointer        BSONType = 0x0C
	TypeJavaScript       BSONType = 0x0D
	TypeSymbol           BSONType = 0x0E
	TypeCodeWithScope    BSONType = 0x0F
	TypeInt32            BSONType = 0x10
	TypeTimestamp        BSONType = 0x11
	TypeInt64            BSONType = 0x12
	TypeDecimal128       BSONType = 0x13
	TypeMinKey           BSONType = 0xFF
	TypeMaxKey           BSONType = 0x7F
)

type BSONType byte

func (bt BSONType) String() string {
	switch bt {
	case '\x01':
		return "double"
	case '\x02':
		return "string"
	case '\x03':
		return "embedded document"
	case '\x04':
		return "array"
	case '\x05':
		return "binary"
	case '\x06':
		return "undefined"
	case '\x07':
		return "ObjectId"
	case '\x08':
		return "boolean"
	case '\x09':
		return "UTC datetime"
	case '\x0A':
		return "null"
	case '\x0B':
		return "regex"
	case '\x0C':
		return "DBPointer"
	case '\x0D':
		return "javascript"
	case '\x0E':
		return "symbol"
	case '\x0F':
		return "code with scope"
	case '\x10':
		return "32-bit integer"
	case '\x11':
		return "timestamp"
	case '\x12':
		return "64-bit integer"
	case '\x13':
		return "128-bit decimal"
	case '\xFF':
		return "min key"
	case '\x7F':
		return "max key"
	default:
		return "invalid"
	}
}
