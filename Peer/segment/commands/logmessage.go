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
	"strings"
)

const CommandLogMessage = "command_log_message"

func LogMessage(s SegmentHost, args []string) (string, error) {
	s.Log("%v", strings.Join(args, " "))
	return Ack, nil
}
