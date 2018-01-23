package bson

import "github.com/skriptble/wilson/bson/objectid"

type Binary struct {
	Subtype byte
	Data    []byte
}

var Undefined struct{}

var Null struct{}

type Regex struct {
	Pattern string
	Options string
}

type DBPointer struct {
	DB      string
	Pointer objectid.ObjectID
}

type JavaScriptCode string

type Symbol string

type CodeWithScope struct {
	Code  string
	Scope *Document
}

type Timestamp struct {
	T uint32
	I uint32
}

var MinKey struct{}

var MaxKey struct{}
