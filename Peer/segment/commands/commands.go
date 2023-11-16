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
	"errors"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

type SegmentHost interface {
	Node() *types.Node
	SetPeerAddress(nodeID string, addr string)
	GetPeerAddress(nodeID string) string
	RequestTransferBytes(nodeAddress string, handle string, newHandle string, dtype string, opcode string) error

	Register(name string, f CommandFunc)

	Variables() variables.Store

	Log(format string, v ...any)

	IsGPUAvailable() bool

	DeleteFromPickleTable(destination string) error
	LoadFromPickleTable(h variables.Handle) ([]variables.Pickle, error)
	SaveToPickleTable(t types.TypeCode, h variables.Handle, opcode string, elementindex int, b []byte) error

	SaveDestination() string
	SetSaveDestination(v string)
	LoadDestination() string
	SetLoadDestination(v string)

	GetTimingInformation() []Timing
	ClearTimingInformation()

	DBType() string
	DBConnectionString() string
}

const Ack = "ack"

var ErrEd25519Unavailable = errors.New("ed25519 arrays are not available as GPU support is disabled")

type CommandFunc func(s SegmentHost, args []string) (string, error)

func RegisterCommands(s SegmentHost) {
	s.Register(CommandInit, Init)
	s.Register(CommandNetInit, NetInit)
	s.Register(CommandLogMessage, LogMessage)
	s.Register(CommandLogVariable, LogVariable)
	s.Register(CommandLogStats, LogStats)
	s.Register(CommandDel, Del)
	s.Register(CommandCleanup, Cleanup)
	s.Register(CommandClearVariableStore, ClearVariableStore)

	s.Register(CommandTransmit, Transmit)

	s.Register(CommandStartSave, StartSave)
	s.Register(CommandSave, Save)
	s.Register(CommandFinishSave, FinishSave)
	s.Register(CommandStartLoad, StartLoad)
	s.Register(CommandLoad, Load)
	s.Register(CommandFinishLoad, FinishLoad)
	s.Register(CommandDelFile, Delfile)

	s.Register(CommandSerialise, Serialise)
	s.Register(CommandDeserialise, Deserialise)
	s.Register(CommandEdFolded, EdFolded)
	s.Register(CommandEdAffine, EdAffine)
	s.Register(CommandEdFoldedProject, EdFoldedProject)
	s.Register(CommandEdAffineProject, EdAffineProject)

	s.Register(CommandAuxDbRead, AuxDBRead)
	s.Register(CommandAuxDbWrite, AuxDBWrite)

	s.Register(CommandNewilist, Newilist)
	s.Register(CommandNewflist, Newflist)
	s.Register(CommandNewArray, NewArray)
	s.Register(CommandArange, Arange)
	s.Register(CommandRandomArray, RandomArray)
	s.Register(CommandRandomPerm, RandomPerm)
	s.Register(CommandConcat, Concat)
	s.Register(CommandByteProject, ByteProject)

	s.Register(CommandLen, Len)
	s.Register(CommandSetLength, SetLength)
	s.Register(CommandCalcBroadcastLength, CalcBroadcastLength)
	s.Register(CommandBroadcastValue, BroadcastValue)
	s.Register(CommandSliceToIndices, SliceToIndices)
	s.Register(CommandAsType, AsType)

	s.Register(CommandPyLen, PyLen)
	s.Register(CommandToList, ToPythonList)

	s.Register(CommandGetItem, GetItem)
	s.Register(CommandLookup, Lookup)
	s.Register(CommandMux, Mux)
	s.Register(CommandSetItem, SetItem)
	s.Register(CommandDelItem, DelItem)
	s.Register(CommandIndex, Index)
	s.Register(CommandVerify, Verify)
	s.Register(CommandEqualInt, EqualInt)
	s.Register(CommandNonZero, Verify)
	s.Register(CommandContains, Contains)
	s.Register(CommandReduceSum, ReduceSum)
	s.Register(CommandReduceISum, ReduceISum)
	s.Register(CommandReduceMin, ReduceMin)
	s.Register(CommandReduceIMin, ReduceIMin)
	s.Register(CommandReduceMax, ReduceMax)
	s.Register(CommandReduceIMax, ReduceIMax)
	s.Register(CommandCumSum, CumSum)
	s.Register(CommandSorted, Sorted)
	s.Register(CommandIndexSorted, IndexSorted)

	s.Register(CommandSHA3256, Sha3_256)
	s.Register(CommandAES256Encrypt, Aes256Encrypt)
	s.Register(CommandAES256Decrypt, Aes256Decrypt)
	s.Register(CommandGrain128aeadv2, Grain128Aeadv2)
	s.Register(CommandECDSA256Keygen, ECDSA256Keygen)
	s.Register(CommandECDSA256PublicKey, ECDSA256PublicKey)
	s.Register(CommandECDSA256Sign, ECDSA256Sign)
	s.Register(CommandECDSA256Verify, ECDSA256Verify)
	s.Register(CommandRSA3072Keygen, RSA3072Keygen)
	s.Register(CommandRSA3072PublicKey, RSA3072PublicKey)
	s.Register(CommandRSA3072Encrypt, RSA3072Encrypt)
	s.Register(CommandRSA3072Decrypt, RSA3072Decrypt)

	s.Register(CommandNewListmap, NewListmap)
	s.Register(CommandListmapKeys, ListmapGetKeys)
	s.Register(CommandListmapGetItem, ListmapGetItem)
	s.Register(CommandListmapContains, ListmapContains)
	s.Register(CommandListmapAddItem, ListmapAddItem)
	s.Register(CommandListmapRemoveItem, ListmapRemoveItem)
	s.Register(CommandListmapIntersectItem, ListmapIntersectItem)
	s.Register(CommandListmapKeysUnique, ListmapKeysUnique)
	s.Register(CommandListmapSetItems, ListmapSetItem)
	s.Register(CommandListmapCopy, ListmapCopy)

	s.Register(CommandEq, Eq)
	s.Register(CommandNe, Ne)
	s.Register(CommandGt, Gt)
	s.Register(CommandLt, Lt)
	s.Register(CommandGe, Ge)
	s.Register(CommandLe, Le)

	s.Register(CommandNeg, Neg)
	s.Register(CommandAbs, Abs)

	s.Register(CommandFloor, Floor)
	s.Register(CommandCeil, Ceil)
	s.Register(CommandRound, Round)
	s.Register(CommandTrunc, Trunc)

	s.Register(CommandAdd, Add)
	s.Register(CommandSub, Sub)
	s.Register(CommandMul, Mul)

	s.Register(CommandFloorDiv, FloorDiv)
	s.Register(CommandTrueDiv, TrueDiv)
	s.Register(CommandMod, Mod)
	s.Register(CommandDivMod, DivMod)
	s.Register(CommandPow, Pow)

	s.Register(CommandLShift, LShift)
	s.Register(CommandRShift, RShift)
	s.Register(CommandAnd, And)
	s.Register(CommandOr, Or)
	s.Register(CommandXor, Xor)
	s.Register(CommandInvert, Invert)

	s.Register(CommandNearest, Nearest)
	s.Register(CommandExp, Exp)
	s.Register(CommandLog, Log)
	s.Register(CommandSin, Sin)
	s.Register(CommandCos, Cos)
}
