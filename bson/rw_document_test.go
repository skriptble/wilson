package bson

import "fmt"

func ExampleRWDocument() {
	rwd := new(RWDocument)
	rwc := RWConstructor{}
	rwm := RWModifierConstructor{}
	rwd.Insert(rwc.Double("foo", 3.14159), rwc.Boolean("bar", true), rwc.Int32("baz", 12345))
	b, err := rwd.Lookup("bar")
	if err != nil {
		panic(err)
	}
	if t := b.Type(); t == '\x08' {
		fmt.Println(b.Boolean())
	}

	err = rwd.Update(rwm.UpdateKey("qux"), "bar").Update(rwm.UpdateKey("baz"), "foobar").Err()
	if err != nil {
	}
}
