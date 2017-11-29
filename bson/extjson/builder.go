package extjson

import (
	"fmt"

	"strconv"

	"github.com/skriptble/wilson/bson/internal/jsonparser"
	"github.com/skriptble/wilson/builder"
)

type docElementParser func([]byte, []byte, jsonparser.ValueType, int) error
type arrayElementParser func(int, []byte, jsonparser.ValueType, int, error)

func ParseObjectToBuilder(s string) (*builder.DocumentBuilder, error) {
	var b builder.DocumentBuilder
	b.Init()

	return &b, jsonparser.ObjectEach([]byte(s), parseDocElement(&b))
}

func ParseArrayToBuilder(s string) (*builder.ArrayBuilder, error) {
	var b builder.ArrayBuilder
	b.Init()

	_, err := jsonparser.ArrayEach([]byte(s), parseArrayElement(&b))

	return &b, err
}

func parseDocElement(b *builder.DocumentBuilder) docElementParser {
	return func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		name := string(key)

		switch dataType {
		case jsonparser.String:
			b.Append(builder.C.String(name, string(value)))

		case jsonparser.Number:
			i, err := jsonparser.ParseInt(value)
			if err == nil {
				b.Append(builder.C.Int64(name, i))
				break
			}

			f, err := jsonparser.ParseFloat(value)
			if err != nil {
				return fmt.Errorf("invalid JSON number: %s", string(value))
			}

			b.Append(builder.C.Double(name, f))

		case jsonparser.Object:
			nested, err := ParseObjectToBuilder(string(value))
			if err != nil {
				return fmt.Errorf("invalid JSON object: %s", string(value))
			}

			b.Append(builder.C.SubDocument(name, nested))

		case jsonparser.Array:
			array, err := ParseArrayToBuilder(string(value))
			if err != nil {
				return fmt.Errorf("invalid JSON array: %s", string(value))
			}

			b.Append(builder.C.Array(name, array))

		case jsonparser.Boolean:
			boolean, err := jsonparser.ParseBoolean(value)
			if err != nil {
				return fmt.Errorf("invalid JSON boolean: %s", string(value))
			}

			b.Append(builder.C.Boolean(name, boolean))

		case jsonparser.Null:
			b.Append(builder.C.Null(name))
		}

		return nil
	}
}

func parseArrayElement(b *builder.ArrayBuilder) arrayElementParser {
	p := parseDocElement(&b.DocumentBuilder)

	return func(index int, value []byte, dataType jsonparser.ValueType, offset int, err error) {
		p([]byte(strconv.FormatUint(uint64(index), 10)), value, dataType, offset)
	}
}
