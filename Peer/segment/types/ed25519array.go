// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package types

// #cgo CFLAGS: -I${SRCDIR}/../../lib
// #cgo LDFLAGS: -L${SRCDIR}/../../lib -lftcrypto -Wl,-rpath=${SRCDIR}/../../lib
// #include "ftcrypto.h"
import "C"
import (
	"errors"
	"fmt"
	"math"
	"sort"

	"filippo.io/edwards25519"
)

var FT_EXTENDED_POINT_BYTES C.ulong = 160

var errLengthsDoNotMatch = errors.New("array sizes do not match")
var errIndexesOutOfRange = errors.New("indexes out of range")

func InitializeGPU() error {
	err := C.ft_crypto_init()
	if err != C.FT_ERR_NO_ERROR {
		return FTError(err)
	}
	return nil
}

type GPUMemoryStats struct {
	Free  uint64
	Total uint64
}

func (m GPUMemoryStats) Used() uint64 {
	return m.Total - m.Free
}

func GetGPUMemoryStats() (GPUMemoryStats, error) {
	free := C.size_t(0)
	total := C.size_t(0)

	err := C.ft_crypto_device_memory(&free, &total)
	if err != C.FT_ERR_NO_ERROR {
		return GPUMemoryStats{}, FTError(err)
	}
	return GPUMemoryStats{
		Free:  uint64(free),
		Total: uint64(total),
	}, nil
}

type FTError C.ft_error

func (e FTError) Error() string {
	cStr := C.ft_error_str(C.int(e))
	return C.GoString(cStr)
}

type Ed25519Array struct {
	handle         C.ft_ge25519_array
	referenceCount int
}

func NewEmptyEd25519Array() *Ed25519Array {
	return &Ed25519Array{nil, 0}
}
func newEd25519ArrayFromHandle(h C.ft_ge25519_array) *Ed25519Array {
	return &Ed25519Array{h, 0}
}

func NewEd25519Array(size int64, value *edwards25519.Scalar) (*Ed25519Array, error) {
	if size == 0 {
		return NewEmptyEd25519Array(), nil
	}

	arr := (C.ft_ge25519_array)(nil)

	var v *C.uchar
	if value != nil {
		bytes := value.Bytes()
		v = (*C.uchar)(&bytes[0])
	}

	err := C.ft_array_init_scalar(&arr, C.size_t(size), v)
	if err != C.FT_ERR_NO_ERROR {
		return nil, FTError(err)
	}

	return newEd25519ArrayFromHandle(arr), nil
}
func NewEd25519ArrayFromBytes(bytes []byte) (*Ed25519Array, error) {
	if len(bytes) == 0 {
		return NewEmptyEd25519Array(), nil
	}

	arr := (C.ft_ge25519_array)(nil)

	resultCode := C.ft_array_from_bytes(&arr, (*C.uchar)(&bytes[0]), C.ulong(len(bytes)))
	if resultCode != C.FT_ERR_NO_ERROR {
		return nil, FTError(resultCode)
	}

	return newEd25519ArrayFromHandle(arr), nil
}

func NewEd25519ArrayFromFoldedBytearrayArray(value *FTBytearrayArray) (*Ed25519Array, error) {
	if value.Width() != C.FT_FOLDED_POINT_BYTES {
		return NewEmptyEd25519Array(), fmt.Errorf("bytearray array was not a b%v", C.FT_FOLDED_POINT_BYTES)
	}

	bytes := make([]byte, 0, len(value.array)*C.FT_FOLDED_POINT_BYTES)
	for _, v := range value.array {
		bytes = append(bytes, v...)
	}

	return NewEd25519ArrayFromFoldedBytes(bytes)
}

func NewEd25519ArrayFromFoldedBytes(bytes []byte) (*Ed25519Array, error) {
	if len(bytes) == 0 {
		return NewEmptyEd25519Array(), nil
	}

	arr := (C.ft_ge25519_array)(nil)

	var result C.long = -1

	resultCode := C.ft_array_from_bytes_folded(&result, &arr, (*C.uchar)(&bytes[0]), C.ulong(len(bytes)))
	if resultCode != C.FT_ERR_NO_ERROR {
		return nil, FTError(resultCode)
	}

	if result != -1 {
		panic(fmt.Sprintf("conversion last failed at index %v", result))
	}

	return newEd25519ArrayFromHandle(arr), nil
}

