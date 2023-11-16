// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package types

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"filippo.io/edwards25519"
)

type Key struct {
	value interface{}
}

var interfaceType = reflect.TypeOf((*interface{})(nil)).Elem()

func NewKey(xs []interface{}) Key {
	arrType := reflect.ArrayOf(len(xs), interfaceType)

	arr := reflect.New(arrType).Elem()
	slc := arr.Slice(0, len(xs)).Interface().([]interface{})

	for i, x := range xs {
		switch v := x.(type) {
		case []byte:
			bArrType := reflect.ArrayOf(len(v), reflect.TypeOf(v[0]))
			bArr := reflect.New(bArrType).Elem()

			bSlc := bArr.Slice(0, len(v)).Interface().([]byte)
			copy(bSlc, v)

			slc[i] = bArr.Interface()
		default:
			slc[i] = x
		}
	}

	return Key{arr.Interface()}
}

func (k *Key) Element(i int) interface{} {
	v := reflect.ValueOf(k.value)
	return v.Index(i).Interface()
}

func (k *Key) ToSlice() []interface{} {
	arr := reflect.ValueOf(k.value)
	arrLen := arr.Len()

	slc := make([]interface{}, arrLen)

	for i := 0; i < arrLen; i++ {
		v := arr.Index(i).Interface()
		switch reflect.TypeOf(v).Kind() {
		case reflect.Array:
			s := reflect.ValueOf(arr.Index(i).Interface())
			bArr := make([]byte, s.Len())
			for i := 0; i < s.Len(); i++ {
				bArr[i] = s.Index(i).Interface().(byte)
			}
			slc[i] = bArr
		default:
			slc[i] = arr.Index(i).Interface()
		}
	}
	return slc
}

func (k *Key) Size() int {
	arr := reflect.ValueOf(k.value)
	arrLen := arr.Len()

	n := 0

	for i := 0; i < arrLen; i++ {
		v := arr.Index(i).Interface()
		switch reflect.TypeOf(v).Kind() {
		case reflect.Array:
			s := reflect.ValueOf(arr.Index(i).Interface())
			n += s.Len() // []byte
		default:
			n += int(arr.Index(i).Type().Size())
		}
	}

	return n
}

type ListMap struct {
	m  map[Key]int64
	tc []TypeCode
}

func (m *ListMap) Equals(other TypeVal) bool {
	return reflect.DeepEqual(m, other)
}

func (m *ListMap) GetBinaryArray(index int) ([]byte, error) {
	a, err := KeysToArrayTypeVals(m.GetKeys(false), m.TypeCodeAsSlice())
	if err != nil {
		return nil, err
	}
	return a[index].GetBinaryArray(-1)
}

func (m *ListMap) TypeCode() TypeCode {
	tcs := make([]string, len(m.tc))
	for i, tc := range m.tc {
		tcs[i] = string(tc)
	}
	return TypeCode(strings.Join(tcs, ""))
}
func (m *ListMap) Name() string {
	return fmt.Sprintf("ListMap(%v)", m.TypeCode())
}

func (xs *ListMap) EstimatedSize() int64 {
	if len(xs.m) == 0 {
		return 0
	}

	var k Key
	for k = range xs.m {
		break
	}

	kSize := k.Size()

	// Rough estimate, numbers of rows times the key size and the value.
	return int64(len(xs.m) * (kSize + 8))
}
func (xs *ListMap) DebugString() string {
	return fmt.Sprintf("%v(Typecode=%v,Width=%v,Length=%v,Memory=%v)", xs.TypeCode(), xs.Name(), len(xs.tc), len(xs.m), PrintSize(uint64(xs.EstimatedSize())))
}

func KeyAt(arrays []ArrayElementTypeVal, index int64) Key {
	keyComponents := make([]interface{}, len(arrays))
	for i, k := range arrays {
		keyComponents[i] = k.Element(index)
	}

	return NewKey(keyComponents)
}

