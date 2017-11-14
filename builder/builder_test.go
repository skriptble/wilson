package builder

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDocumentBuilder(t *testing.T) {
	t.Run("Basic-Construction", func(t *testing.T) {
		b := make([]byte, 41)
		d := new(DocumentBuilder).Append(C.Double("foo", 3.14159), C.SubDocument("bar", C.Double("baz", 3.14159)))
		fmt.Println(d.RequiredBytes())
		n, err := d.WriteDocument(b)
		fmt.Println(n, err)
		fmt.Println(b)
	})
	t.Run("Static-Functions", func(t *testing.T) {
		t.Run("Double", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				f       float64
				size    uint
				b       []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", 3.14159, 13,
					[]byte{
						0x1, 0x66, 0x6f, 0x6f, 0x0,
						0x6e, 0x86, 0x1b, 0xf0, 0xf9,
						0x21, 0x9, 0x40},
					0, 13, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).Double(tc.key, tc.f)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.b) {
							t.Errorf("Written bytes do not match. got %#v; want %v", b, tc.b)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})
	})
}