func NewEd25519ArrayFromAffineBytearrayArray(value *FTBytearrayArray) (*Ed25519Array, error) {
	if value.Width() != C.FT_AFFINE_POINT_BYTES {
		return NewEmptyEd25519Array(), fmt.Errorf("bytearray array was not a b%v", C.FT_AFFINE_POINT_BYTES)
	}

	bytes := make([]byte, 0, len(value.array)*C.FT_AFFINE_POINT_BYTES)
	for _, v := range value.array {
		bytes = append(bytes, v...)
	}

	return NewEd25519ArrayFromAffineBytes(bytes)
}

func NewEd25519ArrayFromAffineBytes(bytes []byte) (*Ed25519Array, error) {
	if len(bytes) == 0 {
		return NewEmptyEd25519Array(), nil
	}

	arr := (C.ft_ge25519_array)(nil)

	resultCode := C.ft_array_from_bytes_affine(&arr, (*C.uchar)(&bytes[0]), C.ulong(len(bytes)))
	if resultCode != C.FT_ERR_NO_ERROR {
		return nil, FTError(resultCode)
	}

	var result C.long = -1
	resultCode = C.ft_array_validate(&result, arr)
	if resultCode != C.FT_ERR_NO_ERROR {
		return nil, FTError(resultCode)
	}

	if result != -1 {
		panic(fmt.Sprintf("conversion last failed at index %v", result))
	}

	return newEd25519ArrayFromHandle(arr), nil
}

func NewEd25519ArrayFromInt(values []*edwards25519.Scalar) (*Ed25519Array, error) {
	if len(values) == 0 {
		return NewEmptyEd25519Array(), nil
	}

	arr := (C.ft_ge25519_array)(nil)

	bytes := ScalarsToBytes(values)

	resultCode := C.ft_array_from_scalars(&arr, (*C.uchar)(&bytes[0]), C.ulong(len(values)))
	if resultCode != C.FT_ERR_NO_ERROR {
		return nil, FTError(resultCode)
	}

	return newEd25519ArrayFromHandle(arr), nil
}

func NewEd25519ArrayFromInt64s(values ...int64) (*Ed25519Array, error) {
	if len(values) == 0 {
		return NewEmptyEd25519Array(), nil
	}

	for _, v := range values {
		if v < 0 {
			return nil, errors.New("only positive integers up to 9223372036854775807 are supported")
		}
	}

	arr := (C.ft_ge25519_array)(nil)

	ints := make([]C.uint64_t, len(values))
	for i, v := range values {
		ints[i] = C.uint64_t(v)
	}

	resultCode := C.ft_array_from_small_scalars(&arr, (*C.uint64_t)(&ints[0]), C.size_t(len(ints)))
	if resultCode != C.FT_ERR_NO_ERROR {
		return nil, FTError(resultCode)
	}

	return newEd25519ArrayFromHandle(arr), nil
}
func NewEd25519ArrayFromInt64sOrPanic(values ...int64) *Ed25519Array {
	xs, err := NewEd25519ArrayFromInt64s(values...)
	if err != nil {
		panic(err)
	}
	return xs
}

func NewEd25519ArrayFromPoint(size int64, points *Ed25519Array, pointIndex int64) (*Ed25519Array, error) {
	if size == 0 {
		return NewEmptyEd25519Array(), nil
	}

	arr := (C.ft_ge25519_array)(nil)

	resultCode := C.ft_array_init_point(&arr, C.size_t(size), points.handle, C.size_t(pointIndex))
	if resultCode != C.FT_ERR_NO_ERROR {
		return nil, FTError(resultCode)
	}
	return newEd25519ArrayFromHandle(arr), nil
}

func (xs *Ed25519Array) String() string {
	return fmt.Sprintf("<Ed25519: %v encrypted values>", xs.Length())
}

func (xs *Ed25519Array) IsEmpty() bool {
	return xs.handle == nil
}

