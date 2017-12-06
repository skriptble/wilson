package extjson

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/skriptble/wilson/bson/internal/jsonparser"
	"github.com/skriptble/wilson/builder"
	"github.com/skriptble/wilson/parser/ast"
)

type wrapperType byte

const (
	None     wrapperType = 0
	ObjectId             = iota
	Symbol
	Int32
	Int64
	Double
	Decimal
	Binary
	Code
	Timestamp
	Regex
	DBPointer
	DateTime
	DBRef
	MinKey
	MaxKey
	Undefined
)

func (w wrapperType) String() string {
	switch w {
	case ObjectId:
		return "ObjectId"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Double:
		return "double"
	case Decimal:
		return "decimal"
	case Binary:
		return "binary"
	case Code:
		return "JavaScript code"
	case Timestamp:
		return "timestamp"
	case Regex:
		return "regex"
	case DBPointer:
		return "dbpointer"
	case DateTime:
		return "datetime"
	case DBRef:
		return "dbref"
	case MinKey:
		return "minkey"
	case MaxKey:
		return "maxkey"
	case Undefined:
		return "undefined"
	}

	return "not a wrapper type key"
}

func wrapperKeyType(key []byte) wrapperType {
	switch string(key) {
	case "$numberInt":
		return Int32
	case "$numberLong":
		return Int64
	case "$oid":
		return ObjectId
	case "$symbol":
		return Symbol
	case "$numberDouble":
		return Double
	case "$numberDecimal":
		return Decimal
	case "$binary":
		return Binary
	case "$code":
		fallthrough
	case "$scope":
		return Code
	case "$timestamp":
		return Timestamp
	case "$regularExpression":
		return Regex
	case "$dbPointer":
		return DBPointer
	case "$date":
		return DateTime
	case "$ref":
		fallthrough
	case "$id":
		fallthrough
	case "$db":
		return DBRef
	case "$minKey":
		return MinKey
	case "$maxKey":
		return MaxKey
	case "$undefined":
		return Undefined
	}

	return None
}

func parseObjectId(data []byte, dataType jsonparser.ValueType) ([12]byte, error) {
	var oid [12]byte

	if dataType != jsonparser.String {
		return oid, fmt.Errorf("$oid value should be string, but instead is %s", dataType.String())
	}

	oidBytes, err := hex.DecodeString(string(data))
	if err != nil || len(oidBytes) != 12 {
		return oid, fmt.Errorf("invalid $oid value string: %s", string(data))
	}

	copy(oid[:], oidBytes[:])

	return oid, nil
}

func parseSymbol(data []byte, dataType jsonparser.ValueType) (string, error) {
	if dataType != jsonparser.String {
		return "", fmt.Errorf("$symbol value should be string, but instead is %s", dataType.String())
	}

	str, err := jsonparser.ParseString(data)
	if err != nil {
		return "", fmt.Errorf("invalid escaping in symbol string: %s", string(data))
	}

	return str, nil
}

func parseInt32(data []byte, dataType jsonparser.ValueType) (int32, error) {
	if dataType != jsonparser.String {
		return 0, fmt.Errorf("$numberInt value should be string, but instead is %s", dataType.String())
	}

	i, err := jsonparser.ParseInt(data)
	if err != nil {
		return 0, fmt.Errorf("invalid $numberInt number value: %s", string(data))
	}

	if i < math.MinInt32 || i > math.MaxInt32 {
		return 0, fmt.Errorf("$numberInt value should be int32 but instead is int64: %d", i)
	}

	return int32(i), nil
}

func parseInt64(data []byte, dataType jsonparser.ValueType) (int64, error) {
	if dataType != jsonparser.String {
		return 0, fmt.Errorf("$numberLong value should be string, but instead is %s", dataType.String())
	}

	i, err := jsonparser.ParseInt(data)
	if err != nil {
		return 0, fmt.Errorf("invalid $numberLong number value: %s", string(data))
	}

	return int64(i), nil
}

func parseDouble(data []byte, dataType jsonparser.ValueType) (float64, error) {
	if dataType != jsonparser.String {
		return 0, fmt.Errorf("$numberDouble value should be string, but instead is %s", dataType.String())
	}

	switch string(data) {
	case "Infinity":
		return math.Inf(1), nil
	case "-Infinity":
		return math.Inf(-1), nil
	case "NaN":
		return math.NaN(), nil
	}

	f, err := jsonparser.ParseFloat(data)
	if err != nil {
		return 0, fmt.Errorf("invalid $numberDouble number value: %s", string(data))
	}

	return f, nil
}

