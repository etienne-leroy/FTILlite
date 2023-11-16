// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package variables

type Pickle struct {
	Destination string
	Dtype       string
	Handle      Handle
	Opcode      string
	Data        []byte
	Index       int
	Chunkindex  int
}
