// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package segment

type Options struct {
	NodeIDString           string
	NodeName               string
	RabbitMQAddr           string
	RabbitMQIncomingPrefix string
	RabbitMQOutgoingPrefix string
	Address                string
	ExternalAddress        string
	ExternalPort           int
	ExternalFQDN           string
	EnableGPU              bool
	DbChunkSize            int
}
