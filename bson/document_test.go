package bson

import (
	"encoding/binary"
	"testing"
)

func TestDocument(t *testing.T) {
	t.Run("enableWriting", func(t *testing.T) {
		t.Run("on-parent", func(t *testing.T) {})
		t.Run("already-enabled", func(t *testing.T) {
			// An empty document should be write enabled
			d := NewDocument(nil)
			var want error = nil
			got := d.enableWriting()
			if want != got {
				t.Errorf("Empty document should be write enabled. got %v; want %v", got, want)
			}
		})
		t.Run("countElements-error", func(t *testing.T) {
			d := NewDocument([]byte{'\x00'})
			var want error = ErrInvalidReadOnlyDocument
			got := d.enableWriting()
			if want != got {
				t.Errorf("countElements error should be retruned from enableWriting. got %v; want %v", got, want)
			}
		})
		t.Run("success", func(t *testing.T) {
			testCases := []struct {
				name string
			}{}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {})
			}
		})
	})
	t.Run("countElements", func(t *testing.T) {
		t.Run("document-too-small", func(t *testing.T) {
			d := NewDocument([]byte{'\x00'})
			var want error = ErrInvalidReadOnlyDocument
			_, got := d.countElements()
			if want != got {
				t.Errorf("countElements error should be retruned from enableWriting. got %v; want %v", got, want)
			}
		})
		t.Run("incorrect-document-length", func(t *testing.T) {
			b := make([]byte, 5)
			binary.LittleEndian.PutUint32(b[0:4], 200)
			d := NewDocument(b)
			var want error = ErrInvalidReadOnlyDocument
			_, got := d.countElements()
			if want != got {
				t.Errorf("validateKey error should be returned from enableWriting. got %v; want %v", got, want)
			}
		})
		t.Run("keySize-error", func(t *testing.T) {
			b := make([]byte, 4+2+1)
			binary.LittleEndian.PutUint32(b[0:4], 7)
			b[4], b[5], b[6] = '\x01', 'f', 'o'
			d := NewDocument(b)
			var want error = ErrInvalidKey
			_, got := d.countElements()
			if want != got {
				t.Errorf("validateKey error should be returned from enableWriting. got %v; want %v", got, want)
			}
		})
		t.Run("valueSize-error", func(t *testing.T) {
			b := make([]byte, 4+2+2+1)
			binary.LittleEndian.PutUint32(b[0:4], 9)
			b[4], b[5], b[6] = '\x01', 'f', '\x00'
			d := NewDocument(b)
			var want error = ErrTooSmall
			_, got := d.countElements()
			if want != got {
				t.Errorf("validateKey error should be returned from enableWriting. got %v; want %v", got, want)
			}
		})
		t.Run("success", func(t *testing.T) {
			testCases := []struct {
				name string
			}{}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {})
			}
		})
	})
	t.Run("Validate", func(t *testing.T) {
		t.Run("read-only", func(t *testing.T) {
			t.Run("document-too-small", func(t *testing.T) {})
			t.Run("incorrect-document-length", func(t *testing.T) {})
			t.Run("validateKey-error", func(t *testing.T) {})
			t.Run("success", func(t *testing.T) {})
		})
		t.Run("read-write", func(t *testing.T) {})
	})
	t.Run("Walk", func(t *testing.T) {})
	t.Run("Keys", func(t *testing.T) {})
	t.Run("validateKey", func(t *testing.T) {})
	t.Run("Index", func(t *testing.T) {
		t.Run("overwrite-existing-index", func(t *testing.T) {})
		t.Run("incorrect-document-length", func(t *testing.T) {})
		t.Run("validateKey-error", func(t *testing.T) {})
		t.Run("valueSize-error", func(t *testing.T) {})
		t.Run("success", func(t *testing.T) {})
	})
	t.Run("Append", func(t *testing.T) {})
	t.Run("Prepend", func(t *testing.T) {})
	t.Run("Lookup", func(t *testing.T) {
		t.Run("read-only-index", func(t *testing.T) {})
		t.Run("read-only-no-index", func(t *testing.T) {})
		t.Run("read-write-index", func(t *testing.T) {})
		t.Run("read-write-no-index", func(t *testing.T) {})
		t.Run("empty-document", func(t *testing.T) {})
	})
	t.Run("Delete", func(t *testing.T) {})
	t.Run("Update", func(t *testing.T) {})
	t.Run("Err", func(t *testing.T) {})
	t.Run("ElementAt", func(t *testing.T) {})
	t.Run("Iterator", func(t *testing.T) {})
	t.Run("Combine", func(t *testing.T) {})
}
