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
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"testing"

	"github.com/AUSTRAC/ftillite/Peer/segment/commands"
	"github.com/AUSTRAC/ftillite/Peer/segment/types"
)

func BenchmarkEd25519ReduceSum(b *testing.B) {
	types.SkipIfGPUUnavailable(b)

	log.SetOutput(io.Discard)

	for indexFactor := int64(7); indexFactor > 0; indexFactor-- {
		for s := 1; s <= 10; s++ {
			size := int64(2_000_000 * s)

			b.Run(fmt.Sprintf("Ed25519::ReduceSum(size=%v,indexes=%v)", size, size*indexFactor), func(b *testing.B) {
				b.StopTimer()

				s := NewTestSegment()
				defer s.variables.Clear()

				// Indexes
				idxs := make([]int64, size*indexFactor)
				for i := 1; i <= int(indexFactor); i++ {
					for j := int64(0); j < size; j++ {
						idxs[i*int(j)] = rand.Int63n(size + 1)
					}
				}
				s.SetVariable("4", types.NewFTIntegerArray(idxs...))

				// Values
				s.SetVariable("5", types.NewFTIntegerArray(size*indexFactor))
				AssertCommand(b, s, commands.CommandRandomArray, "6", "I", "5")
				AssertCommand(b, s, commands.CommandAsType, "7", "6", "E")

				b.StartTimer()
				for n := 0; n < b.N; n++ {
					t := strconv.Itoa(n + 8)

					// Ed25519 source array
					s.SetVariable("1", types.NewFTIntegerArray(size))
					AssertCommand(b, s, commands.CommandRandomArray, "2", "I", "1")
					AssertCommand(b, s, commands.CommandAsType, t, "2", "E")

					AssertCommand(b, s, commands.CommandReduceSum, t, "7", "4")

					x, err := types.GetGPUMemoryStats()
					if err == nil {
						b.ReportMetric(float64(x.Used()), "gpumem")
					}

					AssertCommand(b, s, commands.CommandDel, t)
				}
			})
		}

	}
}