func NewListMapFromArrays(typeCodes []TypeCode, keys []ArrayElementTypeVal, order string) (*ListMap, error) {

	if len(keys) != len(typeCodes) {
		return nil, fmt.Errorf("number of typecodes does not match number of key composites")
	}

	m := make(map[Key]int64)
	count := keys[0].Length()
	uniqueKeys := make([]Key, 0)

	if order != "pos" && order != "any" && order != "rnd" {
		return nil, fmt.Errorf("invalid order value %s", order)
	}

	for row := int64(0); row < count; row++ {
		k := KeyAt(keys, row)
		if _, ok := m[k]; !ok {
			if order == "rnd" {
				uniqueKeys = append(uniqueKeys, k)
			} else {
				m[k] = int64(row)
			}
		} else if order == "pos" {
			return nil, fmt.Errorf("cannot have duplicate keys for order 'pos'")
		}
	}

	if order == "rnd" {
		length := int64(len(uniqueKeys))
		p, err := NewRandomPermFTIntegerArray(length, length)
		if err != nil {
			return nil, err
		}
		permutation := p.Values()
		for i, k := range uniqueKeys {
			m[k] = int64(permutation[i])
		}
	}

	return &ListMap{m, typeCodes}, nil
}

func (m *ListMap) GetItems(keys []ArrayElementTypeVal, defaultVal interface{}) (*FTIntegerArray, error) {
	rows := keys[0].Length()

	ks := make([]Key, rows)
	for i := int64(0); i < rows; i++ {
		ks[i] = KeyAt(keys, i)
	}

	results := make([]int64, rows)
	for i, k := range ks {
		if _, ok := m.m[k]; ok {
			results[i] = m.m[k]
		} else if defaultVal != nil {
			results[i] = defaultVal.(int64)
		} else {
			return nil, fmt.Errorf("key: %v not found in listmap", k)
		}
	}

	return NewFTIntegerArray(results...), nil
}

func KeysToArrayTypeVals(keys []Key, typecodes []TypeCode) ([]ArrayElementTypeVal, error) {

	keyWidth := len(typecodes)
	keyCount := len(keys)

	result := make([]ArrayElementTypeVal, keyWidth)

	for k := 0; k < keyWidth; k++ {
		tc := typecodes[k]
		switch tc.GetBase() {
		case IntegerB:
			xs := make([]int64, keyCount)
			for r := 0; r < keyCount; r++ {
				xs[r] = keys[r].Element(k).(int64)
			}
			result[k] = NewFTIntegerArray(xs...)
		case FloatB:
			xs := make([]float64, keyCount)
			for r := 0; r < keyCount; r++ {
				xs[r] = keys[r].Element(k).(float64)
			}
			result[k] = NewFTFloatArray(xs...)
		case BytearrayB:
			xs := make([][]byte, keyCount)
			for r := 0; r < keyCount; r++ {
				s := reflect.ValueOf(keys[r].Element(k))
				bArr := make([]byte, s.Len())
				for i := 0; i < s.Len(); i++ {
					bArr[i] = s.Index(i).Interface().(byte)
				}
				xs[r] = bArr
			}
			var err error
			result[k], err = NewFTBytearrayArray(tc.Length(), xs...)
			if err != nil {
				return nil, err
			}
		case Ed25519IntB:
			xs := make([]*edwards25519.Scalar, keyCount)
			for r := 0; r < keyCount; r++ {
				tmpNonPtr := keys[r].Element(k).(edwards25519.Scalar)
				xs[r] = &tmpNonPtr
			}
			result[k] = NewFTEd25519IntArray(xs...)
		default:
			return nil, fmt.Errorf("type not supported for KeysToArrayTypeVals")
		}
	}

	return result, nil
}

func (m *ListMap) RemoveItems(keys []ArrayElementTypeVal, ignoreError bool) ([]Key, *FTIntegerArray, *FTIntegerArray, error) {

	rows := keys[0].Length()
	moveItemLen := 0
	vs := make([]int64, rows)
	for i := int64(0); i < rows; i++ {
		k := KeyAt(keys, i)
		if _, ok := m.m[k]; ok {
			vs[i] = m.m[k]
			delete(m.m, k)
			moveItemLen++
		} else if !ignoreError {
			return nil, nil, nil, fmt.Errorf("key: %v not found in listmap", k)
		}
	}

	if len(m.m) < moveItemLen {
		moveItemLen = len(m.m)
	}

	assignedIndex := 0
	movedKeys := make([]Key, moveItemLen)
	oldValues := make([]int64, moveItemLen)
	newValues := make([]int64, moveItemLen)
	replaceOverVal := int64(len(m.m))
	for key, value := range m.m {
		if value >= replaceOverVal {
			movedKeys[assignedIndex] = key
			oldValues[assignedIndex] = value
			newValues[assignedIndex] = vs[assignedIndex]
			m.m[key] = vs[assignedIndex]
			assignedIndex++
		}
		if assignedIndex == moveItemLen {
			break
		}
	}

	if assignedIndex < len(movedKeys) {
		// remove elements that haven't been assigned.
		movedKeys = movedKeys[:assignedIndex]
		oldValues = oldValues[:assignedIndex]
		newValues = newValues[:assignedIndex]
	}

	return movedKeys, NewFTIntegerArray(oldValues...), NewFTIntegerArray(newValues...), nil
}

