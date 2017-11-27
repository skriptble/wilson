package bson

import "github.com/skriptble/wilson/parser/ast"

type Constructor struct{}
type ArrayConstructor struct{}
type ModifierConstructor struct{}

func (Constructor) Double(key string, f float64) *ReaderElement                         { return nil }
func (Constructor) String(key string, val string) *ReaderElement                        { return nil }
func (Constructor) Document(key string, elems ...*ReaderElement) *ReaderElement         { return nil }
func (Constructor) Array(key string, elemens ...*ReaderElement) *ReaderElement          { return nil }
func (Constructor) Binary(key string, b []byte, btype uint) *ReaderElement              { return nil }
func (Constructor) ObjectID(key string, obj [12]byte) *ReaderElement                    { return nil }
func (Constructor) Boolean(key string, b bool) *ReaderElement                           { return nil }
func (Constructor) DateTime(key string, dt int64) *ReaderElement                        { return nil }
func (Constructor) Regex(key string, pattern, options string) *ReaderElement            { return nil }
func (Constructor) DBPointer(key string, dbpointer [12]byte) *ReaderElement             { return nil }
func (Constructor) Javascript(key string, js string) *ReaderElement                     { return nil }
func (Constructor) Symbol(key string, symbol string) *ReaderElement                     { return nil }
func (Constructor) CodeWithScope(key string, js string, scope *Document) *ReaderElement { return nil }
func (Constructor) Int32(key string, i int32) *ReaderElement                            { return nil }
func (Constructor) Uint64(key string, u uint64) *ReaderElement                          { return nil }
func (Constructor) Int64(key string, i int64) *ReaderElement                            { return nil }
func (Constructor) Decimal128(key string, d ast.Decimal128) *ReaderElement              { return nil }

func (ArrayConstructor) Double(f float64) *ReaderElement                         { return nil }
func (ArrayConstructor) String(val string) *ReaderElement                        { return nil }
func (ArrayConstructor) Document(elems ...*ReaderElement) *ReaderElement         { return nil }
func (ArrayConstructor) Array(elemens ...*ReaderElement) *ReaderElement          { return nil }
func (ArrayConstructor) Binary(b []byte, btype uint) *ReaderElement              { return nil }
func (ArrayConstructor) ObjectID(obj [12]byte) *ReaderElement                    { return nil }
func (ArrayConstructor) Boolean(b bool) *ReaderElement                           { return nil }
func (ArrayConstructor) DateTime(dt int64) *ReaderElement                        { return nil }
func (ArrayConstructor) Regex(pattern, options string) *ReaderElement            { return nil }
func (ArrayConstructor) DBPointer(dbpointer [12]byte) *ReaderElement             { return nil }
func (ArrayConstructor) Javascript(js string) *ReaderElement                     { return nil }
func (ArrayConstructor) Symbol(symbol string) *ReaderElement                     { return nil }
func (ArrayConstructor) CodeWithScope(js string, scope *Document) *ReaderElement { return nil }
func (ArrayConstructor) Int32(i int32) *ReaderElement                            { return nil }
func (ArrayConstructor) Uint64(u uint64) *ReaderElement                          { return nil }
func (ArrayConstructor) Int64(i int64) *ReaderElement                            { return nil }
func (ArrayConstructor) Decimal128(d ast.Decimal128) *ReaderElement              { return nil }

func (ModifierConstructor) UpdateKey(key string) Modifier                      { return nil }
func (ModifierConstructor) ConvertToDouble(f float64) Modifier                 { return nil }
func (ModifierConstructor) ConvertToString(val string) Modifier                { return nil }
func (ModifierConstructor) ConvertToDocument(elems ...*ReaderElement) Modifier { return nil }
func (ModifierConstructor) ConvertToArray(elemens ...*ReaderElement) Modifier  { return nil }
func (ModifierConstructor) ConvertToBinary(b []byte, btype uint) Modifier      { return nil }
func (ModifierConstructor) ConvertToObjectID(obj [12]byte) Modifier            { return nil }
func (ModifierConstructor) ConvertToBoolean(b bool) Modifier                   { return nil }
func (ModifierConstructor) ConvertToDateTime(dt int64) Modifier                { return nil }
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