func (xs *Ed25519Array) TypeCode() TypeCode {
	return "E"
}
func (xs *Ed25519Array) Name() string {
	return "Ed25519Array"
}
func (xs *Ed25519Array) EstimatedSize() int64 {
	return xs.Length() * int64(FT_EXTENDED_POINT_BYTES)
}
func (xs *Ed25519Array) DebugString() string {
	return fmt.Sprintf("%v(Length=%v,Memory=%v)", xs.Name(), xs.Length(), PrintSize(uint64(xs.EstimatedSize())))
}
func (xs *Ed25519Array) GetBinaryArray(index int) ([]byte, error) {
	if xs.IsEmpty() {
		return make([]byte, 0), nil
	}
	return xs.ToBytes(), nil
}

func (xs *Ed25519Array) Equals(other TypeVal) bool {
	if ys, ok := other.(*Ed25519Array); ok {
		if xs.IsEmpty() {
			return ys.IsEmpty()
		}
		if ys.IsEmpty() {
			return false
		}

		xsLen := xs.Length()
		ysLen := ys.Length()
		if xsLen != ysLen {
			return false
		}

		results := make([]int64, xs.Length())

		err := C.ft_equal((*C.int64_t)(&results[0]), xs.handle, ys.handle)
		if err != C.FT_ERR_NO_ERROR {
			panic(fmt.Sprintf("ft_equals: %v", err))
		}

		for _, v := range results {
			if v == 0 {
				return false
			}
		}
		return true
	}

	return false
}
func (xs *Ed25519Array) Copy() (*Ed25519Array, error) {
	if xs.IsEmpty() {
		return NewEmptyEd25519Array(), nil
	}
	return xs.GetRange(0, xs.Length(), 1)
}
func (v *Ed25519Array) Lookup(indexes *FTIntegerArray, defaultValue ArrayTypeVal) (ArrayTypeVal, error) {
	if defaultValue == nil {
		zero, err := NewEd25519ArrayFromInt64s(0)
		if err != nil {
			return nil, err
		}
		defaultValue = zero
		defer zero.Free()
	}
	return v.Get(indexes, defaultValue)
}
func (v *Ed25519Array) Get(indexes *FTIntegerArray, defaultValue ArrayTypeVal) (ArrayTypeVal, error) {
	if indexes == nil {
		xs, err := v.Clone()
		if err != nil {
			return nil, err
		}
		return xs.(ArrayTypeVal), nil
	}

	length := v.Length()

	if defaultValue == nil {
		for _, v := range indexes.array {
			if v >= length || v < 0 {
				return nil, fmt.Errorf("out of range: %v", v)
			}
		}

		return v.GetSubset(indexes.array)
	}

	inRangeIndex := make([]int64, 0)
	inRangeIndexValue := make([]int64, 0)

	for i, v := range indexes.array {
		if v >= 0 && v < length {
			inRangeIndex = append(inRangeIndex, int64(i))
			inRangeIndexValue = append(inRangeIndexValue, v)
		}
	}

	if len(inRangeIndex) == len(indexes.array) {
		return v.GetSubset(indexes.array)
	}

	defaultEd25519, ok := defaultValue.(*Ed25519Array)
	if !ok || defaultEd25519.Length() != 1 {
		return nil, errors.New("default value must be a singleton Ed25519 array")
	}

	result, err := NewEd25519ArrayFromPoint(indexes.Length(), defaultEd25519, 0)
	if err != nil {
		return nil, err
	}

	existing, err := v.GetSubset(inRangeIndexValue)
	if err != nil {
		return nil, err
	}
	defer existing.Free()

	err = result.AssignToSubset(inRangeIndex, existing)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (xs *Ed25519Array) Set(indexes *FTIntegerArray, values ArrayTypeVal) error {
	vs, ok := values.(*Ed25519Array)
	if !ok {
		return fmt.Errorf("values is not an %v", xs.Name())
	}
	return xs.AssignToSubset(indexes.array, vs)
}

func (xs *Ed25519Array) Broadcast(length int64) error {
	if xs.Length() != 1 {
		return errors.New("only singleton arrays can be broadcasted")
	}

	if length == 0 {
		C.ft_array_free(xs.handle)
		xs.handle = nil

		return nil
	}

	arr := (C.ft_ge25519_array)(nil)

	resultCode := C.ft_array_init_point(&arr, C.size_t(length), xs.handle, 0)
	if resultCode != C.FT_ERR_NO_ERROR {
		return FTError(resultCode)
	}

	C.ft_array_free(xs.handle)
	xs.handle = arr

	return nil
}

func (xs *Ed25519Array) GetSubset(indexes []int64) (*Ed25519Array, error) {
	arr := (C.ft_ge25519_array)(nil)

	if len(indexes) == 0 {
		return NewEmptyEd25519Array(), nil
	}

	if xs.IsEmpty() {
		return NewEmptyEd25519Array(), errIndexesOutOfRange
	}

	is := make([]C.size_t, len(indexes))
	for i, v := range indexes {
		is[i] = C.size_t(v)
	}

	err := C.ft_array_get_subset(&arr, (*C.size_t)(&is[0]), C.ulong(len(is)), xs.handle)
	if err != C.FT_ERR_NO_ERROR {
		return nil, FTError(err)
	}

	return newEd25519ArrayFromHandle(arr), nil
}
func (xs *Ed25519Array) AssignZeroToSubset(indexes []int64) error {
	if len(indexes) == 0 {
		return nil
	}

	if xs.IsEmpty() {
		if len(indexes) > 0 {
			return errIndexesOutOfRange
		}
		return nil
	}

	uniqueIndexes := make([]C.size_t, 0, len(indexes))
	uniqueIndexeValuesMap := make(map[C.size_t]struct{})
	for _, v := range indexes {
		idx := C.size_t(v)

		if _, exists := uniqueIndexeValuesMap[idx]; !exists {
			uniqueIndexeValuesMap[idx] = struct{}{}
			uniqueIndexes = append(uniqueIndexes, idx)
		}
	}

	sort.Slice(uniqueIndexes, func(i, j int) bool {
		return i < j
	})

	resultCode := C.ft_array_assign_zero_to_subset(xs.handle, (*C.size_t)(&uniqueIndexes[0]), C.ulong(len(uniqueIndexes)))
	if resultCode != C.FT_ERR_NO_ERROR {
		return FTError(resultCode)
	}
	return nil
}
func (xs *Ed25519Array) AssignToSubset(indexes []int64, values *Ed25519Array) error {
	if len(indexes) == 0 {
		return nil
	}

	if xs.IsEmpty() {
		if len(indexes) > 0 {
			return errIndexesOutOfRange
		}
		return nil
	}
	// Due to a bug in libftcrypto, we want to reduce the size of the index array
	// to less than or equal the size of the xs Ed25519 array. We do this by
	// removing duplicate indexes (taking the last), and then taking the
	// subset of the values array using the original position of the index in
	// the indexes array. For example:

	// indexes: [0,24,12,3,12]
	// values: [1,2,3,4,5]

	// uniqueIndexes: [0,24,3,12]
	// uniqueValueIndexes: [0,1,3,4]
	// uniqueValues: [1,2,4,5]

	uniqueIndexeValuesMap := make(map[C.size_t]int64)
	for i, v := range indexes {
		uniqueIndexeValuesMap[C.size_t(v)] = int64(i)
	}

	uniqueIndexes := make([]C.size_t, 0, len(uniqueIndexeValuesMap))
	uniqueValueIndexes := make([]int64, 0, len(uniqueIndexeValuesMap))

	for k, v := range uniqueIndexeValuesMap {
		uniqueIndexes = append(uniqueIndexes, k)
		uniqueValueIndexes = append(uniqueValueIndexes, v)
	}

	uniqueValues, err := values.GetSubset(uniqueValueIndexes)
	if err != nil {
		return err
	}
	defer uniqueValues.Free()

	resultCode := C.ft_array_assign_to_subset(xs.handle, (*C.size_t)(&uniqueIndexes[0]), C.ulong(len(uniqueIndexes)), uniqueValues.handle)
	if resultCode != C.FT_ERR_NO_ERROR {
		return FTError(resultCode)
	}
	return nil
}
func (v *Ed25519Array) Clone() (TypeVal, error) {
	if v.IsEmpty() {
		return NewEmptyEd25519Array(), nil
	}
	return v.GetRange(0, v.Length(), 1)
}
func (v *Ed25519Array) AsType(tc TypeCode) (TypeVal, error) {
	return nil, fmt.Errorf("conversion not supported: %v -> %v", v.TypeCode(), tc)
}
func (v *Ed25519Array) Remove(indexes *FTIntegerArray) error {
	if v.IsEmpty() {
		if indexes.Length() > 0 {
			return errIndexesOutOfRange
		}
		return nil
	}

	is := make([]C.size_t, indexes.Length())
	for i, v := range indexes.Values() {
		is[i] = C.size_t(v)
	}

	err := C.ft_array_delete_subset(v.handle, (*C.size_t)(&is[0]), C.ulong(len(is)))
	if err != C.FT_ERR_NO_ERROR {
		return FTError(err)
	}
	return nil
}

func (xs *Ed25519Array) GetRange(start int64, stop int64, step int64) (*Ed25519Array, error) {
	if xs.IsEmpty() {
		return NewEmptyEd25519Array(), errors.New("can not get range on empty array")
	}

	arr := (C.ft_ge25519_array)(nil)

	err := C.ft_array_get_range(&arr, C.size_t(start), C.size_t(stop), C.size_t(step), xs.handle)

	if err != C.FT_ERR_NO_ERROR {
		return nil, FTError(err)
	}

	return newEd25519ArrayFromHandle(arr), nil
}

func (xs *Ed25519Array) ReferenceCount() int {
	return xs.referenceCount
}
func (xs *Ed25519Array) IncrementReferenceCount() {
	xs.referenceCount++
}

func (xs *Ed25519Array) Free() {
	if xs.referenceCount > 1 {
		xs.referenceCount--
		return
	}

	if xs.handle != nil {
		C.ft_array_free(xs.handle)
		xs.handle = nil
		xs.referenceCount = 0
	}
}

func (xs *Ed25519Array) Length() int64 {
	if xs.IsEmpty() {
		return 0
	}

	result := C.ft_array_get_length(xs.handle)

	return int64(result)
}

func (xs *Ed25519Array) SetLength(x int64) error {
	if xs.IsEmpty() {
		if x == 0 {
			return nil
		}

		arr := (C.ft_ge25519_array)(nil)

		err := C.ft_array_init_scalar(&arr, C.size_t(x), nil)
		if err != C.FT_ERR_NO_ERROR {
			return FTError(err)
		}

		xs.handle = arr
		return nil

	} else if x == 0 {
		C.ft_array_free(xs.handle)
		xs.handle = nil

		return nil
	}

	resultCode := C.ft_array_set_length(xs.handle, C.size_t(x))
	if resultCode != C.FT_ERR_NO_ERROR {
		return FTError(resultCode)
	}
	return nil
}

func (xs *Ed25519Array) ToBytes() []byte {
	if xs.IsEmpty() {
		return make([]byte, 0)
	}

	len := C.ft_array_get_length(xs.handle)
	nBytes := len * FT_EXTENDED_POINT_BYTES
	b := make([]byte, nBytes)

	C.ft_array_to_bytes((*C.uchar)(&b[0]), nBytes, xs.handle)

	return b
}

func (xs *Ed25519Array) ToFoldedBytes() ([]byte, int64) {
	if xs.IsEmpty() {
		return make([]byte, 0), C.FT_FOLDED_POINT_BYTES
	}

	len := C.ft_array_get_length(xs.handle)
	nBytes := len * C.FT_FOLDED_POINT_BYTES
	b := make([]byte, nBytes)

	C.ft_array_to_bytes_folded((*C.uchar)(&b[0]), nBytes, xs.handle)

	return b, C.FT_FOLDED_POINT_BYTES
}

func (xs *Ed25519Array) ToAffineBytes() ([]byte, int64) {
	if xs.IsEmpty() {
		return make([]byte, 0), C.FT_AFFINE_POINT_BYTES
	}

	len := C.ft_array_get_length(xs.handle)
	nBytes := len * C.FT_AFFINE_POINT_BYTES
	b := make([]byte, nBytes)

	C.ft_array_to_bytes_affine((*C.uchar)(&b[0]), nBytes, xs.handle)

	return b, C.FT_AFFINE_POINT_BYTES
}

func (xs *Ed25519Array) Index() *FTIntegerArray {
	if xs.IsEmpty() {
		return NewFTIntegerArray()
	}

	values := make([]int64, xs.Length())
	C.ft_index((*C.int64_t)(&values[0]), xs.handle)

	result := make([]int64, 0)
	for i, v := range values {
		if v == 1 {
			result = append(result, int64(i))
		}
	}

	return &FTIntegerArray{result}
}

func (xs *Ed25519Array) Scale(y *edwards25519.Scalar) error {
	if xs.IsEmpty() {
		return nil
	}

	bytes := y.Bytes()

	C.ft_scale(xs.handle, (*C.uchar)(&bytes[0]))

	return nil
}

func (xs *Ed25519Array) Contains(values ArrayTypeVal) (*FTIntegerArray, error) {
	ys, ok := values.(*Ed25519Array)
	if !ok {
		return nil, fmt.Errorf("values is not an %v", xs.Name())
	}

	if xs.IsEmpty() {
		return NewFTIntegerArray(make([]int64, int(ys.Length()))...), nil
	}

	if ys.IsEmpty() {
		return NewFTIntegerArray(), nil
	}

	results := make([]int64, ys.Length())

	resultCode := C.ft_array_contains((*C.int64_t)(&results[0]), ys.handle, xs.handle)
	if resultCode != C.FT_ERR_NO_ERROR {
		return nil, FTError(resultCode)
	}

	return &FTIntegerArray{results}, nil
}

func (xs *Ed25519Array) CumSum() (ArrayTypeVal, error) {
	if xs.IsEmpty() {
		return NewEmptyEd25519Array(), nil
	}

	ys, err := xs.Copy()
	if err != nil {
		return nil, err
	}

	resultCode := C.ft_array_scan(ys.handle)
	if resultCode != C.FT_ERR_NO_ERROR {
		return nil, FTError(resultCode)
	}

	return ys, nil
}

func (xs Ed25519Array) Mux(condition *FTIntegerArray, ifFalse ArrayTypeVal) (ArrayTypeVal, error) {
	fs, ok := ifFalse.(*Ed25519Array)
	if !ok {
		return nil, fmt.Errorf("ifFalse is not an %v", xs.Name())
	}
	if xs.Length() != fs.Length() || xs.Length() != condition.Length() {
		return nil, errLengthsDoNotMatch
	}
	if xs.IsEmpty() {
		return NewEmptyEd25519Array(), nil
	}

	results, err := xs.Copy()
	if err != nil {
		return nil, err
	}

	resultCode := C.ft_array_mux(results.handle, (*C.int64_t)(&condition.Values()[0]), fs.handle)
	if resultCode != C.FT_ERR_NO_ERROR {
		return nil, FTError(resultCode)
	}
	return results, nil

}

func (xs *Ed25519Array) ReduceSum(indexes *FTIntegerArray, values ArrayTypeVal) error {
	if xs.IsEmpty() {
		if len(indexes.array) > 0 {
			return errIndexesOutOfRange
		}
		return nil
	}

	_, ok := values.(*Ed25519Array)
	if !ok {
		return fmt.Errorf("values is not an %v", xs.Name())
	}

	err := xs.AssignZeroToSubset(indexes.array)
	if err != nil {
		return err
	}

	return xs.ReduceISum(indexes, values)
}

const maxChunkSize int = 4_000_000

func (xs *Ed25519Array) ReduceISum(indexes *FTIntegerArray, values ArrayTypeVal) error {
	if xs.IsEmpty() {
		if len(indexes.array) > 0 {
			return errIndexesOutOfRange
		}
		return nil
	}

	vs, ok := values.(*Ed25519Array)
	if !ok {
		return fmt.Errorf("values is not an %v", xs.Name())
	}

	reduceISumChunk := func(start int64, end int64) error {
		indexesChunk := indexes.array[start:end]
		summandChunk, err := vs.GetRange(int64(start), int64(end), 1)
		if err != nil {
			return err
		}
		defer summandChunk.Free()

		is := make([]C.size_t, len(indexesChunk))
		for i, v := range indexesChunk {
			is[i] = C.size_t(v)
		}

		resultCode := C.ft_reduce_isum(xs.handle, (*C.size_t)(&is[0]), C.ulong(len(is)), summandChunk.handle)
		if resultCode != C.FT_ERR_NO_ERROR {
			return FTError(resultCode)
		}
		return nil
	}

	for i := 0; i < len(indexes.array); i += maxChunkSize {
		end := int64(math.Min(float64(len(indexes.array)), float64(i+maxChunkSize)))

		err := reduceISumChunk(int64(i), end)
		if err != nil {
			return err
		}
	}

	return nil
}

func asEd25519Array(xs ArrayTypeVal) (*Ed25519Array, error) {
	ys, ok := xs.(*Ed25519Array)
	if !ok {
		return NewEmptyEd25519Array(), fmt.Errorf("value is not an %v", xs.Name())
	}
	return ys, nil
}

func (v *Ed25519Array) Eq(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asEd25519Array(other)
	if err != nil {
		return nil, err
	}

	results := make([]int64, v.Length())

	if !v.IsEmpty() {
		resultCode := C.ft_equal((*C.int64_t)(&results[0]), v.handle, bs.handle)

		if resultCode != C.FT_ERR_NO_ERROR {
			return nil, FTError(resultCode)
		}
	}

	return &FTIntegerArray{results}, nil
}

func (v *Ed25519Array) Ne(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asEd25519Array(other)
	if err != nil {
		return nil, err
	}

	if v.Length() != bs.Length() {
		return nil, errLengthsDoNotMatch
	}

	results := make([]int64, v.Length())

	if !v.IsEmpty() {
		resultCode := C.ft_not_equal((*C.int64_t)(&results[0]), v.handle, bs.handle)
		if resultCode != C.FT_ERR_NO_ERROR {
			return nil, FTError(resultCode)
		}
	}

	return &FTIntegerArray{results}, nil
}

func (xs *Ed25519Array) Neg() (ArrayNegTypeVal, error) {
	if xs.IsEmpty() {
		return NewEmptyEd25519Array(), nil
	}

	result, err := xs.Clone()
	if err != nil {
		return NewEmptyEd25519Array(), err
	}

	ed25519Result := result.(*Ed25519Array)
	C.ft_neg(ed25519Result.handle)

	return ed25519Result, nil
}

func (xs *Ed25519Array) Add(other ArrayTypeVal) (ArrayTypeVal, error) {
	ys, err := asEd25519Array(other)
	if err != nil {
		return nil, err
	}

	if xs.Length() != ys.Length() {
		return nil, errLengthsDoNotMatch
	}

	if xs.IsEmpty() {
		return NewEmptyEd25519Array(), nil
	}

	result, err := xs.Copy()
	if err != nil {
		return nil, err
	}

	resultCode := C.ft_add(result.handle, ys.handle)
	if resultCode != C.FT_ERR_NO_ERROR {
		return nil, FTError(resultCode)
	}

	return result, nil
}

func (xs *Ed25519Array) Sub(other ArrayTypeVal) (ArrayTypeVal, error) {
	ys, err := asEd25519Array(other)
	if err != nil {
		return nil, err
	}

	if xs.Length() != ys.Length() {
		return nil, errLengthsDoNotMatch
	}

	if xs.IsEmpty() {
		return NewEmptyEd25519Array(), nil
	}

	result, err := xs.Copy()
	if err != nil {
		return nil, err
	}

	resultCode := C.ft_sub(result.handle, ys.handle)
	if resultCode != C.FT_ERR_NO_ERROR {
		return nil, FTError(resultCode)
	}
	return result, nil
}

func (xs *Ed25519Array) Mul(other ArrayTypeVal) (ArrayTypeVal, error) {
	ys, err := asFTEd25519IntArray(other)
	if err != nil {
		return nil, err
	}

	if xs.Length() != ys.Length() {
		return nil, errLengthsDoNotMatch
	}

	if xs.IsEmpty() {
		return NewEmptyEd25519Array(), nil
	}

	result, err := xs.Copy()
	if err != nil {
		return nil, err
	}

	bytes := ScalarsToBytes(ys.Values())

	resultCode := C.ft_mul(result.handle, (*C.uchar)(&bytes[0]), C.ulong(ys.Length()))
	if resultCode != C.FT_ERR_NO_ERROR {
		return nil, FTError(resultCode)
	}
	return result, nil
}
