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

const CommandDel = "command_del"         // command_del <hTarget :: Handle→any>
const CommandCleanup = "command_cleanup" // command_cleanup

func Del(s SegmentHost, args []string) (string, error) {
	for _, hStr := range args {
		h := variables.Handle(hStr)

		s.Variables().Delete(h)
	}

	return Ack, nil
}

func Cleanup(s SegmentHost, args []string) (string, error) {
	for _, h := range args[1:] {
		s.Variables().Delete(variables.Handle(h))
	}
	return Ack, nil
}
