package bson

import "github.com/skriptble/wilson/parser/ast"

type Constructor struct{}
type ModifierConstructor struct{}

func (Constructor) Double(key string, f float64) *Element                         { return nil }
func (Constructor) String(key string, val string) *Element                        { return nil }
func (Constructor) Document(key string, elems ...*Element) *Element               { return nil }
func (Constructor) Array(key string, elemens ...*Element) *Element                { return nil }
func (Constructor) Binary(key string, b []byte, btype uint) *Element              { return nil }
func (Constructor) ObjectID(key string, obj [12]byte) *Element                    { return nil }
func (Constructor) Boolean(key string, b bool) *Element                           { return nil }
func (Constructor) DateTime(key string, dt int64) *Element                        { return nil }
func (Constructor) Regex(key string, pattern, options string) *Element            { return nil }
func (Constructor) DBPointer(key string, dbpointer [12]byte) *Element             { return nil }
func (Constructor) Javascript(key string, js string) *Element                     { return nil }
func (Constructor) Symbol(key string, symbol string) *Element                     { return nil }
func (Constructor) CodeWithScope(key string, js string, scope *Document) *Element { return nil }
func (Constructor) Int32(key string, i int32) *Element                            { return nil }
func (Constructor) Uint64(key string, u uint64) *Element                          { return nil }
func (Constructor) Int64(key string, i int64) *Element                            { return nil }
func (Constructor) Decimal128(key string, d ast.Decimal128) *Element              { return nil }

func (ModifierConstructor) UpdateKey(key string) Modifier                 { return nil }
func (ModifierConstructor) ConvertToDouble(f float64) Modifier            { return nil }
func (ModifierConstructor) ConvertToString(val string) Modifier           { return nil }
func (ModifierConstructor) ConvertToDocument(elems ...*Element) Modifier  { return nil }
func (ModifierConstructor) ConvertToArray(elemens ...*Element) Modifier   { return nil }
func (ModifierConstructor) ConvertToBinary(b []byte, btype uint) Modifier { return nil }
func (ModifierConstructor) ConvertToObjectID(obj [12]byte) Modifier       { return nil }
func (ModifierConstructor) ConvertToBoolean(b bool) Modifier              { return nil }
func (ModifierConstructor) ConvertToDateTime(dt int64) Modifier           { return nil }
func (ModifierConstructor) ConvertToRegex(pattern, options string) Modifier {
	return nil
}
func (ModifierConstructor) ConvertToDBPointer(dbpointer [12]byte) Modifier { return nil }
func (ModifierConstructor) ConvertToJavascript(js string) Modifier         { return nil }
func (ModifierConstructor) ConvertToSymbol(symbol string) Modifier         { return nil }
func (ModifierConstructor) ConvertToCodeWithScope(js string, scope *Document) Modifier {
	return nil
}
func (ModifierConstructor) ConvertToInt32(i int32) Modifier               { return nil }
func (ModifierConstructor) ConvertToUint64(u uint64) Modifier             { return nil }
func (ModifierConstructor) ConvertToInt64(i int64) Modifier               { return nil }
func (ModifierConstructor) ConvertToDecimal128(d ast.Decimal128) Modifier { return nil }
