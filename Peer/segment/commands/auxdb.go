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
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const CommandAuxDbRead = "command_auxdb_read"   // command_auxdb_read
const CommandAuxDbWrite = "command_auxdb_write" // command_auxdb_write

func AuxDBRead(s SegmentHost, args []string) (string, error) {
	var res []string
	invalidTypeErr := fmt.Errorf("type not supported for operation auxdb_read")
	db, err := sql.Open(s.DBType(), s.DBConnectionString())
	if err != nil {
		return "", err
	}
	defer db.Close()

	// Check DB Connection
	if err := db.Ping(); err != nil {
		return "", err
	}

	// Process arguments
	nCols := (len(args) - 1) / 2
	hs := args[0:(nCols)] // Handles
	query := args[len(args)-1]

	tcs := make([]types.TypeCode, nCols) // Typecodes
	for i, t := range args[(nCols):(len(args) - 1)] {
		tc, err := types.ParseTypeCode(t)
		if err != nil {
			return "", err
		}
		tcs[i] = tc
	}

	// Initialise Arrays
	dInts := make(map[string][]int64)
	dFloats := make(map[string][]float64)
	dBytearrays := make(map[string][][]byte)

	for i := 0; i < nCols; i++ {
		switch tcs[i].GetBase() {
		case types.IntegerB:
			dInts[hs[i]] = []int64{}
		case types.FloatB:
			dFloats[hs[i]] = []float64{}
		case types.BytearrayB:
			dBytearrays[hs[i]] = [][]byte{}
		default:
			return "", invalidTypeErr
		}
	}

	queryResult, err := db.Query(query)
	defer queryResult.Close()

	if err != nil {
		return "", err
	}

	for queryResult.Next() {

		row := make([]interface{}, nCols)
		for i := 0; i < nCols; i++ {
			switch tcs[i].GetBase() {
			case types.IntegerB:
				var tmp int64
				row[i] = &tmp
			case types.FloatB:
				var tmp float64
				row[i] = &tmp
			case types.BytearrayB:
				var tmp = make([]byte, tcs[i].Length())
				row[i] = &tmp
			default:
				return "", invalidTypeErr
			}
		}

		if err := queryResult.Scan(row...); err != nil {
			return "", err
		}

		for i := 0; i < nCols; i++ {
			switch tcs[i].GetBase() {
			case types.IntegerB:
				val := *(row[i].(*int64))
				dInts[hs[i]] = append(dInts[hs[i]], val)
			case types.FloatB:
				val := *(row[i].(*float64))
				dFloats[hs[i]] = append(dFloats[hs[i]], val)
			case types.BytearrayB:
				tmp := make([]byte, tcs[i].Length())
				val := *(row[i].(*[]byte))
				for i := range tmp {
					if i < len(val) {
						tmp[i] = val[i]
					} else {
						tmp[i] = 0
					}
				}
				dBytearrays[hs[i]] = append(dBytearrays[hs[i]], tmp)
			default:
				return "", invalidTypeErr
			}
		}
	}

	for i, h := range hs {
		switch tcs[i].GetBase() {
		case types.IntegerB:
			s.Variables().Set(variables.Handle(h), types.NewFTIntegerArray(dInts[h]...))
		case types.FloatB:
			s.Variables().Set(variables.Handle(h), types.NewFTFloatArray(dFloats[h]...))
		case types.BytearrayB:
			target, err := types.NewFTBytearrayArray(tcs[i].Length(), dBytearrays[h]...)
			if err != nil {
				return "", err
			}

			s.Variables().Set(variables.Handle(h), target)
		}
		res = append(res, fmt.Sprintf("array %s %s", tcs[i], h))
	}

	if len(tcs) > 0 {
		return strings.Join(res, " "), nil
	}
	return "ack", nil
}

func AuxDBWrite(s SegmentHost, args []string) (string, error) {
	tableName := args[0]
	// TODO: add checks for args, including if number of col names == number of handles
	noCols := (len(args) - 1) / 2
	colNames := args[1 : noCols+1]
	hTargetStrs := args[noCols+1:]

	hTargets := make([]variables.Handle, noCols)
	for i, h := range hTargetStrs {
		hTargets[i] = variables.Handle(h)
	}

	db, err := sql.Open(s.DBType(), s.DBConnectionString())
	if err != nil {
		return "", err
	}
	defer db.Close()

	vals := make([][]interface{}, noCols)
	for i := range vals {
		v, err := s.Variables().Get(hTargets[i])
		if err != nil {
			return "", err
		}

		switch v.TypeCode().GetBase() {
		case types.IntegerB:
			tmp := v.(*types.FTIntegerArray)
			vals[i] = make([]interface{}, len(tmp.Values()))
			for j, t := range tmp.Values() {
				vals[i][j] = t
			}
		case types.FloatB:
			tmp := v.(*types.FTFloatArray)
			vals[i] = make([]interface{}, len(tmp.Values()))
			for j, t := range tmp.Values() {
				vals[i][j] = t
			}
		case types.BytearrayB:
			tmp := v.(*types.FTBytearrayArray)
			vals[i] = make([]interface{}, len(tmp.Values()))
			for j, t := range tmp.Values() {
				vals[i][j] = t
			}
		default:
			return "", fmt.Errorf("type not supported for auxdb_write")
		}
	}

	// Check lengths are even amongst arrays
	valLen := -1
	for i := range vals {
		if i == 0 {
			valLen = len(vals[i])
		} else if len(vals[i]) != valLen {
			return "", fmt.Errorf("input arrays are not of the same dimensions")
		}
	}

	for i := range vals[0] {
		colNamesSQL := strings.Join(colNames, ", ")
		paramsSQLTmp := make([]string, noCols)
		for i := range paramsSQLTmp {
			paramsSQLTmp[i] = fmt.Sprintf("$%d", i+1)
		}
		paramsSQL := strings.Join(paramsSQLTmp, ", ")

		sqlStatement := fmt.Sprintf(`
		INSERT INTO %s (%s)
		VALUES (%s)`, tableName, colNamesSQL, paramsSQL)

		fieldVals := make([]interface{}, noCols)
		for j := 0; j < noCols; j++ {
			fieldVals[j] = vals[j][i]
		}

		_, err = db.Exec(sqlStatement, fieldVals...)
		if err != nil {
			//	panic(err)
			return "", err
		}
	}
	return Ack, nil
}
