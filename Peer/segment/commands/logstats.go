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
	"fmt"
	"log"
	"runtime"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"
)

const CommandLogStats = "command_log_stats"

type Timing struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
}

func (t Timing) Duration() time.Duration {
	return t.EndTime.Sub(t.StartTime)
}

type CommandTimingAggregation struct {
	Name               string
	Count              int64
	TotalElapsedTime   int64
	AverageElapsedTime int64
	PercentTotalTime   float64
}

func LogStats(s SegmentHost, args []string) (string, error) {
	totalCount, totalSize, typeTotals := s.Variables().Stats()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	writer := tabwriter.NewWriter(log.Writer(), 0, 8, 1, '\t', 0)
	fmt.Fprintln(writer)
	fmt.Fprintf(writer, "Process memory usage:\t%d\n", memStats.Alloc)
	fmt.Fprintf(writer, "Process GC count:\t%d\n", memStats.NumGC)
	writer.Flush()

	writer = tabwriter.NewWriter(log.Writer(), 0, 8, 1, '\t', 0)
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "Variables (now, point-in-time):")
	fmt.Fprintf(writer, "Type\tCount\tMemory usage (bytes)\n")

	for _, v := range typeTotals {
		fmt.Fprintf(writer, "%v\t%d\t%d\n", v.Type, v.Count, v.EstimatedSize)
	}
	fmt.Fprintf(writer, "\t%v\t%v\n", totalCount, totalSize)

	writer.Flush()

	m := make(map[string]*CommandTimingAggregation)

	count := 0
	totalElapsed := int64(0)

	for _, t := range s.GetTimingInformation() {
		d := t.Duration().Microseconds()
		count++
		totalElapsed += d

		if v, ok := m[t.Name]; ok {
			v.Count++
			v.TotalElapsedTime += d
			v.AverageElapsedTime = v.TotalElapsedTime / v.Count
		} else {
			m[t.Name] = &CommandTimingAggregation{
				Name:               t.Name,
				Count:              1,
				TotalElapsedTime:   d,
				AverageElapsedTime: d,
			}
		}
	}

	ts := make([]*CommandTimingAggregation, 0, len(m))

	for _, v := range m {
		v.PercentTotalTime = (float64(v.TotalElapsedTime) / float64(totalElapsed)) * 100
		ts = append(ts, v)
	}

	sort.Slice(ts, func(i, j int) bool {
		return ts[i].TotalElapsedTime > ts[j].TotalElapsedTime
	})

	writer = tabwriter.NewWriter(log.Writer(), 0, 8, 1, '\t', 0)
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "Commands (cumulative):")
	fmt.Fprintln(writer, "Command\tCount\tAverage Time (μs)\tTotal Time (μs)\tPercentage (%)")
	for _, v := range ts {
		fmt.Fprintf(writer, "%v\t%v\t%v\t%v\t%.5f\n", v.Name, v.Count, v.AverageElapsedTime, v.TotalElapsedTime, v.PercentTotalTime)
	}

	fmt.Fprintf(writer, "\t%v\t-\t%v\t100\n", count, totalElapsed)
	writer.Flush()

	if len(args) > 0 {
		b, err := strconv.ParseBool(args[0])
		if err == nil && b {
			s.ClearTimingInformation()
		}
	}

	return Ack, nil
}
