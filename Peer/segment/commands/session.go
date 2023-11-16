// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package commands

import (
	"errors"
	"fmt"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const (
	CommandStartSave  = "command_startsave"  // command_startsave
	CommandSave       = "command_save"       // command_save
	CommandFinishSave = "command_finishsave" // command_finishsave
	CommandStartLoad  = "command_startload"  // command_startload
	CommandLoad       = "command_load"       // command_load
	CommandFinishLoad = "command_finishload" // command_finishload
	CommandDelFile    = "command_delfile"    // command_delfile
)

func StartSave(s SegmentHost, args []string) (string, error) {
	s.SetSaveDestination(args[0])
	s.DeleteFromPickleTable(s.SaveDestination())

	return Ack, nil
}

func Save(s SegmentHost, args []string) (string, error) {
	h := variables.Handle(args[0])
	v, err := s.Variables().Get(h)
	if err != nil {
		return "", err
	}

	if len(args) != 2 {
		return "", fmt.Errorf("incorrect number of arguments for save. Should be 2, %d provided", len(args))
	}
	opcode := args[1]
	t := v.TypeCode()

	tc := t.GetTypeCodeAsSlice()

	for index, tcstring := range tc {
		b, err := v.GetBinaryArray(index)
		if err != nil {
			return "", err
		}

		err = s.SaveToPickleTable(types.TypeCode(tcstring), h, opcode, index, b)
		if err != nil {
			return "", err
		}
	}

	return Ack, nil
}

func FinishSave(s SegmentHost, args []string) (string, error) {
	s.SetSaveDestination("")
	return Ack, nil
}

func StartLoad(s SegmentHost, args []string) (string, error) {
	s.SetLoadDestination(args[0])
	return Ack, nil
}

func Load(s SegmentHost, args []string) (string, error) {
	t := variables.Handle(args[0])
	h := variables.Handle(args[1])
	expected_tc := args[2:]

	if len(args) < 3 {
		return "", fmt.Errorf("load command requires 3 or more arguments")
	}

	pickles, err := s.LoadFromPickleTable(h)
	if err != nil {
		return "", err
	}

	if len(pickles) == 0 {
		msg := fmt.Sprintf("no data retured from pickle table for handle %s", h)
		return "", errors.New(msg)
	}

	isListMap := expected_tc[0] == "listmap"
	if isListMap {
		expected_tc = expected_tc[1:]
	}

	var lmArray = make([]types.ArrayTypeVal, len(expected_tc))
	var tc = make([]types.TypeCode, len(expected_tc))

	for index, p := range expected_tc {
		var v types.ArrayTypeVal
		d := []byte{}
		t := types.TypeCode(p)

		if len(pickles) == len(expected_tc) && len(pickles[index].Dtype) > 0 {
			tDB := types.TypeCode(pickles[index].Dtype)
			if t != tDB {
				return "", fmt.Errorf("expected type '%v' does not match the variable's stored type '%v'", t, tDB)
			}
			d = pickles[index].Data
		}

		tc[index] = t
		v, err = types.FromBytes(t, d)
		if err != nil {
			return "", err
		}

		lmArray[index] = v
	}

	var v types.TypeVal
	if !isListMap {
		v = lmArray[0]
	} else {
		lmGoArray := make([]types.ArrayElementTypeVal, len(lmArray))

		for i, xs := range lmArray {
			var ok bool
			lmGoArray[i], ok = xs.(types.ArrayElementTypeVal)
			if !ok {
				return "", errors.New("only Integer, Float, Bytearray and Ed25519Int arrays can be used in listmaps")
			}
		}

		v, err = types.NewListMapFromArrays(tc, lmGoArray, "any")
		if err != nil {
			return "", err
		}
	}

	s.Variables().Set(t, v)

	return Ack, nil
}

func FinishLoad(s SegmentHost, args []string) (string, error) {
	s.SetLoadDestination("")
	return Ack, nil
}

func Delfile(s SegmentHost, args []string) (string, error) {
	d := args[0]
	err := s.DeleteFromPickleTable(d)
	if err != nil {
		return "", err
	}
	return Ack, nil
}
