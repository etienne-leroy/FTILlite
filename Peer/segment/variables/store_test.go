// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package variables

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/stretchr/testify/assert"
)

type MyTypeVal struct {
	s string
	r int
}

func NewMyTypeVal(s string) *MyTypeVal {
	return &MyTypeVal{s, 0}
}
func (m *MyTypeVal) TypeCode() types.TypeCode { return "_" }
func (m *MyTypeVal) Equals(other types.TypeVal) bool {
	mOther, ok := other.(*MyTypeVal)
	return ok && m.s == mOther.s
}
func (m *MyTypeVal) EstimatedSize() int64 {
	return 0
}
func (m *MyTypeVal) Name() string {
	return "MyTypeVal"
}
func (m *MyTypeVal) DebugString() string {
	return "MyTypeVal"
}

func (m *MyTypeVal) GetBinaryArray(index int) ([]byte, error) { return []byte(m.s)[index:], nil }
func (m *MyTypeVal) ReferenceCount() int                      { return m.r }
func (m *MyTypeVal) IncrementReferenceCount()                 { m.r++ }
func (m *MyTypeVal) Free() {
	if m.r > 1 {
		m.r--
		return
	}
	m.r = 0
}

func TestReferenceCounting(t *testing.T) {
	s := NewStore()

	s.Set("1", NewMyTypeVal("xyz"))

	v1, _ := s.Get("1")
	v1F := v1.(types.Freer)

	if v1F.ReferenceCount() != 1 {
		t.Fatal("variable should have reference count of 1")
	}

	s.Set("2", v1)

	if v1F.ReferenceCount() != 2 {
		t.Fatal("variable should have reference count of 2")
	}

	s.Delete("1")
	if s.Exists("1") {
		t.Fatal("variable '1' should not exist")
	}

	if v1F.ReferenceCount() != 1 {
		t.Fatal("variable should have reference count of 1")
	}

	s.Delete("2")
	if s.Exists("2") {
		t.Fatal("variable '2' should not exist")
	}

	if v1F.ReferenceCount() != 0 {
		t.Fatal("variable should have reference count of 0")
	}
}

/*
This test attempts to test the thread safety of the variable store.

It creates up to `vs` variables, and then shuffles them around in `grs` Go routines. When all
the Go routines are finished, it checks that each value has the right reference count for the number of
variables pointing to it. Values no longer referenced by a variable are also checked to make sure their
reference count is 0.

This checks that `grs` can read/write to the variable store in parallel, without issue, and that the reference
counts are always updated.
*/
func TestThreadSafety(t *testing.T) {
	s := NewStore().(*store)

	grs := 10000
	vs := 100

	wg := &sync.WaitGroup{}
	startWg := &sync.WaitGroup{}
	startWg.Add(1)

	f := func(from Handle, to Handle) {
		defer wg.Done()
		startWg.Wait()

		vFrom, err := s.Get(from)
		if err != nil {
			panic(err)
		}

		s.Set(to, vFrom)
	}

	allVars := make([]*MyTypeVal, 0, vs)

	for i := 0; i < grs; i++ {
		hFrom := Handle(strconv.Itoa(rand.Intn(vs)))
		hTo := Handle(strconv.Itoa(rand.Intn(vs)))

		if !s.Exists(hFrom) {
			v := NewMyTypeVal(string(hFrom))
			allVars = append(allVars, v)
			s.Set(hFrom, v)
		}

		wg.Add(1)
		go f(hFrom, hTo)
	}

	startWg.Done()
	wg.Wait()

	refCounts := make(map[string]int)

	for _, v := range s.variables {
		m := v.(*MyTypeVal)

		count, ok := refCounts[m.s]
		if ok {
			refCounts[m.s] = count + 1
		} else {
			refCounts[m.s] = 1
		}
	}

	for _, v := range s.variables {
		m := v.(*MyTypeVal)

		count := refCounts[m.s]
		if m.r != count {
			t.Errorf("value %v did not have the right reference count. Expected %v, got %v", m.s, m.r, count)
		}
	}

	for _, v := range allVars {
		if _, inStore := refCounts[v.s]; !inStore && v.r != 0 {
			t.Errorf("value no longer in variable store did not have a reference count of 0")
		}
	}
}

func TestClearStore(t *testing.T) {
	s := NewStore().(*store)

	for i := 0; i < 10; i++ {
		iString := strconv.Itoa(i)
		s.Set(Handle(iString), NewMyTypeVal(iString))
	}

	assert.Equal(t, len(s.variables), 10)

	s.Clear()
	assert.Equal(t, len(s.variables), 0)
}