func parseDecimal(data []byte, dataType jsonparser.ValueType) (ast.Decimal128, error) {
	if dataType != jsonparser.String {
		return ast.Decimal128{}, fmt.Errorf("$numberDecimal value should be string, but instead is %s", dataType.String())
	}

	d, err := ast.ParseDecimal128(string(data))
	if err != nil {
		return ast.Decimal128{}, fmt.Errorf("$invalid $numberDecimal string: %s", string(data))
	}

	return d, nil
}

func parseBinary(data []byte, dataType jsonparser.ValueType) ([]byte, byte, error) {
	if dataType != jsonparser.Object {
		return nil, 0, fmt.Errorf("$binary value should be object, but instead is %s", dataType.String())
	}

	var b []byte
	var subType *int64

	err := jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		switch string(key) {
		case "base64":
			if b != nil {
				return fmt.Errorf("duplicate base64 key in $binary: %s", string(data))
			}

			if dataType != jsonparser.String {
				return fmt.Errorf("$binary base64 value should be string, but instead is %s", dataType.String())
			}

			base64Bytes, err := base64.StdEncoding.DecodeString(string(value))
			if err != nil {
				return fmt.Errorf("invalid $binary base64 string: %s", string(value))
			}

			b = base64Bytes
		case "subType":
			if subType != nil {
				return fmt.Errorf("duplicate subType key in $binary: %s", string(data))
			}

			if dataType != jsonparser.String {
				return fmt.Errorf("$binary subType value should be string, but instead is %s", dataType.String())
			}

			i, err := strconv.ParseInt(string(value), 16, 64)
			if err != nil {
				return fmt.Errorf("invalid $binary subtype string: %s", string(value))
			}

			subType = &i
		default:
			return fmt.Errorf("invalid key in $binary object: %s", string(key))
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	if b == nil {
		return nil, 0, fmt.Errorf("missing base64 field in $binary object: %s", string(data))
	}

	if subType == nil {
		return nil, 0, fmt.Errorf("missing subType field in $binary object: %s", string(data))

	}

	return b, byte(*subType), nil
}

func parseCode(data []byte, dataType jsonparser.ValueType) (string, error) {
	if dataType != jsonparser.String {
		return "", fmt.Errorf("$code value should be a string, but instead is %s", dataType.String())
	}

	str, err := jsonparser.ParseString(data)
	if err != nil {
		return "", fmt.Errorf("invalid escaping in symbol string: %s", string(data))
	}

	return str, nil
}

func parseScope(data []byte, dataType jsonparser.ValueType) (*builder.DocumentBuilder, error) {
	if dataType != jsonparser.Object {
		return nil, fmt.Errorf("$scope value should be an object, but instead is %s", dataType.String())
	}

	b := builder.NewDocumentBuilder()
	err := parseObjectToBuilder(b, string(data), nil, true)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func parseTimestamp(data []byte, dataType jsonparser.ValueType) (uint32, uint32, error) {
	if dataType != jsonparser.Object {
		return 0, 0, fmt.Errorf("$timestamp value should be object, but instead is %s", dataType.String())
	}

	var time *uint32
	var inc *uint32

	err := jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		switch string(key) {
		case "t":
			if time != nil {
				return fmt.Errorf("duplicate t key in $timestamp: %s", string(data))
			}

			if dataType != jsonparser.Number {
				return fmt.Errorf("$timestamp t value should be string, but instead is %s", dataType.String())
			}

			i, err := jsonparser.ParseInt(value)
			if err != nil {
				return fmt.Errorf("invalid $timestamp t number: %s", string(value))
			}

			if i < 0 || i > math.MaxUint32 {
				return fmt.Errorf("$timestamp t number should be uint32: %s", string(value))
			}

			u := uint32(i)
			time = &u

		case "i":
			if inc != nil {
				return fmt.Errorf("duplicate i key in $timestamp: %s", string(data))
			}

			if dataType != jsonparser.Number {
				return fmt.Errorf("$timestamp i value should be string, but instead is %s", dataType.String())
			}

			i, err := jsonparser.ParseInt(value)
			if err != nil {
				return fmt.Errorf("invalid $timestamp i number: %s", string(value))
			}

			if i < 0 || i > math.MaxUint32 {
				return fmt.Errorf("$timestamp i number should be uint32: %s", string(value))
			}

			u := uint32(i)
			inc = &u

		default:
			return fmt.Errorf("invalid key in $timestamp object: %s", string(key))
		}

		return nil
	})

	if err != nil {
		return 0, 0, err
	}

	if time == nil {
		return 0, 0, fmt.Errorf("missing t field in $timestamp object: %s", string(data))
	}

	if inc == nil {
		return 0, 0, fmt.Errorf("missing i field in $timestamp object: %s", string(data))
	}

	return *time, *inc, nil
}

