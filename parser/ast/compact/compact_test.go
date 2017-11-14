package compact

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestElement(t *testing.T) {
	t.Run("panic", func(t *testing.T) {
		handle := func() {
			if got := recover(); got != ErrUninitializedElement {
				want := ErrUninitializedElement
				t.Errorf("Incorrect value for panic. got %s; want %s", got, want)
			}
		}
		t.Run("key", func(t *testing.T) {
			defer handle()
			(*Element)(nil).Key()
		})
		t.Run("type", func(t *testing.T) {
			defer handle()
			(*Element)(nil).Type()
		})
		t.Run("double", func(t *testing.T) {
			defer handle()
			(*Element)(nil).Double()
		})
		t.Run("string", func(t *testing.T) {
			defer handle()
			(*Element)(nil).String()
		})
		t.Run("document", func(t *testing.T) {
			defer handle()
			(*Element)(nil).Document()
		})
	})
	t.Run("key", func(t *testing.T) {
		buf := []byte{
			'\x00', '\x00', '\x00', '\x00',
			'\x02', 'f', 'o', 'o', '\x00',
			'\x00', '\x00', '\x00', '\x00', '\x00',
			'\x00'}
		e := &Element{start: 4, value: 9, data: buf}
		want := "foo"
		got := e.Key()
		if got != want {
			t.Errorf("Unexpected result. got %s; want %s", got, want)
		}
	})
	t.Run("type", func(t *testing.T) {
		buf := []byte{
			'\x00', '\x00', '\x00', '\x00',
			'\x02', 'f', 'o', 'o', '\x00',
			'\x00', '\x00', '\x00', '\x00', '\x00',
			'\x00',
		}
		e := &Element{start: 4, value: 9, data: buf}
		want := byte('\x02')
		got := e.Type()
		if got != want {
			t.Errorf("Unexpected result. got %v; want %v", got, want)
		}
	})
	t.Run("double", func(t *testing.T) {
		buf := []byte{
			'\x00', '\x00', '\x00', '\x00',
			'\x01', 'f', 'o', 'o', '\x00',
			'\x00', '\x00', '\x00', '\x00',
			'\x00', '\x00', '\x00', '\x00',
			'\x00',
		}
		e := &Element{start: 4, value: 9, data: buf}
		binary.LittleEndian.PutUint64(buf[9:17], math.Float64bits(3.14159))
		want := 3.14159
		got := e.Double()
		if got != want {
			t.Errorf("Unexpected result. got %f; want %f", got, want)
		}
	})
	t.Run("string", func(t *testing.T) {
		buf := []byte{
			'\x00', '\x00', '\x00', '\x00',
			'\x02', 'f', 'o', 'o', '\x00',
			'\x00', '\x00', '\x00', '\x00',
			'b', 'a', 'r', '\x00',
			'\x00',
		}
		e := &Element{start: 4, value: 9, data: buf}
		binary.LittleEndian.PutUint32(buf[9:13], 3)
		want := "bar"
		got := e.String()
		if got != want {
			t.Errorf("Unexpected result. got %f; want %f", got, want)
		}
	})
	t.Run("document", func(t *testing.T) {})
}
