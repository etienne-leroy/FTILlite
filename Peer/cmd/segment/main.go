// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/AUSTRAC/ftillite/Peer/segment"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

var NodeIDString string = GetEnvOr("FTILITE_NODE_ID", "0")
var NodeName string = GetEnvOr("FTILITE_NODE_NAME", "PEER0")
var RabbitMQAddr string = GetEnvOr("FTILITE_MQ_ADDR", "amqp://ftillite:ftillite@localhost:5672/")
var IncomingQueuePrefix string = GetEnvOr("FTILITE_INCOMING_QUEUE_PREFIX", "FTILITE_INCOMING_") // Node ID will be appended to this string
var OutgoingQueuePrefix string = GetEnvOr("FTILITE_OUTGOING_QUEUE_PREFIX", "FTILITE_OUTGOING_") // Node ID will be appended to this string
var Address string = GetEnvOr("FTILITE_ADDRESS", "127.0.0.1:50000")
var ExternalPort, _ = strconv.Atoi(GetEnvOr("FTILITE_PORT", "50000"))
var ExternalFQDN string = GetEnvOr("FTILITE_EXTERNAL_FQDN", "localhost")
var DBType string = GetEnvOr("FTILITE_DB_TYPE", "postgres")
var DBConnStr string = GetEnvOr("FTILITE_DB_ADDR", "postgresql://postgres:postgres@localhost:5432/ftillite")
var EnableGPU string = GetEnvOr("FTILITE_ENABLE_GPU", "false")
var DbChunkSize, _ = strconv.Atoi((GetEnvOr("FTILITE_DB_CHUNKSIZE", "1000000000")))

var options = &segment.Options{
	NodeIDString:           NodeIDString,
	NodeName:               NodeName,
	RabbitMQAddr:           RabbitMQAddr,
	RabbitMQIncomingPrefix: IncomingQueuePrefix,
	RabbitMQOutgoingPrefix: OutgoingQueuePrefix,
	Address:                Address,
	ExternalAddress:        fmt.Sprintf("%v:%v", ExternalFQDN, ExternalPort),
	ExternalPort:           ExternalPort,
	ExternalFQDN:           ExternalFQDN,
	EnableGPU:              lenientParseBool(EnableGPU),
	DbChunkSize:            DbChunkSize,
}

var EnableREPL bool = false
var EnableMQ bool = true

func lenientParseBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b
}

func main() {
	if !options.EnableGPU {
		log.Println("NOTICE: GPU support has been disabled, Ed25519 arrays will be unavailable")
	}

	s, err := segment.NewSegment(*options, DBType, DBConnStr)
	if err != nil {
		fmt.Print(err)
		panic("Can't create segment.")
	}
	if EnableMQ {
		go func() {
			err := s.Listen()
			if err != nil {
				log.Panicf("unable to listen for messages: %v", err)
			}
		}()
	}

	go func() {
		result := s.StartHTTPServer()
		if !result {
			panic("Can't start HTTP Server")
		}
	}()

	if EnableREPL {
		log.Println(" Use the 'quit' command to exit.")

		r := bufio.NewReader(os.Stdin)

		quit := false
		for !quit {
			fmt.Print("FTILite> ")
			in, _ := r.ReadString('\n')

			cmd := strings.Split(strings.TrimSpace(in), " ")

			if len(cmd) < 1 {
				continue
			}

			switch cmd[0] {
			case "quit":
				quit = true
				break
			case "show":
				h := variables.Handle(cmd[1])
				v, err := s.GetVariable(h)

				if err != nil {
					fmt.Printf("Unable to show variable: %v\n", err)
					break
				}

				fmt.Printf("< %v\n", v)
				break
			default:
				result, err := s.RunCommand(cmd[0], cmd[1:])
				if err != nil {
					fmt.Printf("Unable to execute command: %v\n", err)
					continue
				}
				fmt.Printf("< %v\n", result)
				break
			}
		}
	} else {
		log.Println(" Press Ctrl-C to exit.")

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
	}
}

func GetEnvOr(name string, value string) string {
	v := os.Getenv(name)
	if len(v) == 0 {
		return value
	}
	return v
}
