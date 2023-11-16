// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package segment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/streadway/amqp"

	"github.com/AUSTRAC/ftillite/Peer/segment/commands"
	fthttp "github.com/AUSTRAC/ftillite/Peer/segment/net/http"
	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

type Segment struct {
	node             *types.Node
	rabbitMQAddr     string
	rabbitMQIncoming string
	rabbitMQOutgoing string
	dbType           string
	dbConnStr        string
	dbChunkSize      int
	gpuEnabled       bool

	inSession       bool
	variables       variables.Store
	commands        map[string]commands.CommandFunc
	httpListener    net.Listener
	httpServer      *http.Server
	segmentNodes    map[string]string
	segmentClient   *fthttp.SegmentClient
	saveDestination string
	loadDestination string

	timings []commands.Timing
}

type CommandTiming struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
}

var GPUHasBeenInitialized = false

var ErrEd25519Unavailable = errors.New("the Ed25519 is not available as GPU is not enabled")

func NewSegment(options Options, dbType string, dbConnStr string) (*Segment, error) {
	if options.EnableGPU && !GPUHasBeenInitialized {
		err := types.InitializeGPU()
		if err != nil {
			return nil, err
		}
		GPUHasBeenInitialized = true
	}

	incomingQueue := fmt.Sprintf("%v%v", options.RabbitMQIncomingPrefix, options.NodeIDString)
	outgoingQueue := fmt.Sprintf("%v%v", options.RabbitMQOutgoingPrefix, options.NodeIDString)

	httpListener, err := net.Listen("tcp", options.Address)
	if err != nil {
		return nil, fmt.Errorf("unable to listen for incoming connections: %w", err)
	}

	node := types.Node{
		NodeIDString: options.NodeIDString,
		Name:         options.NodeName,
		Address:      options.ExternalAddress,
		Port:         options.ExternalPort,
	}

	segmentClient := fthttp.NewSegmentClient(int(node.NodeID()))

	segment := &Segment{
		&node,
		options.RabbitMQAddr,
		incomingQueue,
		outgoingQueue,
		dbType,
		dbConnStr,
		options.DbChunkSize,
		options.EnableGPU,
		false,
		variables.NewStore(),
		make(map[string]commands.CommandFunc),
		httpListener,
		nil,
		make(map[string]string),
		segmentClient,
		"",
		"",
		make([]commands.Timing, 0),
	}

	httpServer := &http.Server{
		Handler: fthttp.NewAnonHandler(
			fthttp.SegmentEndpoints(segment),
		),
	}

	segment.httpServer = httpServer

	log.Printf("RabbitMQ Address: %s", options.RabbitMQAddr)

	commands.RegisterCommands(segment)

	return segment, nil
}

func (s *Segment) StartHTTPServer() bool {
	if err := s.httpServer.Serve(s.httpListener); err != nil {
		if err == http.ErrServerClosed {
			return true
		}
	}
	return false
}

func (s *Segment) Log(format string, args ...any) {
	log.Printf(format, args...)
}
func (s *Segment) IsGPUAvailable() bool {
	return s.gpuEnabled
}
func (s *Segment) Variables() variables.Store {
	return s.variables
}
func (s *Segment) GetTimingInformation() []commands.Timing {
	return s.timings
}
func (s *Segment) ClearTimingInformation() {
	s.timings = make([]commands.Timing, 0)
}
func (s *Segment) Node() *types.Node {
	return s.node
}
func (s *Segment) SetPeerAddress(nodeID string, addr string) {
	s.segmentNodes[nodeID] = addr
}
func (s *Segment) GetPeerAddress(nodeID string) string {
	return s.segmentNodes[nodeID]
}
func (s *Segment) SaveDestination() string {
	return s.saveDestination
}
func (s *Segment) SetSaveDestination(v string) {
	s.saveDestination = v
}
func (s *Segment) LoadDestination() string {
	return s.loadDestination
}
func (s *Segment) SetLoadDestination(v string) {
	s.loadDestination = v
}
func (s *Segment) DBType() string             { return s.dbType }
func (s *Segment) DBConnectionString() string { return s.dbConnStr }

func (s *Segment) Listen() error {
	log.Printf("RabbitMQ Address: %s", s.rabbitMQAddr)

	var conn *amqp.Connection
	var err error

	for {
		conn, err = amqp.Dial(s.rabbitMQAddr)
		if err == nil {
			break
		}

		log.Printf("Error connecting to Rabbit MQ: %v\n", err)
		log.Println("Waiting 5 seconds before attempting to connect again.")
		time.Sleep(5 * time.Second)
	}
	log.Println("Connected to the Rabbit MQ server.")

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		s.rabbitMQIncoming, // name
		false,              // durable
		true,               // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	for d := range msgs {
		name, args, responseRequired, err := parseMessage(d.Body)
		if err != nil {
			log.Printf("Unable to parse incoming message: %v\nBody: %s\n", err, d.Body)
			continue
		}

		resp := s.RunCommandWithLogging(name, args, responseRequired)

		if responseRequired {
			err = ch.Publish("", s.rabbitMQOutgoing, false, false, amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: d.CorrelationId,
				Body:          []byte(resp),
			})

			if err != nil {
				log.Printf("Unable to publish response: %v", err)
			}
		}
	}

	return nil
}

