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
	"strings"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const (
	CommandNewListmap           = "command_newlistmap"
	CommandListmapKeys          = "command_listmap_keys"
	CommandListmapGetItem       = "command_listmap_getitem"
	CommandListmapContains      = "command_listmap_contains"
	CommandListmapAddItem       = "command_listmap_additem"
	CommandListmapRemoveItem    = "command_listmap_removeitem"
	CommandListmapIntersectItem = "command_listmap_intersectitem"
	CommandListmapKeysUnique    = "command_listmap_keys_unique"
	CommandListmapSetItems      = "command_listmap_setitems"
	CommandListmapCopy          = "command_listmap_copy"
)

func NewListmap(s SegmentHost, args []string) (string, error) {
	if len(args) < 4 {
		return "", fmt.Errorf("incorrect paramaters provided to newlistmap")
	}

	hResult := variables.Handle(args[0])
	tcs := types.ParseTypeCodes(args[1])

	if len(tcs) == 0 {
		return "", fmt.Errorf("no valid typecodes found in typecode string input")
	}
	hValueStrs := args[3:]
	hValues := make([]variables.Handle, len(hValueStrs))
	for i, h := range hValueStrs {
		hValues[i] = variables.Handle(h)
	}

	values := make([]types.ArrayElementTypeVal, len(hValues))
	order := args[2]

	for i := 0; i < len(values); i++ {
		tmp, err := variables.GetAs[types.ArrayElementTypeVal](s.Variables(), hValues[i])
		if err != nil {
			return "", err
		}

		values[i] = tmp
	}

	res, err := types.NewListMapFromArrays(tcs, values, order)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, res)

	return fmt.Sprintf("listmap %s %s", args[1], hResult), nil
}

func ListmapGetKeys(s SegmentHost, args []string) (string, error) {

	hTarget := variables.Handle(args[len(args)-1])
	hResultStrs := args[0 : len(args)-1]
	hResults := make([]variables.Handle, len(hResultStrs))
	for i, h := range hResultStrs {
		hResults[i] = variables.Handle(h)
	}

	lm, err := variables.GetAs[*types.ListMap](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}
	keys := lm.GetKeys(false)
	keyVals, err := types.KeysToArrayTypeVals(keys, lm.TypeCodeAsSlice())
	if err != nil {
		return "", err
	}

	res := make([]string, len(keyVals))
	for i, h := range hResults {
		s.Variables().Set(h, keyVals[i])
		res[i] = fmt.Sprintf("array %s %s", keyVals[i].TypeCode(), h)
	}

	return strings.Join(res, " "), nil
}

func ListmapGetItem(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hTarget := variables.Handle(args[1])

	var defaultVal interface{} = nil
	var err error
	if args[2] != "0" {
		hDefault := variables.Handle(args[2])
		defaultVal, err = variables.GetAsInteger(s.Variables(), hDefault)
		if err != nil {
			return "", err
		}
	}

	// TODO: check width of keys match listmap keys
	hKeyStrs := args[3:]
	hKeys := make([]variables.Handle, len(hKeyStrs))
	for i, h := range hKeyStrs {
		hKeys[i] = variables.Handle(h)
	}
	keys := make([]types.ArrayElementTypeVal, len(hKeys))

	for i := 0; i < len(keys); i++ {
		tmp, err := variables.GetAs[types.ArrayElementTypeVal](s.Variables(), hKeys[i])
		if err != nil {
			return "", err
		}

		keys[i] = tmp
	}

	lm, err := variables.GetAs[*types.ListMap](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	resVals, err := lm.GetItems(keys, defaultVal)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, resVals)
	return fmt.Sprintf("array i %s", hResult), nil
}

func ListmapContains(s SegmentHost, args []string) (string, error) {
	if len(args) < 2 {
		return "", nil
	}

	hResult := variables.Handle(args[0])
	hTarget := variables.Handle(args[1])

	// TODO: check width of keys match listmap keys
	hKeyStrs := args[2:]
	hKeys := make([]variables.Handle, len(hKeyStrs))
	for i, h := range hKeyStrs {
		hKeys[i] = variables.Handle(h)
	}
	keys := make([]types.ArrayElementTypeVal, len(hKeys))

	for i := 0; i < len(keys); i++ {
		tmp, err := variables.GetAs[types.ArrayElementTypeVal](s.Variables(), hKeys[i])
		if err != nil {
			return "", err
		}

		keys[i] = tmp
	}

	lm, err := variables.GetAs[*types.ListMap](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, lm.Contains(keys))

	return fmt.Sprintf("array i %s", hResult), nil
}

func ListmapAddItem(s SegmentHost, args []string) (string, error) {
	hResultStrs := args[0 : len(args)-4]
	hResults := make([]variables.Handle, len(hResultStrs))
	for i, h := range hResultStrs {
		hResults[i] = variables.Handle(h)
	}

	hIgnoreErr := false
	if args[len(args)-2] == "1" {
		hIgnoreErr = true
	}

	hResultVals := variables.Handle(args[len(args)-4])
	hTarget := variables.Handle(args[len(args)-3])
	hKeyStrs := strings.Split(args[len(args)-1], "_")
	hKeys := make([]variables.Handle, len(hKeyStrs))
	for i, h := range hKeyStrs {
		hKeys[i] = variables.Handle(h)
	}

	keys := make([]types.ArrayElementTypeVal, len(hKeys))

	for i := 0; i < len(keys); i++ {
		tmp, err := variables.GetAs[types.ArrayElementTypeVal](s.Variables(), hKeys[i])
		if err != nil {
			return "", err
		}

		keys[i] = tmp
	}

	lm, err := variables.GetAs[*types.ListMap](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}
	newKeys, newValues, err := lm.AddItems(keys, hIgnoreErr)
	if err != nil {
		return "", err
	}

	res := make([]string, len(newKeys))
	for i, h := range hResults {
		if len(newKeys) == len(hResults) {
			s.Variables().Set(h, newKeys[i])
		} else {
			// TODO: properly handle when len(newKeys) != len(hResults)
			return "", fmt.Errorf("no changes to be made")
		}
		res[i] = fmt.Sprintf("array %s %s", newKeys[i].TypeCode(), h)
	}

	s.Variables().Set(hResultVals, newValues)

	res = append(res, fmt.Sprintf("array i %s", hResultVals))

	return strings.Join(res, " "), nil
}

