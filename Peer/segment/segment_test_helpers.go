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
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

func NewTestSegment() *Segment {
	o := Options{
		NodeIDString: "0",
		EnableGPU:    (*types.EnableGPUFlag || types.EnableGPUEnv),
		DbChunkSize:  1000000000,
	}
	s, _ := NewSegment(o, "sqlite3", "file::memory:?cache=shared")
	return s
}

type Helper interface {
	Helper()
	Fatal(args ...any)
}

func AssertCommand(t Helper, s *Segment, name string, args ...string) {
	t.Helper()

	_, err := s.RunCommand(name, args)
	if err != nil {
		t.Fatal(err)
	}
}

func AssertCommandFailure(t *testing.T, s *Segment, name string, args []string, errorString string) {
	t.Helper()

	_, err := s.RunCommand(name, args)
	if err == nil {
		t.Fatal("command was meant to fail, but did not")
	}
	if errorString != "" && !strings.Contains(err.Error(), errorString) {
		t.Fatalf("error did not contain errorString: got '%s'", err)
	}
}

func AssertCommandResponse(t *testing.T, s *Segment, name string, args []string, expected string) {
	t.Helper()

	actual, err := s.RunCommand(name, args)
	if err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Fatalf("Expected: %v, Actual: %v", expected, actual)
	}
}

func AssertVariable[T types.TypeVal](t *testing.T, s *Segment, h variables.Handle) T {
	t.Helper()
	var v T
	var err error
	v, err = variables.GetAs[T](s.Variables(), h)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func AssertValue[T types.TypeVal](t *testing.T, s *Segment, h variables.Handle, v T) {
	t.Helper()
	actual := AssertVariable[T](t, s, h)
	if !v.Equals(actual) {
		t.Fatalf("variable '%v' value was %v not %v", h, actual, v)
	}
}
func AssertValueNot[T types.TypeVal](t *testing.T, s *Segment, h variables.Handle, v T) {
	t.Helper()
	actual := AssertVariable[T](t, s, h)
	if v.Equals(actual) {
		t.Fatalf("variable '%v' value was equal to %v", h, actual)
	}
}
func AssertNoVariable(t *testing.T, s *Segment, h variables.Handle) {
	t.Helper()
	if s.variables.Exists(h) {
		t.Fatalf("variable should not exist, but does: %v", h)
	}
}
func CreateTmpTable(q string) (*sql.DB, error) {
	connStr := "file::memory:?cache=shared"
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		log.Fatalf("cannot open an SQLite memory database: %v", err)
	}

	_, err = db.Exec(q)
	if err != nil {
		log.Fatalf("cannot create schema: %v", err)
	}
	return db, err
}

func StrToByteSlice(str string, size int) []byte {
	b := make([]byte, size)
	for i, s := range str {
		b[i] = byte(s)
	}
	return b
}

// PrintVariableStore will print the contents of the variable store after the test assertions have run.
// This function is useful when you have a failing test and want to see the variable store contents.
func PrintVariableStore(t *testing.T, s *Segment) {
	t.Cleanup(func() {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(w, "Name\tValue")
		s.variables.ForEach(func(k variables.Handle, v types.TypeVal) {
			fmt.Fprintf(w, "%v\t%v\n", k, v)
		})
		w.Flush()
	})
}

func IsEd25519ArrayEmpty(t *testing.T, s *Segment, h variables.Handle) bool {
	arr, err := variables.GetAs[*types.Ed25519Array](s.Variables(), h)

	if err != nil {
		t.Fatalf("get variable failure with error %s", err)
	}

	return arr.IsEmpty()
}

func AssertEd25519ArrayEmpty(t *testing.T, s *Segment, h variables.Handle) {
	if !IsEd25519ArrayEmpty(t, s, h) {
		t.Fatalf("array is not empty")
	}
}

func AssertEd25519ArrayNotEmpty(t *testing.T, s *Segment, h variables.Handle) {
	if IsEd25519ArrayEmpty(t, s, h) {
		t.Fatalf("array is empty")
	}
}
