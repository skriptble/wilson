package bson

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"time"

	"bytes"

	"github.com/skriptble/wilson/bson/decimal"
	"github.com/skriptble/wilson/bson/objectid"
)

var tBinary = reflect.TypeOf(Binary{})
var tBool = reflect.TypeOf(false)
var tCodeWithScope = reflect.TypeOf(CodeWithScope{})
var tDBPointer = reflect.TypeOf(DBPointer{})
var tDecimal = reflect.TypeOf(decimal.Decimal128{})
var tDocument = reflect.TypeOf((*Document)(nil))
var tFloat32 = reflect.TypeOf(float32(0))
var tFloat64 = reflect.TypeOf(float64(0))
var tInt = reflect.TypeOf(int(0))
var tInt32 = reflect.TypeOf(int32(0))
var tInt64 = reflect.TypeOf(int64(0))
var tJavaScriptCode = reflect.TypeOf(JavaScriptCode(""))
var tOID = reflect.TypeOf(objectid.ObjectID{})
var tReader = reflect.TypeOf(Reader(nil))
var tRegex = reflect.TypeOf(Regex{})
var tString = reflect.TypeOf("")
var tSymbol = reflect.TypeOf(Symbol(""))
var tTime = reflect.TypeOf(time.Time{})
var tTimestamp = reflect.TypeOf(Timestamp{})
var tUint = reflect.TypeOf(uint(0))
var tUint32 = reflect.TypeOf(uint32(0))
var tUint64 = reflect.TypeOf(uint64(0))

var tEmpty = reflect.TypeOf((*interface{})(nil)).Elem()

type Unmarshaler interface {
	UnmarshalBSON([]byte) error
}

type DocumentUnmarshaler interface {
	UnmarshalBSONDocument(*Document) error
}

type Decoder struct {
	pReader    *peekLengthReader
	bsonReader Reader
}

type peekLengthReader struct {
	io.Reader
	length [4]byte
	pos    int32
}

func newPeekLengthReader(r io.Reader) *peekLengthReader {
	return &peekLengthReader{Reader: r, pos: -1}
}

func (r *peekLengthReader) peekLength() (int32, error) {
	_, err := io.ReadFull(r, r.length[:])
	if err != nil {
		return 0, err
	}

	// Mark that the length has been read.
	r.pos = 0

	return readi32(r.length[:]), nil
}

func (r *peekLengthReader) Read(b []byte) (int, error) {
	// If either peekLength hasn't been called or the length has been read past, read from the
	// io.Reader.
	if r.pos < 0 || r.pos > 3 {
		return r.Reader.Read(b)
	}

	// Read as much of the length as possible into the buffer
	bytesToRead := 4 - r.pos
	if len(b) < int(bytesToRead) {
		bytesToRead = int32(len(b))
	}

	r.pos += int32(copy(b, r.length[r.pos:r.pos+bytesToRead]))

	// Because we use io.ReadFull everywhere, we don't need to read any further since it will be
	// read in a subsequent call to Read.
	return int(bytesToRead), nil
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{pReader: newPeekLengthReader(r)}
}

func (d *Decoder) Decode(v interface{}) error {
	switch t := v.(type) {
	case Unmarshaler:
		err := d.decodeToReader()
		if err != nil {
			return err
		}

		return t.UnmarshalBSON(d.bsonReader)
	case io.Writer:
		err := d.decodeToReader()
		if err != nil {
			return err
		}

		_, err = t.Write(d.bsonReader)
		return err
	case []byte:
		length, err := d.pReader.peekLength()
		if err != nil {
			return err
		}

		if len(t) < int(length) {
			return ErrTooSmall
		}

		_, err = io.ReadFull(d.pReader, t)
		if err != nil {
			return err
		}

		_, err = Reader(t).Validate()
		return err
	default:
		rval := reflect.ValueOf(v)
		return d.reflectDecode(rval)
	}
}

func (d *Decoder) decodeToReader() error {
	var err error
	d.bsonReader, err = NewFromIOReader(d.pReader)
	if err != nil {
		return err
	}

	_, err = d.bsonReader.Validate()
	return err

}

func (d *Decoder) reflectDecode(val reflect.Value) error {
	switch val.Kind() {
	case reflect.Map:
		return d.decodeIntoMap(val)
	case reflect.Slice, reflect.Array:
		return d.decodeIntoSlice(val)
	case reflect.Struct:
		return d.decodeIntoStruct(val)
	case reflect.Ptr:
		v := val.Elem()

		if v.Kind() == reflect.Struct {
			return d.decodeIntoStruct(v)
		}

		fallthrough
	default:
		return fmt.Errorf("cannot decode BSON document to type %s", val.Type())
	}
}

