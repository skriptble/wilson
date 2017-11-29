package builder

import (
	"strconv"

	"github.com/skriptble/wilson/parser/ast"
)

type ArrayElementer interface {
	ArrayElement(pos uint) Elementer
}

type ArrayElementFunc func(pos uint) Elementer

func (aef ArrayElementFunc) ArrayElement(pos uint) Elementer {
	return aef(pos)
}

type ArrayBuilder struct {
	DocumentBuilder
	current uint
}

func (ab *ArrayBuilder) Append(elems ...ArrayElementer) *ArrayBuilder {
	ab.Init()
	for _, arrelem := range elems {
		sizer, f := arrelem.ArrayElement(ab.current).Element()
		ab.current++
		ab.funcs = append(ab.funcs, f)
		ab.sizers = append(ab.sizers, sizer)
	}
	return ab
}

func (ArrayConstructor) SubDocument(db *DocumentBuilder) ArrayElementFunc {
	return func(pos uint) Elementer {
		key := strconv.FormatUint(uint64(pos), 10)
		return C.SubDocument(key, db)
	}
}

func (ArrayConstructor) SubDocumentWithElements(elems ...Elementer) ArrayElementFunc {
	return func(pos uint) Elementer {
		key := strconv.FormatUint(uint64(pos), 10)
		return C.SubDocumentWithElements(key, elems...)
	}
}

func (ArrayConstructor) Array(arr *ArrayBuilder) ArrayElementFunc {
	return func(pos uint) Elementer {
		key := strconv.FormatUint(uint64(pos), 10)
		return C.Array(key, arr)
	}
}

func (ArrayConstructor) ArrayWithElements(elems ...ArrayElementer) ArrayElementFunc {
	return func(pos uint) Elementer {
		key := strconv.FormatUint(uint64(pos), 10)
		return C.ArrayWithElements(key, elems...)
	}
}

func (ArrayConstructor) Double(f float64) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.Double(strconv.FormatUint(uint64(pos), 10), f)
	}
}

func (ArrayConstructor) String(s string) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.String(strconv.FormatUint(uint64(pos), 10), s)
	}
}

func (ArrayConstructor) Binary(b []byte) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.Binary(strconv.FormatUint(uint64(pos), 10), b)
	}
}

func (ArrayConstructor) BinaryWithSubtype(b []byte, btype byte) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.BinaryWithSubtype(strconv.FormatUint(uint64(pos), 10), b, btype)
	}
}

func (ArrayConstructor) Undefined() ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.Undefined(strconv.FormatUint(uint64(pos), 10))
	}
}

func (ArrayConstructor) ObjectId(oid [12]byte) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.ObjectId(strconv.FormatUint(uint64(pos), 10), oid)
	}
}

func (ArrayConstructor) Boolean(b bool) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.Boolean(strconv.FormatUint(uint64(pos), 10), b)
	}
}

func (ArrayConstructor) DateTime(dt int64) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.DateTime(strconv.FormatUint(uint64(pos), 10), dt)
	}
}

func (ArrayConstructor) Null() ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.Null(strconv.FormatUint(uint64(pos), 10))
	}
}

func (ArrayConstructor) Regex(pattern string, options string) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.Regex(strconv.FormatUint(uint64(pos), 10), pattern, options)
	}
}

func (ArrayConstructor) DBPointer(ns string, oid [12]byte) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.DBPointer(strconv.FormatUint(uint64(pos), 10), ns, oid)
	}
}

func (ArrayConstructor) JavaScriptCode(code string) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.JavaScriptCode(strconv.FormatUint(uint64(pos), 10), code)
	}
}

func (ArrayConstructor) Symbol(symbol string) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.Symbol(strconv.FormatUint(uint64(pos), 10), symbol)
	}
}

func (ArrayConstructor) CodeWithScope(code string, scope []byte) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.CodeWithScope(strconv.FormatUint(uint64(pos), 10), code, scope)
	}
}

func (ArrayConstructor) Int32(i int32) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.Int32(strconv.FormatUint(uint64(pos), 10), i)
	}
}

func (ArrayConstructor) Timestamp(t uint64) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.Timestamp(strconv.FormatUint(uint64(pos), 10), t)
	}
}

func (ArrayConstructor) Int64(i int64) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.Int64(strconv.FormatUint(uint64(pos), 10), i)
	}
}

func (ArrayConstructor) Decimal(d ast.Decimal128) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.Decimal(strconv.FormatUint(uint64(pos), 10), d)
	}
}

func (ArrayConstructor) MinKey() ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.MinKey(strconv.FormatUint(uint64(pos), 10))
	}
}

func (ArrayConstructor) MaxKey() ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.MaxKey(strconv.FormatUint(uint64(pos), 10))
	}
}
