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
	"fmt"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const (
	CommandSerialise       = "command_serialise"   // command_serialise <hResult :: Handle->[](int64)>
	CommandDeserialise     = "command_deserialise" // command_deserialise <hResult :: Handle->[](int64)>
	CommandEdFolded        = "command_ed_folded"
	CommandEdAffine        = "command_ed_affine"
	CommandEdFoldedProject = "command_ed_folded_project"
	CommandEdAffineProject = "command_ed_affine_project"
)

func Serialise(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hValues := variables.Handle(args[1])

	v, err := s.Variables().Get(hValues)
	if err != nil {
		return "", err
	}

	b, err := v.GetBinaryArray(-1)
	if err != nil {
		return "", err
	}
	width := v.TypeCode().Length()
	bs := make([][]byte, 0)

	if width > 0 && len(b) > 0 {
		bs = make([][]byte, len(b)/width)
		idx := 0
		for i := 0; i < len(b)/width; i++ {
			bs[i] = b[idx : idx+width]
			idx += width
		}
	}

	s.Variables().Set(hTarget, types.NewFTBytearrayArrayOrPanic(width, bs...))

	return fmt.Sprintf("array b%v %v", width, hTarget), nil
}

func Deserialise(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hValues := variables.Handle(args[1])

	target, err := s.Variables().Get(hTarget)
	if err != nil {
		return "", err
	}
	t := target.TypeCode()

	values, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), hValues)
	if err != nil {
		return "", err
	}

	value, err := values.GetBinaryArray(-1)
	if err != nil {
		return "", err
	}

	v, err := types.FromBytes(t, value)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, v)

	return Ack, nil
}

func EdFolded(s SegmentHost, args []string) (string, error) {
	if !s.IsGPUAvailable() {
		return "", ErrEd25519Unavailable
	}

	hTarget := variables.Handle(args[0])
	hSource := variables.Handle(args[1])

	source, err := variables.GetAs[*types.Ed25519Array](s.Variables(), hSource)
	if err != nil {
		return "", err
	}

	bs, width := source.ToFoldedBytes()

	result, err := types.NewFTBytearrayArrayFromBytes(bs, width)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, result)

	return fmt.Sprintf("array b%v %v", width, hTarget), nil
}

func EdAffine(s SegmentHost, args []string) (string, error) {
	if !s.IsGPUAvailable() {
		return "", ErrEd25519Unavailable
	}

	hTarget := variables.Handle(args[0])
	hSource := variables.Handle(args[1])

	source, err := variables.GetAs[*types.Ed25519Array](s.Variables(), hSource)
	if err != nil {
		return "", err
	}

	bs, width := source.ToAffineBytes()

	result, err := types.NewFTBytearrayArrayFromBytes(bs, width)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, result)

	return fmt.Sprintf("array b%v %v", width, hTarget), nil
}

func EdFoldedProject(s SegmentHost, args []string) (string, error) {
	if !s.IsGPUAvailable() {
		return "", ErrEd25519Unavailable
	}

	hTarget := variables.Handle(args[0])
	hSource := variables.Handle(args[1])

	source, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), hSource)
	if err != nil {
		return "", err
	}

	target, err := types.NewEd25519ArrayFromFoldedBytearrayArray(source)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, target)

	return fmt.Sprintf("array E %v", hTarget), nil
}

func EdAffineProject(s SegmentHost, args []string) (string, error) {
	if !s.IsGPUAvailable() {
		return "", ErrEd25519Unavailable
	}

	hTarget := variables.Handle(args[0])
	hSource := variables.Handle(args[1])

	source, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), hSource)
	if err != nil {
		return "", err
	}

	target, err := types.NewEd25519ArrayFromAffineBytearrayArray(source)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, target)

	return fmt.Sprintf("array E %v", hTarget), nil
}
