package seedr

import (
	"reflect"
	"testing"
)

func TestMNSliceSeq(t *testing.T) {
	g := mnSliceSeq([]interface{}{1, 2, 3}, 2)

	res := []interface{}{}
	for i := 0; i < 6; i++ {
		res = append(res, g.Next())
	}

	expect := []interface{}{1, 1, 2, 2, 3, 3}
	if !reflect.DeepEqual(res, expect) {
		t.Errorf("expected %v, got %v", expect, res)
	}

}

func TestLoop(t *testing.T) {
	vals := []int{1, 2, 3}
	l := Loop(vals)
	ret := []interface{}{}
	for i := 0; i < len(vals)*2+1; i++ {
		ret = append(ret, l.Next())
	}
	if !reflect.DeepEqual(ret, []interface{}{1, 2, 3, 1, 2, 3, 1}) {
		t.Errorf("Loop failed")
	}
}

func TestPickRandom(t *testing.T) {
	vals := []int{1, 2, 3}
	l := PickRandom(vals)
	ret := map[interface{}]bool{}
	for i := 0; i < 100 || len(ret) != len(vals); i++ {
		ret[l.Next()] = true
	}
	if !reflect.DeepEqual(ret, map[interface{}]bool{1: true, 2: true, 3: true}) {
		t.Errorf("PickRandom failed")
	}
}
