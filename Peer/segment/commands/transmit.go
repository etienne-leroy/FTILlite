// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package commands

import "github.com/AUSTRAC/ftillite/Peer/segment/variables"

const CommandTransmit = "command_transmit" // command_transmit

func Transmit(s SegmentHost, args []string) (string, error) {
	newHandle := args[1]
	handle := args[2]
	targetNode := args[0]
	nodeAddress := s.GetPeerAddress(targetNode)
	dtype := args[3]
	opcode := args[4]

	if s.Node().NodeIDString == targetNode {
		v, err := s.Variables().Get(variables.Handle(handle))
		if err != nil {
			return "", err
		}

		s.Variables().Set(variables.Handle(newHandle), v)

	} else {
		err := s.RequestTransferBytes(nodeAddress, handle, newHandle, dtype, opcode)
		if err != nil {
			return "Error", err
		}
	}

	return "ack", nil
}