func (d *Decoder) createEmptyValue(r Reader, t reflect.Type) (reflect.Value, error) {
	var val reflect.Value

	switch t.Kind() {
	case reflect.Map:
		val = reflect.MakeMap(t)
	case reflect.Ptr:
		if t == tDocument {
			val = reflect.ValueOf(NewDocument(0))
			break
		}

		empty, err := d.createEmptyValue(r, t.Elem())
		if err != nil {
			return val, err
		}

		val = reflect.New(empty.Type())
		val.Elem().Set(empty)
	case reflect.Slice:
		length := 0
		_, err := r.readElements(func(_ *Element) error {
			length++
			return nil
		})

		if err != nil {
			return val, err
		}

		val = reflect.MakeSlice(t.Elem(), length, length)
	case reflect.Struct:
		val = reflect.New(t)
	default:
		if t == tReader {
			val = reflect.ValueOf(r)
			break
		}

		val = reflect.Zero(t)
	}

	return val, nil
}

func (d *Decoder) getReflectValue(v *Value, containerType reflect.Type, outer reflect.Type) (reflect.Value, error) {
	var val reflect.Value

	for containerType.Kind() == reflect.Ptr {
		containerType = containerType.Elem()
	}

	switch v.Type() {
	case 0x1:
		f := v.Double()

		switch containerType {
		case tInt32:
			if math.Floor(f) == f && f <= float64(math.MaxInt32) {
				val = reflect.ValueOf(int32(f))
			}
		case tInt64:
			if math.Floor(f) == f && f <= float64(math.MaxInt64) {
				val = reflect.ValueOf(int64(f))
			}
		case tInt:
			if math.Floor(f) != f || f > float64(math.MaxInt64) {
				break
			}

			i := int64(f)
			if int64(int(i)) == i {
				val = reflect.ValueOf(int(i))
			}

		case tFloat32:
			if f > math.MaxFloat32 {
				return val, nil
			}

			fallthrough
		case tFloat64, tEmpty:
			val = reflect.ValueOf(f)
		default:
			return val, nil
		}

	case 0x2:
		if containerType != tString && containerType != tEmpty {
			return val, nil
		}

		val = reflect.ValueOf(v.StringValue())
	case 0x4:
		if containerType.Kind() == reflect.Slice || containerType.Kind() == reflect.Array {
			d := NewDecoder(bytes.NewBuffer(v.ReaderDocument()))
			err := d.decodeIntoSlice(val)
			if err != nil {
				return val, err
			}

			break
		}

		fallthrough

	case 0x3:
		r := v.ReaderDocument()

		typeToCreate := containerType
		if typeToCreate == tEmpty {
			typeToCreate = outer
		}

		empty, err := d.createEmptyValue(r, typeToCreate)
		if err != nil {
			return val, err
		}

		d := NewDecoder(bytes.NewBuffer(r))
		err = d.Decode(empty.Interface())
		if err != nil {
			return val, err
		}

		if reflect.PtrTo(typeToCreate) == empty.Type() {
			empty = empty.Elem()
		}

		val = empty

	case 0x5:
		if containerType != tBinary && containerType != tEmpty {
			return val, nil
		}

		st, data := v.Binary()
		val = reflect.ValueOf(Binary{Subtype: st, Data: data})
	case 0x6:
		if containerType != tEmpty {
			return val, nil
		}

		val = reflect.ValueOf(Undefined)
	case 0x7:
		if containerType != tOID && containerType != tEmpty {
			return val, nil
		}

		val = reflect.ValueOf(v.ObjectID())
	case 0x8:
		if containerType != tBool && containerType != tEmpty {
			return val, nil
		}

		val = reflect.ValueOf(v.Boolean())
	case 0x9:
		if containerType != tTime && containerType != tEmpty {
			return val, nil
		}

		val = reflect.ValueOf(v.DateTime())
	case 0xA:
		if containerType != tEmpty {
			return val, nil
		}

		val = reflect.ValueOf(Null)
	case 0xB:
		if containerType != tRegex && containerType != tEmpty {
			return val, nil
		}

		p, o := v.Regex()
		val = reflect.ValueOf(Regex{Pattern: p, Options: o})
	case 0xC:
		if containerType != tDBPointer && containerType != tEmpty {
			return val, nil
		}

		db, p := v.DBPointer()
		val = reflect.ValueOf(DBPointer{DB: db, Pointer: p})
	case 0xD:
		if containerType != tJavaScriptCode && containerType != tString && containerType != tEmpty {
			return val, nil
		}

		val = reflect.ValueOf(v.Javascript())
	case 0xE:
		if containerType != tSymbol && containerType != tString && containerType != tEmpty {
			return val, nil
		}

		val = reflect.ValueOf(v.Symbol())
	case 0xF:
		if containerType != tCodeWithScope && containerType != tEmpty {
			return val, nil
		}

		code, scope := v.MutableJavascriptWithScope()
		val = reflect.ValueOf(CodeWithScope{Code: code, Scope: scope})
	case 0x10:
		i := v.Int32()

		switch containerType {
		case tUint32, tUint64, tUint:
			if i < 0 {
				return val, nil
			}

			fallthrough
		case tInt32, tInt64, tInt, tFloat32, tFloat64, tEmpty:
			val = reflect.ValueOf(i)
		default:
			return val, nil
		}

	case 0x11:
		if containerType != tTimestamp && containerType != tEmpty {
			return val, nil
		}

		t, i := v.Timestamp()
		val = reflect.ValueOf(Timestamp{T: t, I: i})
	case 0x12:
		i := v.Int64()

		switch containerType {
		case tInt:
			// Check the value can fit in an int
			if int64(int(i)) != i {
				return val, nil
			}

		case tUint:
			if i < 0 || int64(uint(i)) != i {
				return val, nil
			}

		case tUint64:
			if i < 0 {
				return val, nil
			}

			val = reflect.ValueOf(i)
		case tInt64, tFloat32, tFloat64, tEmpty:
		default:
			return val, nil
		}

		val = reflect.ValueOf(i)
	case 0x13:
		if containerType != tDecimal && containerType != tEmpty {
			return val, nil
		}

		val = reflect.ValueOf(v.Decimal128())
	case 0xFF:
		if containerType != tEmpty {
			return val, nil
		}

		val = reflect.ValueOf(MinKey)
	case 0x7f:
		if containerType != tEmpty {
			return val, nil
		}

		val = reflect.ValueOf(MaxKey)
	default:
		return val, fmt.Errorf("invalid BSON type: %s", v.Type())
	}

	return val, nil
}

