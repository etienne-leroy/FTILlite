// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package crypto

import (
	"encoding/binary"
	"math/big"
)

// Stream wraps a byte array and maintains position information of where the previous write/read occured
type Stream struct {
	bs  []byte
	pos int
}

// Write will write the rw to the stream and advance the position to the end of where the bytes were written
func (b *Stream) Write(rw ReaderWriter) {
	rw.WriteTo(b.bs[b.pos : b.pos+rw.Length()])
	b.pos += rw.Length()
}

// Read will read from the rw and advance the position to the end of where the bytes where read from
func (b *Stream) Read(rw ReaderWriter) {
	rw.ReadFrom(b.bs[b.pos : b.pos+rw.Length()])
	b.pos += rw.Length()
}

// WriteToBytes will write all rws to a new byte array
func WriteToBytes(rws ...ReaderWriter) []byte {
	size := 0
	for _, rw := range rws {
		size += rw.Length()
	}
	w := Stream{make([]byte, size), 0}
	for _, rw := range rws {
		w.Write(rw)
	}
	return w.bs
}

// ReaderWriter interface is used by the Stream type to read or write objects
type ReaderWriter interface {
	// Length returns the number of bytes that will be read or written to the Stream by this ReaderWriter
	Length() int

	// WriteTo will serialiase the current object to the bs array
	WriteTo(bs []byte)

	// ReadFrom will deserialise the bs array into the current object
	ReadFrom(bs []byte)
}

type AsInteger struct {
	i *int
}

func (rw AsInteger) Length() int {
	return 8
}
func (rw AsInteger) WriteTo(bs []byte) {
	binary.BigEndian.PutUint64(bs, uint64(*rw.i))
}
func (rw AsInteger) ReadFrom(bs []byte) {
	*rw.i = int(binary.BigEndian.Uint64(bs))
}

type AsBigInteger struct {
	s int
	b *big.Int
}

func (i AsBigInteger) Length() int {
	return i.s
}
func (i AsBigInteger) WriteTo(bs []byte) {
	i.b.FillBytes(bs)
}
func (i AsBigInteger) ReadFrom(bs []byte) {
	i.b.SetBytes(bs)
}
