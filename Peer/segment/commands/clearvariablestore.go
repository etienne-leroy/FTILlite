// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package commands

const CommandClearVariableStore = "command_clearvariablestore" // command_clearvariablestore

func ClearVariableStore(s SegmentHost, args []string) (string, error) {
	s.Variables().Clear()

	return Ack, nil
}