func (d *Decoder) decodeIntoMap(mapVal reflect.Value) error {
	err := d.decodeToReader()
	if err != nil {
		return err
	}

	itr, err := d.bsonReader.Iterator()
	if err != nil {
		return err
	}

	valType := mapVal.Type().Elem()

	for itr.Next() {
		elem := itr.Element()

		v, err := d.getReflectValue(elem.value, valType, mapVal.Type())
		if err != nil {
			return err
		}

		k := reflect.ValueOf(elem.Key())
		mapVal.SetMapIndex(k, v)
	}

	return itr.Err()
}

func (d *Decoder) decodeIntoSlice(sliceVal reflect.Value) error {
	sliceLength := sliceVal.Len()

	err := d.decodeToReader()
	if err != nil {
		return err
	}

	itr, err := d.bsonReader.Iterator()
	if err != nil {
		return err
	}

	i := 0
	for itr.Next() {
		if i >= sliceLength {
			return ErrTooSmall
		}

		elem := reflect.ValueOf(itr.Element().Clone())
		sliceVal.Index(i).Set(elem)
		i++
	}

	return itr.Err()
}

func matchesField(key string, field string, sType reflect.Type) bool {
	sField, found := sType.FieldByName(field)
	if !found {
		return false
	}

	tag, ok := sField.Tag.Lookup("bson")
	if !ok {
		if len(sField.Tag) == 0 || strings.ContainsRune(tag, ':') {
			return strings.ToLower(key) == strings.ToLower(field)
		}

		tag = string(sField.Tag)
	}

	var fieldKey string
	i := strings.IndexRune(tag, ',')
	if i == -1 {
		fieldKey = tag
	} else {
		fieldKey = tag[:i]
	}

	return fieldKey == key
}

func (d *Decoder) decodeIntoStruct(structVal reflect.Value) error {
	err := d.decodeToReader()
	if err != nil {
		return err
	}

	itr, err := d.bsonReader.Iterator()
	if err != nil {
		return err
	}

	var notFound reflect.Value
	sType := structVal.Type()

	for itr.Next() {
		elem := itr.Element()

		field := structVal.FieldByNameFunc(func(field string) bool {
			return matchesField(elem.Key(), field, sType)
		})
		if field == notFound {
			continue
		}

		v, err := d.getReflectValue(elem.value, field.Type(), structVal.Type())
		if err != nil {
			return err
		}

		field.Set(v)
	}

	return itr.Err()
}
