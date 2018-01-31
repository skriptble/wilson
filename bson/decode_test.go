package bson

import (
	"bytes"
	"testing"

	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestDecoder(t *testing.T) {
	t.Run("byte slice", func(t *testing.T) {
		testCases := []struct {
			name     string
			reader   *bytes.Buffer
			expected []byte
			actual   []byte
			err      error
		}{
			{
				"nil",
				bytes.NewBuffer([]byte{0x5, 0x0, 0x0, 0x0, 0x0}),
				nil,
				nil,
				ErrTooSmall,
			},
			{
				"empty slice",
				bytes.NewBuffer([]byte{0x5, 0x0, 0x0, 0x0}),
				nil,
				[]byte{},
				ErrTooSmall,
			},
			{
				"too small",
				bytes.NewBuffer([]byte{
					0x5, 0x0, 0x0, 0x0, 0x0,
				}),
				nil,
				make([]byte, 0x4),
				ErrTooSmall,
			},
			{
				"empty doc",
				bytes.NewBuffer([]byte{
					0x5, 0x0, 0x0, 0x0, 0x0,
				}),
				[]byte{0x5, 0x0, 0x0, 0x0, 0x0},
				make([]byte, 0x5),
				nil,
			},
			{
				"non-empty doc",
				bytes.NewBuffer([]byte{
					// length
					0x17, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - null
					0xa,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,

					// null terminator
					0x0,
				}),
				[]byte{
					// length
					0x17, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - null
					0xa,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,

					// null terminator
					0x0,
				},
				make([]byte, 0x17),
				nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				d := NewDecoder(tc.reader)

				err := d.Decode(tc.actual)
				require.Equal(t, tc.err, err)
				if err != nil {
					return
				}

				require.True(t, bytes.Equal(tc.expected, tc.actual))
			})
		}
	})

	t.Run("Reader", func(t *testing.T) {
		testCases := []struct {
			name     string
			reader   *bytes.Buffer
			expected Reader
			actual   Reader
			err      error
		}{
			{
				"nil",
				bytes.NewBuffer([]byte{0x5, 0x0, 0x0, 0x0, 0x0}),
				nil,
				nil,
				ErrTooSmall,
			},
			{
				"empty slice",
				bytes.NewBuffer([]byte{0x5, 0x0, 0x0, 0x0}),
				nil,
				[]byte{},
				ErrTooSmall,
			},
			{
				"too small",
				bytes.NewBuffer([]byte{
					0x5, 0x0, 0x0, 0x0, 0x0,
				}),
				nil,
				make([]byte, 0x4),
				ErrTooSmall,
			},
			{
				"empty doc",
				bytes.NewBuffer([]byte{
					0x5, 0x0, 0x0, 0x0, 0x0,
				}),
				[]byte{0x5, 0x0, 0x0, 0x0, 0x0},
				make([]byte, 0x5),
				nil,
			},
			{
				"non-empty doc",
				bytes.NewBuffer([]byte{
					// length
					0x17, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - null
					0xa,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,

					// null terminator
					0x0,
				}),
				[]byte{
					// length
					0x17, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - null
					0xa,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,

					// null terminator
					0x0,
				},
				make([]byte, 0x17),
				nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				d := NewDecoder(tc.reader)

				err := d.Decode(tc.actual)
				require.Equal(t, tc.err, err)
				if err != nil {
					return
				}

				require.True(t, bytes.Equal(tc.expected, tc.actual))
			})
		}
	})

	t.Run("io.Writer", func(t *testing.T) {
		testCases := []struct {
			name     string
			reader   *bytes.Buffer
			expected *bytes.Buffer
			actual   *bytes.Buffer
			err      error
		}{
			{
				"empty doc",
				bytes.NewBuffer([]byte{
					0x5, 0x0, 0x0, 0x0, 0x0,
				}),
				bytes.NewBuffer([]byte{
					0x5, 0x0, 0x0, 0x0, 0x0,
				}),
				bytes.NewBuffer([]byte{}),
				nil,
			},
			{
				"non-empty doc",
				bytes.NewBuffer([]byte{
					// length
					0x17, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - null
					0xa,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,

					// null terminator
					0x0,
				}),
				bytes.NewBuffer([]byte{
					// length
					0x17, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - null
					0xa,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,

					// null terminator
					0x0,
				}),
				bytes.NewBuffer([]byte{}),
				nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				d := NewDecoder(tc.reader)

				err := d.Decode(tc.actual)
				require.Equal(t, tc.err, err)
				if err != nil {
					return
				}

				require.Equal(t, tc.expected, tc.actual)
			})
		}
	})

	t.Run("Unmarshaler", func(t *testing.T) {
		testCases := []struct {
			name     string
			reader   *bytes.Buffer
			expected *Document
			actual   *Document
			err      error
		}{
			{
				"empty doc",
				bytes.NewBuffer([]byte{
					0x5, 0x0, 0x0, 0x0, 0x0,
				}),
				NewDocument(0),
				NewDocument(0),
				nil,
			},
			{
				"non-empty doc",
				bytes.NewBuffer([]byte{
					// length
					0x17, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - null
					0xa,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,

					// null terminator
					0x0,
				}),
				NewDocument(2).Append(
					C.String("foo", "bar"),
					C.Null("baz"),
				),
				NewDocument(0),
				nil,
			},
			{
				"nested doc",
				bytes.NewBuffer([]byte{
					// length
					0x26, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - document
					0x3,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,

					// -- begin subdocument --

					// length
					0xf, 0x0, 0x0, 0x0,

					// type - int32
					0x10,
					// key - "bang"
					0x62, 0x61, 0x6e, 0x67, 0x0,
					// value - int32(12)
					0xc, 0x0, 0x0, 0x0,

					// null terminator
					0x0,

					// -- end subdocument

					// null terminator
					0x0,
				}),
				NewDocument(2).Append(
					C.String("foo", "bar"),
					C.SubDocumentFromElements("baz",
						C.Int32("bang", 12),
					),
				),
				NewDocument(0),
				nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				d := NewDecoder(tc.reader)

				err := d.Decode(tc.actual)
				require.Equal(t, tc.err, err)
				if err != nil {
					return
				}

				require.True(t, documentComparer(tc.expected, tc.actual))
			})
		}
	})

	t.Run("map", func(t *testing.T) {
		testCases := []struct {
			name     string
			reader   *bytes.Buffer
			expected map[string]interface{}
			actual   map[string]interface{}
			err      error
		}{
			{
				"empty doc",
				bytes.NewBuffer([]byte{
					0x5, 0x0, 0x0, 0x0, 0x0,
				}),
				make(map[string]interface{}),
				make(map[string]interface{}),
				nil,
			},
			{
				"non-empty doc",
				bytes.NewBuffer([]byte{
					// length
					0x1b, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - int32
					0x10,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,
					// value - 32
					0x20, 0x0, 0x0, 0x0,

					// null terminator
					0x0,
				}),
				map[string]interface{}{
					"foo": "bar",
					"baz": int32(32),
				},
				make(map[string]interface{}),
				nil,
			},
			{
				"nested doc",
				bytes.NewBuffer([]byte{
					// length
					0x26, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - document
					0x3,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,

					// -- begin subdocument --

					// length
					0xf, 0x0, 0x0, 0x0,

					// type - int32
					0x10,
					// key - "bang"
					0x62, 0x61, 0x6e, 0x67, 0x0,
					// value - int32(12)
					0xc, 0x0, 0x0, 0x0,

					// null terminator
					0x0,

					// -- end subdocument

					// null terminator
					0x0,
				}),
				map[string]interface{}{
					"foo": "bar",
					"baz": map[string]interface{}{
						"bang": int32(12),
					},
				},
				make(map[string]interface{}),
				nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				d := NewDecoder(tc.reader)

				err := d.Decode(tc.actual)
				require.Equal(t, tc.err, err)
				if err != nil {
					return
				}

				require.True(t, cmp.Equal(tc.expected, tc.actual))
			})
		}
	})

	t.Run("element slice", func(t *testing.T) {
		testCases := []struct {
			name     string
			reader   *bytes.Buffer
			expected []*Element
			actual   []*Element
			err      error
		}{
			{
				"empty doc",
				bytes.NewBuffer([]byte{
					0x5, 0x0, 0x0, 0x0, 0x0,
				}),
				[]*Element{},
				[]*Element{},
				nil,
			},
			{
				"non-empty doc",
				bytes.NewBuffer([]byte{
					// length
					0x1b, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - int32
					0x10,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,
					// value - 32
					0x20, 0x0, 0x0, 0x0,

					// null terminator
					0x0,
				}),
				[]*Element{
					C.String("foo", "bar"),
					C.Int32("baz", 32),
				},
				make([]*Element, 2),
				nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				d := NewDecoder(tc.reader)

				err := d.Decode(tc.actual)
				require.Equal(t, tc.err, err)
				if err != nil {
					return
				}

				elementSliceEqual(t, tc.expected, tc.actual)
			})
		}
	})

	t.Run("struct", func(t *testing.T) {
		testCases := []struct {
			name     string
			reader   *bytes.Buffer
			expected interface{}
			actual   interface{}
			err      error
		}{
			{
				"empty doc",
				bytes.NewBuffer([]byte{
					0x5, 0x0, 0x0, 0x0, 0x0,
				}),
				&struct{}{},
				&struct{}{},
				nil,
			},
			{
				"non-empty doc",
				bytes.NewBuffer([]byte{
					// length
					0x25, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - int32
					0x10,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,
					// value - 32
					0x20, 0x0, 0x0, 0x0,

					// type - regex
					0xb,
					// key - "r"
					0x72, 0x0,
					// value - pattern("WoRd")
					0x57, 0x6f, 0x52, 0x64, 0x0,
					// value - options("i")
					0x69, 0x0,

					// null terminator
					0x0,
				}),
				&struct {
					Foo string
					Baz int32
					R   Regex
				}{
					"bar",
					32,
					Regex{Pattern: "WoRd", Options: "i"},
				},
				&struct {
					Foo string
					Baz int32
					R   Regex
				}{},
				nil,
			},
			{
				"nested doc",
				bytes.NewBuffer([]byte{
					// length
					0x26, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - document
					0x3,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,

					// -- begin subdocument --

					// length
					0xf, 0x0, 0x0, 0x0,

					// type - int32
					0x10,
					// key - "bang"
					0x62, 0x61, 0x6e, 0x67, 0x0,
					// value - int32(12)
					0xc, 0x0, 0x0, 0x0,

					// null terminator
					0x0,

					// -- end subdocument

					// null terminator
					0x0,
				}),
				&struct {
					Foo string
					Baz struct {
						Bang int32
					}
				}{
					"bar",
					struct{ Bang int32 }{12},
				},
				&struct {
					Foo string
					Baz struct {
						Bang int32
					}
				}{},
				nil,
			},
			{
				"struct tags",
				bytes.NewBuffer([]byte{
					// length
					0x1b, 0x0, 0x0, 0x0,

					// type - string
					0x2,
					// key - "foo"
					0x66, 0x6f, 0x6f, 0x0,
					// value - string length
					0x4, 0x0, 0x0, 0x0,
					// value - string "bar"
					0x62, 0x61, 0x72, 0x0,

					// type - int32
					0x10,
					// key - "baz"
					0x62, 0x61, 0x7a, 0x0,
					// value - 32
					0x20, 0x0, 0x0, 0x0,

					// null terminator
					0x0,
				}),
				&struct {
					A string `bson:"foo"`
					B int32  `bson:"baz,omitempty"`
				}{
					"bar",
					32,
				},
				&struct {
					A string `bson:"foo"`
					B int32  `bson:"baz,omitempty"`
				}{},
				nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				d := NewDecoder(tc.reader)

				err := d.Decode(tc.actual)
				require.Equal(t, tc.err, err)
				if err != nil {
					return
				}

				require.True(t, reflect.DeepEqual(tc.expected, tc.actual))
			})
		}
	})
}

func elementSliceEqual(t *testing.T, e1 []*Element, e2 []*Element) {
	require.Equal(t, len(e1), len(e2))

	for i := range e1 {
		require.True(t, readerElementComparer(e1[i], e2[i]))
	}
}