func parseMessage(b []byte) (name string, args []string, responseRequired bool, err error) {
	var req map[string]string
	err = json.Unmarshal(b, &req)
	if err != nil {
		return "", nil, false, err
	}

	// Handles auxdb_read and auxdb_write params that have string values surrounded with '__'
	x1 := strings.Split(req["command"], "__")

	// Splits on the non-string params
	x2 := strings.Split(strings.TrimSpace(x1[0]), " ")
	name = x2[0]

	// Checks if a response is required, or if command can be executed silently
	responseRequired, err = strconv.ParseBool(req["response_required"])

	if err != nil {
		return "", nil, false, err
	}

	if len(x1) > 1 {
		args = append(x2[1:], x1[1])
	} else {
		args = x2[1:]
	}

	return name, args, responseRequired, nil
}

func (s *Segment) argsLogString(args []string) string {
	argsStrings := make([]string, len(args))

	for i, arg := range args {
		if v, err := s.variables.Get(variables.Handle(arg)); err == nil && arg != "0" {
			argsStrings[i] = arg + "::" + v.DebugString()
		} else {
			argsStrings[i] = arg
		}
	}

	return strings.Join(argsStrings, " ")
}
func (s *Segment) memLogString() string {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var gpuMemStats types.GPUMemoryStats

	if s.gpuEnabled {
		gpuMemStats, _ = types.GetGPUMemoryStats()
	}

	return fmt.Sprintf(
		"Sys: %v, Heap: %v, GPU: %v/%v",
		types.PrintSize(memStats.Sys),
		types.PrintSize(memStats.Alloc),
		types.PrintSize(gpuMemStats.Used()),
		types.PrintSize(gpuMemStats.Total))
}

func (s *Segment) RunCommandWithLogging(name string, args []string, responseRequired bool) (resp string) {
	start := time.Now()

	defer func() {
		t := commands.Timing{
			Name:      name,
			StartTime: start,
			EndTime:   time.Now(),
		}
		if name != commands.CommandLogStats {
			s.timings = append(s.timings, t)
		}

		elapsed := time.Since(start)

		rr := ""
		if !responseRequired {
			rr = " (response not sent)"
		}

		log.Printf("+%vμs %v(%v) -> %v%v - %v\n", elapsed.Microseconds(), name, s.argsLogString(args), resp, rr, s.memLogString())
	}()

	log.Printf("> %v(%v) - %v",
		name,
		s.argsLogString(args),
		s.memLogString(),
	)

	var err error
	resp, err = s.RunCommand(name, args)
	if err != nil {
		resp = fmt.Sprintf("error %v", err.Error())
	}

	return resp
}

func (s *Segment) RunCommand(name string, args []string) (resp string, err error) {
	if !strings.HasPrefix(name, "command_") {
		name = fmt.Sprintf("command_%v", name)
	}

	if name == "command_error" {
		panic("purposely crash")
	}

	f, ok := s.commands[name]
	if !ok {
		return "", fmt.Errorf("unknown command '%v'", name)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered: %v - Stack trace:\n%v\n", r, string(debug.Stack()))
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("unknown panic")
			}
			resp = ""
		}
	}()

	resp, err = f(s, args)

	if err != nil {
		return "", err
	}

	return resp, nil
}

func (s *Segment) Register(name string, f commands.CommandFunc) {
	s.commands[name] = f
}

func (s *Segment) GetVariable(h variables.Handle) (types.TypeVal, error) {
	return s.variables.Get(h)
}

func (s *Segment) SetVariable(h variables.Handle, value types.TypeVal) {
	s.variables.Set(h, value)
}

func (s *Segment) DeleteVariable(h variables.Handle) {
	s.variables.Delete(h)
}

func (s *Segment) RequestTransferBytes(nodeAddress string, handle string, newHandle string, dtype string, opcode string) error {
	return s.segmentClient.RequestTransmission(nodeAddress, handle, newHandle, dtype, opcode)
}

func (s *Segment) TransferBytes(nodeAddress string, handle string, newHandle string, dtype string, opcode string) error {
	// Read the data and assign to variable

	// if dtype indicates listmap need to break into multiple requests
	t, err := types.ParseTypeCode(dtype)
	if err != nil {
		return err
	}

	typeCodes := t.GetTypeCodeAsSlice()

	var lmArray = make([]types.ArrayTypeVal, len(typeCodes))

	isListMap := opcode == "listmap"

	var v types.TypeVal

	for index, tc := range typeCodes {
		var writer bytes.Buffer
		arraylength, err := s.segmentClient.ReceiveTransmission(nodeAddress, handle, tc, index, &writer)
		if err != nil {
			return err
		}

		arraySize := len(writer.Bytes())
		rawData := make([]byte, arraySize)
		_, err = writer.Read(rawData)
		if err != nil {
			return err
		}

		if tc.GetBase() == types.BytearrayB && tc.Length() != arraylength {
			panic("ReceiveTransmission has a different BytearrayArray width than the type code")
		}

		xs, err := types.FromBytes(tc, rawData)
		if err != nil {
			return err
		}

		lmArray[index] = xs
	}

	if !isListMap {
		v = lmArray[0]
	} else {
		lmGoArray := make([]types.ArrayElementTypeVal, len(lmArray))

		for i, xs := range lmArray {
			var ok bool
			lmGoArray[i], ok = xs.(types.ArrayElementTypeVal)
			if !ok {
				return errors.New("only Integer, Float, Bytearray and Ed25519Int arrays can be used in listmaps")
			}
		}

		v, err = types.NewListMapFromArrays(typeCodes, lmGoArray, "any")
		if err != nil {
			return err
		}
	}

	s.SetVariable(variables.Handle(newHandle), v)

	return nil
}

func (s *Segment) SegmentNodes() map[string]string {
	return s.segmentNodes
}
