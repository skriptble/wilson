package bson

import (
	"bytes"
	"testing"
	"time"
)

func TestReaderElement(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		t.Run("nil ReaderElement", func(t *testing.T) {
			rdr := (*Element)(nil)
			want := ErrNilElement
			_, got := rdr.Validate()
			if got != want {
				t.Errorf("Did not receive expected error. got %s; want %s", got, want)
			}
		})
		t.Run("keySize error", func(t *testing.T) {
			rdr := Element{start: 0, value: 1, data: []byte{0x0A, 'x'}}
			want := ErrInvalidKey
			_, got := rdr.Validate()
			if got != want {
				t.Errorf("Did not receive expected error. got %s; want %s", got, want)
			}
		})
		t.Run("validateValue error", func(t *testing.T) {
			rdr := Element{start: 0, value: 3, data: []byte{0x01, 'x', 0x00, 0x00}}
			want := ErrTooSmall
			_, got := rdr.Validate()
			if got != want {
				t.Errorf("Did not receive expected error. got %s; want %s", got, want)
			}
		})
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"string",
				&Element{
					start: 0, value: 3,
					data: []byte{0x02, 'x', 0x00, 0x02, 0x00, 0x00, 0x00, 'y', 0x00},
				},
				9, nil,
			},
			{"null", &Element{value: 3, data: []byte{0x0A, 'x', 0x00}}, 3, nil},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.Validate()
				if size != tc.size {
					t.Errorf("Incorrect size returned for validated element. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Incorrect error returned from Validate. got %s; want %s", err, tc.err)
				}
			})
		}
	})
	t.Run("keySize", func(t *testing.T) {
		testCases := []struct {
			name  string
			elem  *Element
			total uint32
			err   error
		}{
			{
				"does not run off end of data", &Element{start: 0, value: 100, data: []byte{0x0A, 'f', 'o', 'o'}},
				3, ErrInvalidKey,
			},
			{
				"stops iteration at start of value",
				&Element{start: 0, value: 4, data: []byte{0x0A, 'f', 'o', 'o', 0x00}},
				3, ErrInvalidKey,
			},
			{
				"returns invalid key error", &Element{start: 0, value: 4, data: []byte{0x0A, 'f', 'o', 'o'}},
				3, ErrInvalidKey,
			},
			{
				"returns correct size on success",
				&Element{start: 0, value: 5, data: []byte{0x0A, 'f', 'o', 'o', 0x00}},
				4, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				total, err := tc.elem.keySize()
				if total != tc.total {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", total, tc.total)
				}
				if err != tc.err {
					t.Errorf("Did not receive correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("valueSize", func(t *testing.T) {
		t.Run("returns too small", func(t *testing.T) {
			testCases := []struct {
				name string
				elem *Element
				size uint32
			}{
				{"subdoc <4", &Element{start: 0, value: 2, data: []byte{0x03, 0x00, 0x00, 0x00}}, 0},
				{"array <4", &Element{start: 0, value: 2, data: []byte{0x04, 0x00, 0x00, 0x00}}, 0},
				{"code-with-scope <4", &Element{start: 0, value: 2, data: []byte{0x0F, 0x00, 0x00, 0x00}}, 0},
				{"subdoc >4", &Element{start: 0, value: 2, data: []byte{0x03, 0x00, 0xFF, 0x00, 0x00, 0x00}}, 4},
				{"array >4", &Element{start: 0, value: 2, data: []byte{0x04, 0x00, 0xFF, 0x00, 0x00, 0x00}}, 4},
				{"code-with-scope >4", &Element{start: 0, value: 2, data: []byte{0x0F, 0x00, 0xFF, 0x00, 0x00, 0x00}}, 4},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					want := ErrTooSmall
					size, got := tc.elem.valueSize()
					if size != tc.size {
						t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
					}
					if got != want {
						t.Errorf("Did not return correct error. got %v; want %v", got, want)
					}
				})
			}
		})
	})
	t.Run("validateValue", testValidateValue)
	t.Run("MarshalBSON", func(t *testing.T) {})
	t.Run("Value Methods", func(t *testing.T) {
		t.Run("Double", func(t *testing.T) {
			testCases := []struct {
				name  string
				elem  *Element
				val   float64
				fault error
			}{
				{"Nil Element", nil, 0, ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, 0, ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, 0, ErrUninitializedElement,
				},
				{"Not Double",
					&Element{start: 0, value: 2, data: []byte{0x02, 0x00}}, 0,
					ElementTypeError{"compact.Element.Double", BSONType(0x02)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{0x01, 0x00, 0x6E, 0x86, 0x1B, 0xF0, 0xF9, 0x21, 0x9, 0x40},
					},
					3.14159, nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					val := tc.elem.Double()
					if val != tc.val {
						t.Errorf("Did not return correct value. got %.5f; want %.5f", val, tc.val)
					}
				})
			}
		})
		t.Run("String", func(t *testing.T) {
			testCases := []struct {
				name  string
				elem  *Element
				val   string
				fault error
			}{
				{"Nil Element", nil, "", ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, "", ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, "", ErrUninitializedElement,
				},
				{"Not String",
					&Element{start: 0, value: 2, data: []byte{0x01, 0x00}}, "",
					ElementTypeError{"compact.Element.String", BSONType(0x01)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{0x02, 0x00, 0x04, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00},
					},
					"foo", nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					val := tc.elem.StringValue()
					if val != tc.val {
						t.Errorf("Did not return correct value. got %s; want %s", val, tc.val)
					}
				})
			}
		})
		t.Run("Embedded Document", func(t *testing.T) {
			testCases := []struct {
				name  string
				elem  *Element
				val   Reader
				fault error
			}{
				{"Nil Element", nil, nil, ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, nil, ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, nil, ErrUninitializedElement,
				},
				{"Not Document",
					&Element{start: 0, value: 2, data: []byte{0x01, 0x00}}, nil,
					ElementTypeError{"compact.Element.Document", BSONType(0x01)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{0x03, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00},
					},
					Reader{0x05, 0x00, 0x00, 0x00, 0x00}, nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					val := tc.elem.ReaderDocument()
					if !bytes.Equal(val, tc.val) {
						t.Errorf("Did not return correct value. got %v; want %v", val, tc.val)
					}
				})
			}
		})
		t.Run("Array", func(t *testing.T) {
			testCases := []struct {
				name  string
				elem  *Element
				val   Reader
				fault error
			}{
				{"Nil Element", nil, nil, ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, nil, ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, nil, ErrUninitializedElement,
				},
				{"Not Array",
					&Element{start: 0, value: 2, data: []byte{0x01, 0x00}}, nil,
					ElementTypeError{"compact.Element.Array", BSONType(0x01)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{0x04, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00},
					},
					Reader{0x05, 0x00, 0x00, 0x00, 0x00}, nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					val := tc.elem.ReaderArray()
					if !bytes.Equal(val, tc.val) {
						t.Errorf("Did not return correct value. got %v; want %v", val, tc.val)
					}
				})
			}
		})
		t.Run("Binary", func(t *testing.T) {
			testCases := []struct {
				name    string
				elem    *Element
				subtype byte
				val     []byte
				fault   error
			}{
				{"Nil Element", nil, 0x00, nil, ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, 0x00, nil, ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, 0x00, nil, ErrUninitializedElement,
				},
				{"Not Binary",
					&Element{start: 0, value: 2, data: []byte{0x01, 0x00}}, 0x00, nil,
					ElementTypeError{"compact.Element.Binary", BSONType(0x01)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{0x05, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 'f', 'o', 'o'},
					},
					0x00, []byte{'f', 'o', 'o'}, nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					subtype, val := tc.elem.Binary()
					if subtype != tc.subtype {
						t.Errorf("Did not return correct subtype. got %v; want %v", subtype, tc.subtype)
					}
					if !bytes.Equal(val, tc.val) {
						t.Errorf("Did not return correct value. got %v; want %v", val, tc.val)
					}
				})
			}
		})
		t.Run("ObjectID", func(t *testing.T) {
			var empty [12]byte
			testCases := []struct {
				name  string
				elem  *Element
				val   [12]byte
				fault error
			}{
				{"Nil Element", nil, empty, ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, empty, ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, empty, ErrUninitializedElement,
				},
				{"Not ObjectID",
					&Element{start: 0, value: 2, data: []byte{0x01, 0x00}}, empty,
					ElementTypeError{"compact.Element.ObejctID", BSONType(0x01)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{
							0x07, 0x00,
							0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
							0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C,
						},
					},
					[12]byte{
						0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
						0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C,
					}, nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					val := tc.elem.ObjectID()
					if !bytes.Equal(val[:], tc.val[:]) {
						t.Errorf("Did not return correct value. got %v; want %v", val, tc.val)
					}
				})
			}
		})
		t.Run("Boolean", func(t *testing.T) {
			testCases := []struct {
				name  string
				elem  *Element
				val   bool
				fault error
			}{
				{"Nil Element", nil, false, ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, false, ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, false, ErrUninitializedElement,
				},
				{"Not Boolean",
					&Element{start: 0, value: 2, data: []byte{0x01, 0x00}}, false,
					ElementTypeError{"compact.Element.Boolean", BSONType(0x01)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{0x08, 0x00, 0x01},
					},
					true, nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					val := tc.elem.Boolean()
					if val != tc.val {
						t.Errorf("Did not return correct value. got %v; want %v", val, tc.val)
					}
				})
			}
		})
		t.Run("UTC DateTime", func(t *testing.T) {
			var empty time.Time
			testCases := []struct {
				name  string
				elem  *Element
				val   time.Time
				fault error
			}{
				{"Nil Element", nil, empty, ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, empty, ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, empty, ErrUninitializedElement,
				},
				{"Not UTC DateTime",
					&Element{start: 0, value: 2, data: []byte{0x01, 0x00}}, empty,
					ElementTypeError{"compact.Element.DateTime", BSONType(0x01)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{0x09, 0x00, 0x80, 0x38, 0x17, 0xB0, 0x60, 0x01, 0x00, 0x00},
					},
					time.Unix(1514782800000/1000, 1514782800000%1000*1000000), nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					val := tc.elem.DateTime()
					if val != tc.val {
						t.Errorf("Did not return correct value. got %v; want %v", val, tc.val)
					}
				})
			}
		})
		t.Run("Regex", func(t *testing.T) {
			testCases := []struct {
				name    string
				elem    *Element
				pattern string
				options string
				fault   error
			}{
				{"Nil Element", nil, "", "", ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, "", "", ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, "", "", ErrUninitializedElement,
				},
				{"Not Regex",
					&Element{start: 0, value: 2, data: []byte{0x01, 0x00}}, "", "",
					ElementTypeError{"compact.Element.Regex", BSONType(0x01)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{0x0B, 0x00, 'f', 'o', 'o', 0x00, 'b', 'a', 'r', 0x00},
					},
					"foo", "bar", nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					pattern, options := tc.elem.Regex()
					if pattern != tc.pattern {
						t.Errorf("Did not return correct pattern. got %v; want %v", pattern, tc.pattern)
					}
					if options != tc.options {
						t.Errorf("Did not return correct value. got %v; want %v", options, tc.options)
					}
				})
			}
		})
		t.Run("DBPointer", func(t *testing.T) {
			var empty [12]byte
			testCases := []struct {
				name    string
				elem    *Element
				ns      string
				pointer [12]byte
				fault   error
			}{
				{"Nil Element", nil, "", empty, ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, "", empty, ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, "", empty, ErrUninitializedElement,
				},
				{"Not DBPointer",
					&Element{start: 0, value: 2, data: []byte{0x01, 0x00}}, "", empty,
					ElementTypeError{"compact.Element.DBPointer", BSONType(0x01)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{
							0x0C, 0x00,
							0x04, 0x00, 0x00, 0x00,
							'f', 'o', 'o', 0x00,
							0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
							0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C,
						},
					},
					"foo", [12]byte{
						0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
						0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C,
					}, nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					ns, pointer := tc.elem.DBPointer()
					if ns != tc.ns {
						t.Errorf("Did not return correct namespace. got %v; want %v", ns, tc.ns)
					}
					if !bytes.Equal(pointer[:], tc.pointer[:]) {
						t.Errorf("Did not return correct pointer. got %v; want %v", pointer, tc.pointer)
					}
				})
			}
		})
		t.Run("JavaScript", func(t *testing.T) {
			testCases := []struct {
				name  string
				elem  *Element
				val   string
				fault error
			}{
				{"Nil Element", nil, "", ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, "", ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, "", ErrUninitializedElement,
				},
				{"Not Javascript",
					&Element{start: 0, value: 2, data: []byte{0x01, 0x00}}, "",
					ElementTypeError{"compact.Element.Javascript", BSONType(0x01)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{0x0D, 0x00, 0x04, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00},
					},
					"foo", nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					val := tc.elem.Javascript()
					if val != tc.val {
						t.Errorf("Did not return correct value. got %s; want %s", val, tc.val)
					}
				})
			}
		})
		t.Run("Symbol", func(t *testing.T) {
			testCases := []struct {
				name  string
				elem  *Element
				val   string
				fault error
			}{
				{"Nil Element", nil, "", ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, "", ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, "", ErrUninitializedElement,
				},
				{"Not Javascript",
					&Element{start: 0, value: 2, data: []byte{0x01, 0x00}}, "",
					ElementTypeError{"compact.Element.Symbol", BSONType(0x01)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{0x0E, 0x00, 0x04, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00},
					},
					"foo", nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					val := tc.elem.Symbol()
					if val != tc.val {
						t.Errorf("Did not return correct value. got %s; want %s", val, tc.val)
					}
				})
			}
		})
		t.Run("Code With Scope", func(t *testing.T) {
			testCases := []struct {
				name  string
				elem  *Element
				code  string
				scope Reader
				fault error
			}{
				{"Nil Element", nil, "", nil, ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, "", nil, ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, "", nil, ErrUninitializedElement,
				},
				{"Not JavascriptWithScope",
					&Element{start: 0, value: 2, data: []byte{0x01, 0x00}}, "", nil,
					ElementTypeError{"compact.Element.JavascriptWithScope", BSONType(0x01)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{
							0x0F, 0x00,
							0x11, 0x00, 0x00, 0x00,
							0x04, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00,
							0x05, 0x00, 0x00, 0x00, 0x00,
						},
					},
					"foo", Reader{0x05, 0x00, 0x00, 0x00, 0x00}, nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					code, scope := tc.elem.ReaderJavascriptWithScope()
					if code != tc.code {
						t.Errorf("Did not return correct code. got %s; want %s", code, tc.code)
					}
					if !bytes.Equal(scope, tc.scope) {
						t.Errorf("Did not return correct scope. got %v; want %v", scope, tc.scope)
					}
				})
			}
		})
		t.Run("Int32", func(t *testing.T) {
			testCases := []struct {
				name  string
				elem  *Element
				val   int32
				fault error
			}{
				{"Nil Element", nil, 0, ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, 0, ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, 0, ErrUninitializedElement,
				},
				{"Not Int32",
					&Element{start: 0, value: 2, data: []byte{0x02, 0x00}}, 0,
					ElementTypeError{"compact.Element.Int32", BSONType(0x02)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{0x10, 0x00, 0xFF, 0x00, 0x00, 0x00},
					},
					255, nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					val := tc.elem.Int32()
					if val != tc.val {
						t.Errorf("Did not return correct value. got %.5f; want %.5f", val, tc.val)
					}
				})
			}
		})
		t.Run("Timestamp", func(t *testing.T) {
			testCases := []struct {
				name  string
				elem  *Element
				val   uint64
				fault error
			}{
				{"Nil Element", nil, 0, ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, 0, ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, 0, ErrUninitializedElement,
				},
				{"Not Timestamp",
					&Element{start: 0, value: 2, data: []byte{0x02, 0x00}}, 0,
					ElementTypeError{"compact.Element.Timestamp", BSONType(0x02)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{0x11, 0x00, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					},
					255, nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					val := tc.elem.Timestamp()
					if val != tc.val {
						t.Errorf("Did not return correct value. got %.5f; want %.5f", val, tc.val)
					}
				})
			}
		})
		t.Run("Int64", func(t *testing.T) {
			testCases := []struct {
				name  string
				elem  *Element
				val   int64
				fault error
			}{
				{"Nil Element", nil, 0, ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, 0, ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, 0, ErrUninitializedElement,
				},
				{"Not Int64",
					&Element{start: 0, value: 2, data: []byte{0x02, 0x00}}, 0,
					ElementTypeError{"compact.Element.Int64", BSONType(0x02)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{0x12, 0x00, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					},
					255, nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					val := tc.elem.Int64()
					if val != tc.val {
						t.Errorf("Did not return correct value. got %.5f; want %.5f", val, tc.val)
					}
				})
			}
		})
		t.Run("Decimal128", func(t *testing.T) {
			var empty Decimal128
			testCases := []struct {
				name  string
				elem  *Element
				val   Decimal128
				fault error
			}{
				{"Nil Element", nil, empty, ErrUninitializedElement},
				{"Empty Element value",
					&Element{start: 0, value: 0, data: nil}, empty, ErrUninitializedElement,
				},
				{"Empty Element data",
					&Element{start: 0, value: 2, data: nil}, empty, ErrUninitializedElement,
				},
				{"Not Int64",
					&Element{start: 0, value: 2, data: []byte{0x02, 0x00}}, empty,
					ElementTypeError{"compact.Element.Decimal128", BSONType(0x02)},
				},
				{"Success",
					&Element{
						start: 0, value: 2,
						data: []byte{
							0x13, 0x00,
							0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
							0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						},
					},
					Decimal128{h: 255, l: 255}, nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					defer func() {
						fault := recover()
						if fault != tc.fault {
							t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
						}
					}()

					val := tc.elem.Decimal128()
					if val != tc.val {
						t.Errorf("Did not return correct value. got %.5f; want %.5f", val, tc.val)
					}
				})
			}
		})
	})
	t.Run("Key", func(t *testing.T) {
		testCases := []struct {
			name  string
			elem  *Element
			key   string
			fault error
		}{
			{"Nil Element", nil, "", ErrUninitializedElement},
			{"Empty Element value",
				&Element{start: 0, value: 0, data: nil}, "", ErrUninitializedElement,
			},
			{"Empty Element data",
				&Element{start: 0, value: 2, data: nil}, "", ErrUninitializedElement,
			},
			{"Success",
				&Element{
					start: 0, value: 5,
					data: []byte{0x01, 'f', 'o', 'o', 0x00, 0x6E, 0x86, 0x1B, 0xF0, 0xF9, 0x21, 0x9, 0x40},
				},
				"foo", nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				defer func() {
					fault := recover()
					if fault != tc.fault {
						t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
					}
				}()

				key := tc.elem.Key()
				if key != tc.key {
					t.Errorf("Did not return correct key. got %s; want %s", key, tc.key)
				}
			})
		}
	})
	t.Run("Type", func(t *testing.T) {
		testCases := []struct {
			name  string
			elem  *Element
			etype byte
			fault error
		}{
			{"Nil Element", nil, 0x0, ErrUninitializedElement},
			{"Empty Element value",
				&Element{start: 0, value: 0, data: nil}, 0x00, ErrUninitializedElement,
			},
			{"Empty Element data",
				&Element{start: 0, value: 2, data: nil}, 0x00, ErrUninitializedElement,
			},
			{"Success",
				&Element{
					start: 0, value: 5,
					data: []byte{0x01, 'f', 'o', 'o', 0x00, 0x6E, 0x86, 0x1B, 0xF0, 0xF9, 0x21, 0x9, 0x40},
				},
				0x01, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				defer func() {
					fault := recover()
					if fault != tc.fault {
						t.Errorf("Did not return the correct error for panic. got %v; want %v", fault, tc.fault)
					}
				}()

				etype := tc.elem.Type()
				if etype != tc.etype {
					t.Errorf("Did not return correct type. got %v; want %v", etype, tc.etype)
				}
			})
		}
	})
}

