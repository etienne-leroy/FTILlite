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
	"database/sql"
	"time"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

func (s *Segment) SaveToPickleTable(t types.TypeCode, h variables.Handle, opcode string, elementindex int, b []byte) error {
	db, err := sql.Open(s.dbType, s.dbConnStr)
	if err != nil {
		return err
	}
	defer db.Close()

	// Check DB Connection
	if err := db.Ping(); err != nil {
		return err
	}
	sqlStatement := `
		INSERT INTO pickle (destination, dtype, opcode, handle, data, elementindex, chunkindex, created)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for chunkindex, chunk := range chunkSlice(b, s.dbChunkSize) {
		_, err = tx.Exec(sqlStatement, s.saveDestination, t, opcode, h, chunk, elementindex, chunkindex, time.Now())

		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil

}

func (s *Segment) LoadFromPickleTable(h variables.Handle) ([]variables.Pickle, error) {
	db, err := sql.Open(s.dbType, s.dbConnStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Check DB Connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	sqlStatement := `SELECT destination, dtype, handle, opcode, data, elementindex, chunkindex
	FROM pickle
	WHERE destination = $1 AND handle = $2
	ORDER BY elementindex ASC, chunkindex ASC`

	var pickles []variables.Pickle
	var index int = -1
	var data []byte
	var p variables.Pickle
	var c variables.Pickle

	rows, err := db.Query(sqlStatement, s.loadDestination, h)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&p.Destination, &p.Dtype, &p.Handle, &p.Opcode, &p.Data, &p.Index, &p.Chunkindex); err != nil {
			return pickles, err
		}
		if index == -1 {
			index = p.Index
		}
		if index == p.Index {
			// append the bytes to the existing data
			data = append(data, p.Data...)
		} else {
			c.Data = data
			pickles = append(pickles, c) //add the completed
			data = p.Data
			index = p.Index
		}
		c = p
	}

	p.Data = data
	pickles = append(pickles, p)

	if err = rows.Err(); err != nil {
		return pickles, err
	}

	return pickles, nil

}

func (s *Segment) DeleteFromPickleTable(destination string) error {
	db, err := sql.Open(s.dbType, s.dbConnStr)
	if err != nil {
		return err
	}
	defer db.Close()

	// Check DB Connection
	if err := db.Ping(); err != nil {
		return err
	}

	sqlStatement := `DELETE FROM pickle
	WHERE destination = $1`

	_, err = db.Exec(sqlStatement, destination)
	if err != nil {
		return err
	}

	return nil
}

func chunkSlice(slice []byte, chunkSize int) [][]byte {
	var chunks [][]byte

	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}
