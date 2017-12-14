package bson

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
)

func ExampleReaderValidate() {
	rdr := make(Reader, 500)
	rdr[250], rdr[251], rdr[252], rdr[253], rdr[254] = '\x05', '\x00', '\x00', '\x00', '\x00'
	n, err := rdr[250:].Validate()
	fmt.Println(n, err)

	// Output: 5 <nil>
}

func BenchmarkReaderValidate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rdr := make(Reader, 500)
		rdr[250], rdr[251], rdr[252], rdr[253], rdr[254] = '\x05', '\x00', '\x00', '\x00', '\x00'
		rdr[250:].Validate()
	}

}

func TestReader(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		t.Run("TooShort", func(t *testing.T) {
			want := ErrTooSmall
			_, got := Reader{'\x00', '\x00'}.Validate()
			if got != want {
				t.Errorf("Did not get expected error. got %v; want %v", got, want)
			}
		})
		t.Run("InvalidLength", func(t *testing.T) {
			want := ErrInvalidLength
			r := make(Reader, 5)
			binary.LittleEndian.PutUint32(r[0:4], 200)
			_, got := r.Validate()
			if got != want {
				t.Errorf("Did not get expected error. got %v; want %v", got, want)
			}
		})
		t.Run("keyLength-error", func(t *testing.T) {
			want := ErrInvalidKey
			r := make(Reader, 8)
			binary.LittleEndian.PutUint32(r[0:4], 8)
			r[4], r[5], r[6], r[7] = '\x02', 'f', 'o', 'o'
			_, got := r.Validate()
			if got != want {
				t.Errorf("Did not get expected error. got %v; want %v", got, want)
			}
		})
		t.Run("Missing-Null-Terminator", func(t *testing.T) {
			want := ErrInvalidReadOnlyDocument
			r := make(Reader, 9)
			binary.LittleEndian.PutUint32(r[0:4], 9)
			r[4], r[5], r[6], r[7], r[8] = '\x0A', 'f', 'o', 'o', '\x00'
			_, got := r.Validate()
			if got != want {
				t.Errorf("Did not get expected error. got %v; want %v", got, want)
			}
		})
		t.Run("validateValue-error", func(t *testing.T) {
			want := ErrTooSmall
			r := make(Reader, 11)
			binary.LittleEndian.PutUint32(r[0:4], 11)
			r[4], r[5], r[6], r[7], r[8], r[9], r[10] = '\x01', 'f', 'o', 'o', '\x00', '\x01', '\x02'
			_, got := r.Validate()
			if got != want {
				t.Errorf("Did not get expected error. got %v; want %v", got, want)
			}
		})
		testCases := []struct {
			name string
			r    Reader
			want uint32
			err  error
		}{
			{"null", Reader{'\x08', '\x00', '\x00', '\x00', '\x0A', 'x', '\x00', '\x00'}, 8, nil},
			{"subdocument",
				Reader{
					'\x15', '\x00', '\x00', '\x00',
					'\x03',
					'f', 'o', 'o', '\x00',
					'\x0B', '\x00', '\x00', '\x00', '\x0A', 'a', '\x00',
					'\x0A', 'b', '\x00', '\x00', '\x00',
				},
				21, nil,
			},
			{"array",
				Reader{
					'\x15', '\x00', '\x00', '\x00',
					'\x04',
					'f', 'o', 'o', '\x00',
					'\x0B', '\x00', '\x00', '\x00', '\x0A', '1', '\x00',
					'\x0A', '2', '\x00', '\x00', '\x00',
				},
				21, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := tc.r.Validate()
				if err != tc.err {
					t.Errorf("Returned error does not match. got %v; want %v", err, tc.err)
				}
				if got != tc.want {
					t.Errorf("Returned size does not match expected size. got %v; want %v", got, tc.want)
				}
			})
		}
	})
	t.Run("Keys", func(t *testing.T) {
		testCases := []struct {
			name      string
			r         Reader
			want      Keys
			err       error
			recursive bool
		}{
			{"one",
				Reader{
					'\x08', '\x00', '\x00', '\x00', '\x0A', 'x', '\x00', '\x00',
				},
				Keys{{Name: "x"}}, nil, false,
			},
			{"two",
				Reader{
					'\x0B', '\x00', '\x00', '\x00', '\x0A', 'x', '\x00',
					'\x0A', 'y', '\x00', '\x00',
				},
				Keys{{Name: "x"}, {Name: "y"}}, nil, false,
			},
			{"one-flat",
				Reader{
					'\x15', '\x00', '\x00', '\x00',
					'\x03',
					'f', 'o', 'o', '\x00',
					'\x0B', '\x00', '\x00', '\x00', '\x0A', 'a', '\x00',
					'\x0A', 'b', '\x00', '\x00', '\x00',
				},
				Keys{{Name: "foo"}}, nil, false,
			},
			{"one-recursive",
				Reader{
					'\x15', '\x00', '\x00', '\x00',
					'\x03',
					'f', 'o', 'o', '\x00',
					'\x0B', '\x00', '\x00', '\x00', '\x0A', 'a', '\x00',
					'\x0A', 'b', '\x00', '\x00', '\x00',
				},
				Keys{{Name: "foo"}, {Prefix: []string{"foo"}, Name: "a"}, {Prefix: []string{"foo"}, Name: "b"}}, nil, true,
			},
			{"one-array-recursive",
				Reader{
					'\x15', '\x00', '\x00', '\x00',
					'\x04',
					'f', 'o', 'o', '\x00',
					'\x0B', '\x00', '\x00', '\x00', '\x0A', '1', '\x00',
					'\x0A', '2', '\x00', '\x00', '\x00',
				},
				Keys{{Name: "foo"}, {Prefix: []string{"foo"}, Name: "1"}, {Prefix: []string{"foo"}, Name: "2"}}, nil, true,
			},
			{"invalid-subdocument",
				Reader{
					'\x15', '\x00', '\x00', '\x00',
					'\x03',
					'f', 'o', 'o', '\x00',
					'\x0B', '\x00', '\x00', '\x00', '\x01', '1', '\x00',
					'\x0A', '2', '\x00', '\x00', '\x00',
				},
				nil, ErrTooSmall, true,
			},
			{"invalid-array",
				Reader{
					'\x15', '\x00', '\x00', '\x00',
					'\x04',
					'f', 'o', 'o', '\x00',
					'\x0B', '\x00', '\x00', '\x00', '\x01', '1', '\x00',
					'\x0A', '2', '\x00', '\x00', '\x00',
				},
				nil, ErrTooSmall, true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := tc.r.Keys(tc.recursive)
				if err != tc.err {
					t.Errorf("Returned error does not match. got %v; want %v", err, tc.err)
				}
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("Returned keys do not match expected keys. got %v; want %v", got, tc.want)
				}
			})
		}
	})
	t.Run("Lookup", func(t *testing.T) {
		testCases := []struct {
			name string
			r    Reader
			key  []string
			want *ReaderElement
			err  error
		}{
			{"first",
				Reader{
					'\x08', '\x00', '\x00', '\x00', '\x0A', 'x', '\x00', '\x00',
				},
				[]string{"x"},
				&ReaderElement{start: 4, value: 7}, nil,
			},
			{"first-second",
				Reader{
					'\x15', '\x00', '\x00', '\x00',
					'\x03',
					'f', 'o', 'o', '\x00',
					'\x0B', '\x00', '\x00', '\x00', '\x0A', 'a', '\x00',
					'\x0A', 'b', '\x00', '\x00', '\x00',
				},
				[]string{"foo", "b"},
				&ReaderElement{start: 7, value: 10}, nil,
			},
			{"first-second-array",
				Reader{
					'\x15', '\x00', '\x00', '\x00',
					'\x04',
					'f', 'o', 'o', '\x00',
					'\x0B', '\x00', '\x00', '\x00', '\x0A', '1', '\x00',
					'\x0A', '2', '\x00', '\x00', '\x00',
				},
				[]string{"foo", "2"},
				&ReaderElement{start: 7, value: 10}, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := tc.r.Lookup(tc.key...)
				if err != tc.err {
					t.Errorf("Returned error does not match. got %v; want %v", err, tc.err)
				}
				if !readerElementEqual(got, tc.want) {
					t.Errorf("Returned element does not match expected element. got %v; want %v", got, tc.want)
				}
			})
		}
	})
	t.Run("ElementAt", func(t *testing.T) {})
	t.Run("Iterator", func(t *testing.T) {})
}

func readerElementEqual(e1, e2 *ReaderElement) bool {
	if e1.start != e2.start {
		return false
	}
	if e1.value != e2.value {
		return false
	}
	return true
}
