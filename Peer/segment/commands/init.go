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
	"strings"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
)

const (
	CommandInit    = "command_init"    // command_init
	CommandNetInit = "command_netinit" // command_netinit
)

func Init(s SegmentHost, args []string) (string, error) {
	s.ClearTimingInformation()

	s.Variables().Set("0", types.NewFTIntegerArray(s.Node().NodeID()))

	var gpuEnabled string
	if s.IsGPUAvailable() {
		gpuEnabled = "gpu"
	} else {
		gpuEnabled = "no_gpu"
	}

	return fmt.Sprintf("%s %s %v %s", s.Node().NodeIDString, s.Node().Name, s.Node().Address, gpuEnabled), nil
}

func NetInit(s SegmentHost, args []string) (string, error) {
	for _, j := range args {
		raw := strings.Split(j, "~")
		s.SetPeerAddress(raw[0], raw[1])
	}

	return "", nil
}