func parseRegex(data []byte, dataType jsonparser.ValueType) (string, string, error) {
	if dataType != jsonparser.Object {
		return "", "", fmt.Errorf("$regularExpression value should be object, but instead is %s", dataType.String())
	}

	var pat *string
	var opt *string

	err := jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		switch string(key) {
		case "pattern":
			if pat != nil {
				return fmt.Errorf("duplicate pattern key in $regularExpression: %s", string(data))
			}

			if dataType != jsonparser.String {
				return fmt.Errorf("$regularExpression pattern value should be string, but instead is %s", dataType.String())
			}

			str := string(value)
			pat = &str
		case "options":
			if opt != nil {
				return fmt.Errorf("duplicate options key in $regularExpression: %s", string(data))
			}

			if dataType != jsonparser.String {
				return fmt.Errorf("$regularExpression options value should be string, but instead is %s", dataType.String())
			}

			str := string(value)
			opt = &str
		default:
			return fmt.Errorf("invalid key in $regularExpression object: %s", string(key))
		}

		return nil
	})

	if err != nil {
		return "", "", err
	}

	if pat == nil {
		return "", "", fmt.Errorf("missing pattern field in $regularExpression object: %s", string(data))
	}

	if opt == nil {
		return "", "", fmt.Errorf("missing subType field in $regularExpression object: %s", string(data))

	}

	unescapedRegex, err := jsonparser.Unescape([]byte(*pat), nil)
	if err != nil {
		return "", "", fmt.Errorf("invalid escaped $regularExpression pattern: %s", *pat)
	}

	return string(unescapedRegex), *opt, nil
}

func parseDBPointer(data []byte, dataType jsonparser.ValueType) (string, [12]byte, error) {
	var oid [12]byte
	var ns *string
	oidFound := false

	if dataType != jsonparser.Object {
		return "", oid, fmt.Errorf("$dbPointer value should be object, but instead is: %s", dataType.String())
	}

	err := jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		switch string(key) {
		case "$ref":
			if ns != nil {
				return fmt.Errorf("duplicate $ref key in $dbPointer: %s", string(data))
			}

			if dataType != jsonparser.String {
				return fmt.Errorf("$dbPointer $ref value should be string, but instead is %s", dataType.String())
			}

			str := string(value)
			ns = &str
		case "$id":
			if oidFound {
				return fmt.Errorf("duplicate $id key in $dbPointer: %s", string(data))
			}

			if dataType != jsonparser.Object {
				return fmt.Errorf("$dbPointer $id value should be object, but instead is %s", dataType.String())
			}

			id, err := parseDBPointerObjectId(value)
			if err != nil {
				return err
			}

			oid = id
			oidFound = true
		default:
			return fmt.Errorf("invalid key in $dbPointer object: %s", string(key))
		}

		return nil
	})

	if err != nil {
		return "", oid, err
	}

	if ns == nil {
		return "", oid, fmt.Errorf("missing $ref field in $dbPointer object: %s", string(data))
	}

	if !oidFound {
		return "", oid, fmt.Errorf("missing $id field in $dbPointer object: %s", string(data))
	}

	return *ns, oid, nil
}