func (m *ListMap) AddItems(keys []ArrayElementTypeVal, ignoreError bool) ([]ArrayElementTypeVal, *FTIntegerArray, error) {

	rows := keys[0].Length()
	newValues := make([]int64, 0)
	procKeys := make([]Key, 0)
	for i := int64(0); i < rows; i++ {
		k := KeyAt(keys, i)
		if _, ok := m.m[k]; !ok {
			newV := int64(len(m.m))
			m.m[k] = newV
			newValues = append(newValues, newV)
			if ignoreError { // For merge items only
				procKeys = append(procKeys, k)
			}
		} else if !ignoreError {
			return nil, nil, fmt.Errorf("key: %v already exists in listmap", k)
		}
	}

	if ignoreError { // For merge items only
		arrVals, err := KeysToArrayTypeVals(procKeys, m.tc)
		if err != nil {
			return nil, nil, err
		}
		return arrVals, NewFTIntegerArray(newValues...), nil
	}

	return keys, NewFTIntegerArray(newValues...), nil
}

func (m *ListMap) GetKeys(sorted bool) []Key {
	keys := make([]Key, len(m.m))

	// return keys in any order
	if !sorted {
		for k, i := range m.m {
			keys[i] = k
		}
		return keys
	}

	// else return keys in order of value (ascending)
	type kv struct {
		Key   Key
		Value int64
	}

	ss := make([]kv, 0)
	for k, v := range m.m {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value < ss[j].Value
	})

	for i, kv := range ss {
		keys[i] = kv.Key
	}

	return keys
}

func (m *ListMap) TypeCodeAsSlice() []TypeCode {
	return m.tc
}

func (m *ListMap) IntersectItems(other []ArrayElementTypeVal) ([]ArrayElementTypeVal, error) {

	keys := m.GetKeys(false)
	count := other[0].Length()

	keyIntersect := make([]Key, 0)

	for _, k1 := range keys {
		for row2 := int64(0); row2 < count; row2++ {
			k2 := KeyAt(other, row2)
			if k1 == k2 {
				keyIntersect = append(keyIntersect, k1)
				break
			}
		}
	}

	arrVals, err := KeysToArrayTypeVals(keyIntersect, m.tc)
	if err != nil {
		return nil, err
	}

	return arrVals, nil
}

func (m *ListMap) Contains(keys []ArrayElementTypeVal) *FTIntegerArray {
	if len(m.m) == 0 {
		return NewFTIntegerArray()
	}

	rows := keys[0].Length()

	newKeys := make([]Key, rows)
	for i := int64(0); i < rows; i++ {
		k := KeyAt(keys, i)
		newKeys[i] = k
	}

	containsArr := make([]int64, rows)
	for i, k := range newKeys {
		if _, ok := m.m[k]; ok {
			containsArr[i] = 1
		} else {
			containsArr[i] = 0
		}
	}
	return NewFTIntegerArray(containsArr...)
}

func ToArray(items []interface{}) interface{} {
	switch len(items) {
	case 1:
		return Key{(*[1]interface{})(items)}
	case 2:
		return Key{(*[2]interface{})(items)}
	case 3:
		return Key{(*[3]interface{})(items)}
	case 4:
		return Key{(*[4]interface{})(items)}
	case 5:
		return Key{(*[5]interface{})(items)}
	default:
		arrType := reflect.ArrayOf(len(items), interfaceType)

		arr := reflect.New(arrType).Elem()
		slc := arr.Slice(0, len(items)).Interface().([]interface{})
		copy(slc, items)

		// for i, x := range xs {
		// 	switch v := x.(type) {
		// 	case []byte:
		// 		bArrType := reflect.ArrayOf(len(v), reflect.TypeOf(v[0]))
		// 		bArr := reflect.New(bArrType).Elem()

		// 		bSlc := bArr.Slice(0, len(v)).Interface().([]byte)
		// 		copy(bSlc, v)

		// 		slc[i] = bArr.Interface()
		// 	default:
		// 		slc[i] = x
		// 	}
		// }

		return Key{arr.Interface()}
	}
}