func ListmapRemoveItem(s SegmentHost, args []string) (string, error) {
	hResultStrs := args[0 : len(args)-5]
	hResults := make([]variables.Handle, len(hResultStrs))
	for i, h := range hResultStrs {
		hResults[i] = variables.Handle(h)
	}

	ignoreErr := false
	if args[len(args)-2] == "1" {
		ignoreErr = true
	}

	hResultOldVals := variables.Handle(args[len(args)-5])
	hResultNewVals := variables.Handle(args[len(args)-4])
	hTarget := variables.Handle(args[len(args)-3])
	hKeyStrs := strings.Split(args[len(args)-1], "_")
	hKeys := make([]variables.Handle, len(hKeyStrs))
	for i, h := range hKeyStrs {
		hKeys[i] = variables.Handle(h)
	}

	keys := make([]types.ArrayElementTypeVal, len(hKeys))

	for i := 0; i < len(keys); i++ {
		tmp, err := variables.GetAs[types.ArrayElementTypeVal](s.Variables(), hKeys[i])
		if err != nil {
			return "", err
		}

		keys[i] = tmp
	}

	lm, err := variables.GetAs[*types.ListMap](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}
	movedKeysTmp, oldValues, newValues, err := lm.RemoveItems(keys, ignoreErr)
	if err != nil {
		return "", err
	}

	// TODO: change return type of remove to ArrayTypeVal to clean this up
	movedKeys, err := types.KeysToArrayTypeVals(movedKeysTmp, lm.TypeCodeAsSlice())
	if err != nil {
		return "", err
	}

	res := make([]string, len(movedKeys))
	for i, h := range hResults {
		s.Variables().Set(h, movedKeys[i])
		res[i] = fmt.Sprintf("array %s %s", movedKeys[i].TypeCode().GetBase(), h)
	}
	s.Variables().Set(hResultOldVals, oldValues)
	res = append(res, fmt.Sprintf("array i %s", hResultOldVals))

	s.Variables().Set(hResultNewVals, newValues)
	res = append(res, fmt.Sprintf("array i %s", hResultNewVals))

	return strings.Join(res, " "), nil
}

func ListmapIntersectItem(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hKeyStrs := strings.Split(args[len(args)-1], "_")
	hKeys := make([]variables.Handle, len(hKeyStrs))
	for i, h := range hKeyStrs {
		hKeys[i] = variables.Handle(h)
	}

	keys := make([]types.ArrayElementTypeVal, len(hKeys))

	for i := 0; i < len(keys); i++ {
		tmp, err := variables.GetAs[types.ArrayElementTypeVal](s.Variables(), hKeys[i])
		if err != nil {
			return "", err
		}

		keys[i] = tmp
	}

	lmTarget, err := variables.GetAs[*types.ListMap](s.Variables(), variables.Handle(args[len(args)-2]))
	if err != nil {
		return "", err
	}

	intersectKeys, err := lmTarget.IntersectItems(keys)
	if err != nil {
		return "", err
	}

	res, err := types.NewListMapFromArrays(lmTarget.TypeCodeAsSlice(), intersectKeys, "pos")

	s.Variables().Set(hResult, res)

	return fmt.Sprintf("listmap %s %s", res.TypeCode(), hResult), nil
}

func ListmapKeysUnique(s SegmentHost, args []string) (string, error) {
	keyArrays := make([]types.ArrayElementTypeVal, len(args))

	for i, h := range args {
		xs, err := variables.GetAs[types.ArrayElementTypeVal](s.Variables(), variables.Handle(h))
		if err != nil {
			return "", err
		}
		keyArrays[i] = xs
	}

	m := map[types.Key]struct{}{}

	for i := int64(0); i < keyArrays[0].Length(); i++ {
		xs := make([]interface{}, len(keyArrays))
		for j, keyArray := range keyArrays {
			xs[j] = keyArray.Element(i)
		}

		k := types.NewKey(xs)

		if _, exists := m[k]; exists {
			return "bool 0", nil
		}
		m[k] = struct{}{}
	}

	return "bool 1", nil
}

func ListmapSetItem(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hTarget := variables.Handle(args[1])

	valTarget, err := s.Variables().Get(hTarget)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, valTarget)

	return Ack, nil
}

func ListmapCopy(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hTarget := variables.Handle(args[1])

	lmTarget, err := variables.GetAs[*types.ListMap](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	keysTarget := lmTarget.GetKeys(true)
	tcs := lmTarget.TypeCode().GetTypeCodeAsSlice()

	arrVals, err := types.KeysToArrayTypeVals(keysTarget, tcs)
	if err != nil {
		return "", err
	}

	res, err := types.NewListMapFromArrays(tcs, arrVals, "pos")
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, res)

	return fmt.Sprintf("listmap %s %s", lmTarget.TypeCode(), hResult), nil
}