func parseDBPointerObjectId(data []byte) ([12]byte, error) {
	var oid [12]byte
	oidFound := false

	err := jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		switch string(key) {
		case "$oid":
			if oidFound {
				return fmt.Errorf("duplicate $id key in $dbPointer $oid: %s", string(data))
			}

			if dataType != jsonparser.String {
				return fmt.Errorf("$dbPointer $id $oid value should be string, but instead is %s", dataType.String())
			}

			var err error
			oid, err = parseObjectId(value, dataType)
			if err != nil {
				return fmt.Errorf("invalid $dbPointer $id $oid value: %s", err)
			}

			oidFound = true
		default:
			return fmt.Errorf("invalid key in $dbPointer $id object: %s", string(key))
		}

		return nil
	})

	if err != nil {
		return oid, err
	}

	if !oidFound {
		return oid, fmt.Errorf("missing $oid field in $dbPointer $id object: %s", string(data))
	}

	return oid, nil
}

const RFC3339Milli = "2006-01-02T15:04:05.999Z07:00"

func parseDatetime(data []byte, dataType jsonparser.ValueType) (int64, error) {
	switch dataType {
	case jsonparser.String:
		return parseDatetimeString(data)
	case jsonparser.Object:
		return parseDatetimeObject(data)
	}

	return 0, fmt.Errorf("$date value should be string or object, but instead is %s", dataType.String())
}

func parseDatetimeString(data []byte) (int64, error) {
	t, err := time.Parse(RFC3339Milli, string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid $date value string: %s", string(data))
	}

	return t.Unix() * 1000, nil
}

func parseDatetimeObject(data []byte) (int64, error) {
	var d *int64

	err := jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		switch string(key) {
		case "$numberLong":
			if d != nil {
				return fmt.Errorf("duplicate $numberLong key in $dbPointer: %s", string(data))
			}

			if dataType != jsonparser.String {
				return fmt.Errorf("$date $numberLong field should be string, but instead is %s", dataType.String())
			}

			i, err := parseInt64(value, dataType)
			if err != nil {
				return err
			}

			d = &i

		default:
			return fmt.Errorf("invalid key in $date object: %s", string(key))
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	if d == nil {
		return 0, fmt.Errorf("missing $numberLong field in $date object: %s", string(data))
	}

	return *d, nil
}

func parseRef(data []byte, dataType jsonparser.ValueType) (string, error) {
	if dataType != jsonparser.String {
		return "", fmt.Errorf("$ref value should be string, but instead is %s", dataType.String())
	}

	str, err := jsonparser.ParseString(data)
	if err != nil {
		return "", fmt.Errorf("invalid escaping in $ref string: %s", string(data))
	}

	return str, nil
}

func parseDB(data []byte, dataType jsonparser.ValueType) (string, error) {
	if dataType != jsonparser.String {
		return "", fmt.Errorf("$db value should be string, but instead is %s", dataType.String())
	}

	str, err := jsonparser.ParseString(data)
	if err != nil {
		return "", fmt.Errorf("invalid escaping in $db string: %s", string(data))
	}

	return str, nil
}

func parseMinKey(data []byte, dataType jsonparser.ValueType) error {
	if dataType != jsonparser.Number {
		return fmt.Errorf("$minKey value should be number, but instead is %s", dataType.String())
	}

	i, err := jsonparser.ParseInt(data)
	if err != nil {
		return fmt.Errorf("$minkKey value number is invalid integer: %s", string(data))
	}

	if i != 1 {
		return fmt.Errorf("$minKey value must be 1, but instead is %d", i)
	}

	return nil
}

func parseMaxKey(data []byte, dataType jsonparser.ValueType) error {
	if dataType != jsonparser.Number {
		return fmt.Errorf("$maxKey value should be number, but instead is %s", dataType.String())
	}

	i, err := jsonparser.ParseInt(data)
	if err != nil {
		return fmt.Errorf("$maxkKey value number is invalid integer: %s", string(data))
	}

	if i != 1 {
		return fmt.Errorf("$maxKey value must be 1, but instead is %d", i)
	}

	return nil
}

func parseUndefined(data []byte, dataType jsonparser.ValueType) error {
	if dataType != jsonparser.Boolean {
		return fmt.Errorf("undefined value should be boolean, but instead is %s", dataType.String())
	}

	b, err := jsonparser.ParseBoolean(data)
	if err != nil {
		return fmt.Errorf("$undefined value boolean is invalid: %s", string(data))
	}

	if !b {
		return fmt.Errorf("$undefined balue boolean should be true, but instead is %s", b)
	}

	return nil
}
