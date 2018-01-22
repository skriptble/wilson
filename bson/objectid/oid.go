// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// Based on gopkg.in/mgo.v2/bson by Gustavo Niemeyer
// See THIRD-PARTY-NOTICES for original license terms.

package objectid

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

type ObjectID [12]byte

var objectIDCounter = readRandomUint32()
var processUnique = readRandomUint32()

func New() ObjectID {
	var b [12]byte

	binary.BigEndian.PutUint32(b[0:4], uint32(time.Now().Unix()))
	binary.BigEndian.PutUint32(b[4:8], processUnique)
	binary.BigEndian.PutUint32(b[9:12], atomic.AddUint32(&objectIDCounter, 1))

	return b
}

func (id *ObjectID) Hex() string {
	return hex.EncodeToString(id[:])
}

func readRandomUint32() uint32 {
	var b [4]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot read random object id: %v", err))
	}

	return (uint32(b[0]) << 0) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24)
}