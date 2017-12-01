package extjson

import (
	"fmt"

	"errors"

	"github.com/skriptble/wilson/bson/internal/jsonparser"
	"github.com/skriptble/wilson/builder"
)

type parseState struct {
	wtype         wrapperType
	firstKey      bool
	currentValue  []byte
	subDocBuilder *builder.DocumentBuilder
	docBuilder    *builder.DocumentBuilder
	code          *string
	scope         *builder.DocumentBuilder
}

func newParseState(b *builder.DocumentBuilder) *parseState {
	var subdoc builder.DocumentBuilder
	b.Init()

	return &parseState{firstKey: true, subDocBuilder: &subdoc, docBuilder: b}
}

func (s *parseState) parseElement(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
	wtype := wrapperKeyType(key)

	if s.wtype != wtype && !s.firstKey {
		return fmt.Errorf(
			"previous key in the object were %s, but the %s is %s",
			s.wtype.String(),
			string(key),
			wtype.String(),
		)
	}

	// The only wrapper type that allows more than one top-level key is $code/$scope
	if s.wtype != None && s.wtype != Code && !s.firstKey {
		return errors.New("%s wrapper object cannot have more than one key")
	}

	s.firstKey = true

	if s.wtype == None {
		return parseDocElement(s.subDocBuilder)(key, value, dataType, offset)
	}

	k := string(key)

	switch s.wtype {
	case ObjectId:
		oid, err := parseObjectId(value, dataType)
		if err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.ObjectId(k, oid))
	case Symbol:
		str, err := parseSymbol(value, dataType)
		if err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.Symbol(k, str))
	case Int32:
		i, err := parseInt32(value, dataType)
		if err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.Int32(k, i))
	case Int64:
		i, err := parseInt64(value, dataType)
		if err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.Int64(k, i))
	case Double:
		f, err := parseDouble(value, dataType)
		if err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.Double(k, f))
	case Decimal:
		d, err := parseDecimal(value, dataType)
		if err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.Decimal(k, d))
	case Binary:
		b, t, err := parseBinary(value, dataType)
		if err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.BinaryWithSubtype(k, b, t))
	case Code:
		switch string(key) {
		case "$code":
			code, err := parseCode(value, dataType)
			if err != nil {
				return err
			}

			s.code = &code
		case "$scope":
			b, err := parseScope(value, dataType)
			if err != nil {
				return err
			}

			s.scope = b
		}
	case Timestamp:
		t, i, err := parseTimestamp(value, dataType)
		if err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.Timestamp(k, t, i))
	case Regex:
		p, o, err := parseRegex(value, dataType)
		if err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.Regex(k, p, o))
	case DBPointer:
		ns, oid, err := parseDBPointer(value, dataType)
		if err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.DBPointer(k, ns, oid))
	case DateTime:
		d, err := parseDatetime(value, dataType)
		if err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.DateTime(k, d))
	case MinKey:
		if err := parseMinKey(value, dataType); err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.MinKey(k))
	case MaxKey:
		if err := parseMaxKey(value, dataType); err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.MaxKey(k))
	case Undefined:
		if err := parseUndefined(value, dataType); err != nil {
			return err
		}

		s.docBuilder.Append(builder.C.Undefined(k))
	}

	return nil
}
