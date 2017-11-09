package builder

type Builder struct {
}

type ElementConstructor interface {
	Element() Element
}

type ElementConstructorFunc func() Element

func (ecf ElementConstructorFunc) Element() Element {
	return ecf()
}

type Element func() (length uint, ep ElementPrinter)

// writer can be:
//
// - []byte
// - io.WriterAt
// - io.WriteSeeker
// - io.Writer
//
// If it is not one of these values, the implementations should panic.
type ElementPrinter func(start uint, writer interface{})

func (Builder) Document(elems ...Element) {}
