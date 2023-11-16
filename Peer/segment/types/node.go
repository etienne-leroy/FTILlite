// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package types

import (
	"log"
	"strconv"
)

type Node struct {
	NodeIDString string
	Name         string
	Address      string
	Port         int
}

func (n *Node) NodeID() int64 {
	nodeID, err := strconv.Atoi(n.NodeIDString)
	if err != nil {
		log.Panicf("Node ID must be an integer, got: %v", n.NodeIDString)
	}
	return int64(nodeID)
}
