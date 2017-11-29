package builder

import "strconv"

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
		db.Key = strconv.FormatUint(uint64(pos), 10)
		return db
	}
}

func (ArrayConstructor) Array(arr *ArrayBuilder) ArrayElementFunc {
	return func(pos uint) Elementer {
		arr.Key = strconv.FormatUint(uint64(pos), 10)
		return arr
	}
}

func (ArrayConstructor) Double(f float64) ArrayElementFunc {
	return func(pos uint) Elementer {
		return C.Double(strconv.FormatUint(uint64(pos), 10), f)
	}
}
