package bson

import "github.com/skriptble/wilson/parser/ast"

type RWConstructor struct{}
type RWModifierConstructor struct{}

func (RWConstructor) Double(key string, f float64) *RWElement                           { return nil }
func (RWConstructor) String(key string, val string) *RWElement                          { return nil }
func (RWConstructor) Document(key string, elems ...*RWElement) *RWElement               { return nil }
func (RWConstructor) Array(key string, elemens ...*RWElement) *RWElement                { return nil }
func (RWConstructor) Binary(key string, b []byte, btype uint) *RWElement                { return nil }
func (RWConstructor) ObjectID(key string, obj [12]byte) *RWElement                      { return nil }
func (RWConstructor) Boolean(key string, b bool) *RWElement                             { return nil }
func (RWConstructor) DateTime(key string, dt int64) *RWElement                          { return nil }
func (RWConstructor) Regex(key string, pattern, options string) *RWElement              { return nil }
func (RWConstructor) DBPointer(key string, dbpointer [12]byte) *RWElement               { return nil }
func (RWConstructor) Javascript(key string, js string) *RWElement                       { return nil }
func (RWConstructor) Symbol(key string, symbol string) *RWElement                       { return nil }
func (RWConstructor) CodeWithScope(key string, js string, scope *RWDocument) *RWElement { return nil }
func (RWConstructor) Int32(key string, i int32) *RWElement                              { return nil }
func (RWConstructor) Uint64(key string, u uint64) *RWElement                            { return nil }
func (RWConstructor) Int64(key string, i int64) *RWElement                              { return nil }
func (RWConstructor) Decimal128(key string, d ast.Decimal128) *RWElement                { return nil }

func (RWModifierConstructor) UpdateKey(key string) RWModifier                  { return nil }
func (RWModifierConstructor) ConvertToDouble(f float64) RWModifier             { return nil }
func (RWModifierConstructor) ConvertToString(val string) RWModifier            { return nil }
func (RWModifierConstructor) ConvertToDocument(elems ...*RWElement) RWModifier { return nil }
func (RWModifierConstructor) ConvertToArray(elemens ...*RWElement) RWModifier  { return nil }
func (RWModifierConstructor) ConvertToBinary(b []byte, btype uint) RWModifier  { return nil }
func (RWModifierConstructor) ConvertToObjectID(obj [12]byte) RWModifier        { return nil }
func (RWModifierConstructor) ConvertToBoolean(b bool) RWModifier               { return nil }
func (RWModifierConstructor) ConvertToDateTime(dt int64) RWModifier            { return nil }
func (RWModifierConstructor) ConvertToRegex(pattern, options string) RWModifier {
	return nil
}
func (RWModifierConstructor) ConvertToDBPointer(dbpointer [12]byte) RWModifier { return nil }
func (RWModifierConstructor) ConvertToJavascript(js string) RWModifier         { return nil }
func (RWModifierConstructor) ConvertToSymbol(symbol string) RWModifier         { return nil }
func (RWModifierConstructor) ConvertToCodeWithScope(js string, scope *RWDocument) RWModifier {
	return nil
}
func (RWModifierConstructor) ConvertToInt32(i int32) RWModifier               { return nil }
func (RWModifierConstructor) ConvertToUint64(u uint64) RWModifier             { return nil }
func (RWModifierConstructor) ConvertToInt64(i int64) RWModifier               { return nil }
func (RWModifierConstructor) ConvertToDecimal128(d ast.Decimal128) RWModifier { return nil }