func testValidateValue(t *testing.T) {
	t.Run("Double", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Too Small",
				&Element{
					start: 0, value: 2,
					data: []byte{0x01, 0x00, 0x00, 0x00},
				},
				0, ErrTooSmall,
			},
			{"Success",
				&Element{
					start: 0, value: 2,
					data: []byte{0x01, 0x00, 0x6E, 0x86, 0x1B, 0xF0, 0xF9, 0x21, 0x9, 0x40},
				},
				8, nil,
			},
		}

		for _, tc := range testCases {
			size, err := tc.elem.validateValue(false)
			if size != tc.size {
				t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
			}
			if err != tc.err {
				t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
			}
		}
	})
	t.Run("String", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			deep bool
			size uint32
			err  error
		}{
			{"Too Small <4",
				&Element{
					start: 0, value: 2,
					data: []byte{0x02, 0x00, 0x00, 0x00},
				},
				true, 0, ErrTooSmall,
			},
			{"Too Small >4",
				&Element{
					start: 0, value: 2,
					data: []byte{0x02, 0x00, 0xFF, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00},
				},
				true, 4, ErrTooSmall,
			},
			{"Invalid String Value",
				&Element{
					start: 0, value: 2,
					data: []byte{0x02, 0x00, 0x03, 0x00, 0x00, 0x00, 'f', 'o', 'o'},
				},
				false, 4, ErrInvalidString,
			},
			{"Shouldn't Deep Validate",
				&Element{
					start: 0, value: 2,
					data: []byte{0x02, 0x00, 0x03, 0x00, 0x00, 0x00, 'f', 'o', 'o'},
				},
				true, 7, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(tc.deep)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("Embedded Document/Array", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			deep bool
			size uint32
			err  error
		}{
			{"Document/too small <4",
				&Element{
					start: 0, value: 2, data: []byte{0x03, 0x00, 0x00, 0x00},
				}, true, 0, ErrTooSmall,
			},
			{"Document/too small >4",
				&Element{
					start: 0, value: 2, data: []byte{0x03, 0x00, 0xFF, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00},
				}, true, 4, ErrTooSmall,
			},
			{"Document/invalid document <5",
				&Element{
					start: 0, value: 2, data: []byte{0x03, 0x00, 0x03, 0x00, 0x00, 0x00, 'f', 'o', 'o'},
				}, true, 4, ErrInvalidReadOnlyDocument,
			},
			{"Document/shouldn't deep validate",
				&Element{
					start: 0, value: 2, data: []byte{0x03, 0x00, 0x09, 0x00, 0x00, 0x00, 'f', 'o', 'o', 'o', 'o'},
				}, true, 9, nil,
			},
			{"Document/should deep validate",
				&Element{
					start: 0, value: 2, data: []byte{0x03, 0x00, 0x09, 0x00, 0x00, 0x00, 'f', 'o', 'o', 'o', 'o'},
				}, false, 9, ErrInvalidKey,
			},
			{"Document/success",
				&Element{
					start: 0, value: 2, data: []byte{0x03, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x0A, 'f', 'o', 'o', 0x00, 0x00},
				}, false, 10, nil,
			},
			{"Array/too small <4",
				&Element{
					start: 0, value: 2, data: []byte{0x04, 0x00, 0x00, 0x00},
				}, true, 0, ErrTooSmall,
			},
			{"Array/too small >4",
				&Element{
					start: 0, value: 2, data: []byte{0x04, 0x00, 0xFF, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00},
				}, true, 4, ErrTooSmall,
			},
			{"Array/invalid document <5",
				&Element{
					start: 0, value: 2, data: []byte{0x04, 0x00, 0x03, 0x00, 0x00, 0x00, 'f', 'o', 'o'},
				}, true, 4, ErrInvalidReadOnlyDocument,
			},
			{"Array/shouldn't deep validate",
				&Element{
					start: 0, value: 2, data: []byte{0x04, 0x00, 0x09, 0x00, 0x00, 0x00, 'f', 'o', 'o', 'o', 'o'},
				}, true, 9, nil,
			},
			{"Array/should deep validate",
				&Element{
					start: 0, value: 2, data: []byte{0x04, 0x00, 0x09, 0x00, 0x00, 0x00, 'f', 'o', 'o', 'o', 'o'},
				}, false, 9, ErrInvalidKey,
			},
			{"Array/success",
				&Element{
					start: 0, value: 2, data: []byte{0x04, 0x00, 0x08, 0x00, 0x00, 0x00, 0x0A, '0', 0x00, 0x00},
				}, false, 8, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(tc.deep)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}

	})
	t.Run("Binary", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Value Too Small",
				&Element{
					start: 0, value: 2, data: []byte{0x05, 0x00, 0x00},
				},
				0, ErrTooSmall,
			},
			{"Invalid Binary Subtype",
				&Element{
					start: 0, value: 2, data: []byte{0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x7F},
				},
				5, ErrInvalidBinarySubtype,
			},
			{"Length Too Small",
				&Element{
					start: 0, value: 2, data: []byte{0x05, 0x00, 0xFF, 0x00, 0x00, 0x00, 0x00},
				},
				5, ErrTooSmall,
			},
			{"Success",
				&Element{
					start: 0, value: 2, data: []byte{0x05, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 'h', 'i'},
				},
				7, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("Undefined", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Success",
				&Element{
					start: 0, value: 2, data: []byte{0x06, 0x00},
				},
				0, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("ObjectID", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Value Too Small",
				&Element{
					start: 0, value: 2, data: []byte{0x07, 0x00, 0x00},
				},
				0, ErrTooSmall,
			},
			{"Success",
				&Element{
					start: 0, value: 2, data: []byte{
						0x07, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				12, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("Boolean", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Too Small",
				&Element{
					start: 0, value: 2, data: []byte{0x08, 0x00},
				},
				0, ErrTooSmall,
			},
			{"Invalid Binary Type",
				&Element{
					start: 0, value: 2, data: []byte{0x08, 0x00, 0x03},
				},
				1, ErrInvalidBooleanType,
			},
			{"True",
				&Element{
					start: 0, value: 2, data: []byte{0x08, 0x00, 0x01},
				},
				1, nil,
			},
			{"False",
				&Element{
					start: 0, value: 2, data: []byte{0x08, 0x00, 0x00},
				},
				1, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("UTC DateTime", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Too Small",
				&Element{
					start: 0, value: 2, data: []byte{0x09, 0x00},
				},
				0, ErrTooSmall,
			},
			{"Success",
				&Element{
					start: 0, value: 2, data: []byte{
						0x09, 0x00,
						0x01, 0x02, 0x03, 0x04,
						0x05, 0x06, 0x07, 0x08,
					},
				},
				8, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("Null", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Success",
				&Element{
					start: 0, value: 2, data: []byte{0x0A, 0x00},
				},
				0, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("Regex", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"First Invalid String",
				&Element{
					start: 0, value: 2, data: []byte{0x0B, 0x00, 'f', 'o', 'o'},
				},
				3, ErrInvalidString,
			},
			{"Second Invalid String",
				&Element{
					start: 0, value: 2, data: []byte{0x0B, 0x00, 'f', 'o', 'o', 0x00, 'b', 'a', 'r'},
				},
				7, ErrInvalidString,
			},
			{"Success",
				&Element{
					start: 0, value: 2, data: []byte{0x0B, 0x00, 'f', 'o', 'o', 0x00, 'b', 'a', 'r', 0x00},
				},
				8, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("DBPointer", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Too Small",
				&Element{
					start: 0, value: 2, data: []byte{0x0C, 0x00},
				},
				0, ErrTooSmall,
			},
			{"Length Too Large",
				&Element{
					start: 0, value: 2, data: []byte{0x0C, 0x00, 0xFF, 0x00, 0x00, 0x00, 0x00},
				},
				4, ErrTooSmall,
			},
			{"Success",
				&Element{
					start: 0, value: 2, data: []byte{
						0x0C, 0x00,
						0x04, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00,
						0x01, 0x02, 0x03, 0x04,
						0x05, 0x06, 0x07, 0x08,
						0x09, 0x0A, 0x0B, 0x0C,
					},
				},
				20, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("JavaScript", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			deep bool
			size uint32
			err  error
		}{
			{"Too Small <4",
				&Element{
					start: 0, value: 2,
					data: []byte{0x0D, 0x00, 0x00, 0x00},
				},
				true, 0, ErrTooSmall,
			},
			{"Too Small >4",
				&Element{
					start: 0, value: 2,
					data: []byte{0x0D, 0x00, 0xFF, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00},
				},
				true, 4, ErrTooSmall,
			},
			{"Invalid String Value",
				&Element{
					start: 0, value: 2,
					data: []byte{0x0D, 0x00, 0x03, 0x00, 0x00, 0x00, 'f', 'o', 'o'},
				},
				false, 4, ErrInvalidString,
			},
			{"Shouldn't Deep Validate",
				&Element{
					start: 0, value: 2,
					data: []byte{0x0D, 0x00, 0x03, 0x00, 0x00, 0x00, 'f', 'o', 'o'},
				},
				true, 7, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(tc.deep)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("Symbol", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			deep bool
			size uint32
			err  error
		}{
			{"Too Small <4",
				&Element{
					start: 0, value: 2,
					data: []byte{0x0E, 0x00, 0x00, 0x00},
				},
				true, 0, ErrTooSmall,
			},
			{"Too Small >4",
				&Element{
					start: 0, value: 2,
					data: []byte{0x0E, 0x00, 0xFF, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00},
				},
				true, 4, ErrTooSmall,
			},
			{"Invalid String Value",
				&Element{
					start: 0, value: 2,
					data: []byte{0x0E, 0x00, 0x03, 0x00, 0x00, 0x00, 'f', 'o', 'o'},
				},
				false, 4, ErrInvalidString,
			},
			{"Shouldn't Deep Validate",
				&Element{
					start: 0, value: 2,
					data: []byte{0x0E, 0x00, 0x03, 0x00, 0x00, 0x00, 'f', 'o', 'o'},
				},
				true, 7, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(tc.deep)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("Code With Scope", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			deep bool
			size uint32
			err  error
		}{
			{"Too Small <4",
				&Element{
					start: 0, value: 2,
					data: []byte{0x0F, 0x00, 0x00, 0x00},
				},
				true, 0, ErrTooSmall,
			},
			{"Too Small >4",
				&Element{
					start: 0, value: 2,
					data: []byte{0x0F, 0x00, 0xFF, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00},
				},
				true, 4, ErrTooSmall,
			},
			{"Shouldn't Deep Validate",
				&Element{
					start: 0, value: 2,
					data: []byte{0x0F, 0x00, 0x07, 0x00, 0x00, 0x00, 'f', 'o', 'o'},
				},
				true, 7, nil,
			},
			{"Deep Validate String Too Large",
				&Element{
					start: 0, value: 2,
					data: []byte{
						0x0F, 0x00,
						0x0C, 0x00, 0x00, 0x00,
						0xFF, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00,
					},
				},
				false, 8, ErrStringLargerThanContainer,
			},
			{"Deep Validate Invalid String",
				&Element{
					start: 0, value: 2,
					data: []byte{
						0x0F, 0x00,
						0x10, 0x00, 0x00, 0x00,
						0x02, 0x00, 0x00, 0x00, 'f', 'o', 'o',
						0xFF, 0x01, 0x02, 0x03, 0x04,
					},
				},
				false, 8, ErrInvalidString,
			},
			{"Deep Validate Invalid Document",
				&Element{
					start: 0, value: 2,
					data: []byte{
						0x0F, 0x00,
						0x11, 0x00, 0x00, 0x00,
						0x04, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00,
						0xFF, 0x00, 0x00, 0x00, 0x00,
					},
				},
				false, 12, ErrInvalidLength,
			},
			{"Success",
				&Element{
					start: 0, value: 2,
					data: []byte{
						0x0F, 0x00,
						0x11, 0x00, 0x00, 0x00,
						0x04, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00,
						0x05, 0x00, 0x00, 0x00, 0x00,
					},
				},
				false, 17, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(tc.deep)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("Int32", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Too Small",
				&Element{
					start: 0, value: 2,
					data: []byte{0x10, 0x00, 0x00, 0x00},
				},
				0, ErrTooSmall,
			},
			{"Success",
				&Element{
					start: 0, value: 2,
					data: []byte{0x10, 0x00, 0x01, 0x02, 0x03, 0x04},
				},
				4, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("Timestamp", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Too Small",
				&Element{
					start: 0, value: 2,
					data: []byte{0x11, 0x00, 0x00, 0x00},
				},
				0, ErrTooSmall,
			},
			{"Success",
				&Element{
					start: 0, value: 2,
					data: []byte{0x11, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
				},
				8, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("Int64", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Too Small",
				&Element{
					start: 0, value: 2,
					data: []byte{0x12, 0x00, 0x00, 0x00},
				},
				0, ErrTooSmall,
			},
			{"Success",
				&Element{
					start: 0, value: 2,
					data: []byte{0x12, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
				},
				8, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("Decimal128", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Too Small",
				&Element{
					start: 0, value: 2,
					data: []byte{0x13, 0x00, 0x00, 0x00},
				},
				0, ErrTooSmall,
			},
			{"Success",
				&Element{
					start: 0, value: 2,
					data: []byte{
						0x13, 0x00,
						0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
						0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x0F,
					},
				},
				16, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("MinKey", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Success",
				&Element{
					start: 0, value: 2, data: []byte{0xFF, 0x00},
				},
				0, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("MaxKey", func(t *testing.T) {
		testCases := []struct {
			name string
			elem *Element
			size uint32
			err  error
		}{
			{"Success",
				&Element{
					start: 0, value: 2, data: []byte{0x7F, 0x00},
				},
				0, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				size, err := tc.elem.validateValue(false)
				if size != tc.size {
					t.Errorf("Did not return correct number of bytes read. got %d; want %d", size, tc.size)
				}
				if err != tc.err {
					t.Errorf("Did not return correct error. got %v; want %v", err, tc.err)
				}
			})
		}
	})
	t.Run("Invalid Element", func(t *testing.T) {
		want := ErrInvalidElement
		var wantSize uint32 = 0
		gotSize, got := (&Element{start: 0, value: 2, data: []byte{0xEE, 0x00}}).validateValue(false)
		if gotSize != wantSize {
			t.Errorf("Did not return correct number of bytes read. got %d; want %d", gotSize, wantSize)
		}
		if got != want {
			t.Errorf("Did not return correct error. got %v; want %v", got, want)
		}
	})
}
