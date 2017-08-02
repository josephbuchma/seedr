package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

var TestTime = TimeNowSecondPrecision()

func TimeNowSecondPrecision() time.Time {
	t := time.Now()
	t = t.Add(-time.Duration(t.Nanosecond()))
	return t.UTC()
}

func AssertEqualJSON(t *testing.T, exp, got interface{}) {
	ej, err := json.Marshal(exp)
	if err != nil {
		panic(err)
	}
	gj, err := json.Marshal(got)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(ej, gj) {
		t.Errorf("Expeted:\n%s\ngot:\n%s\n", ej, gj)
	}

}
func AssertDeepEqual(t *testing.T, exp, got interface{}) {
	if !reflect.DeepEqual(exp, got) {
		//t.Errorf("Expeted:\n%#v\ngot:\n%#v\n", exp, got)
		panic(fmt.Sprintf("Expeted:\n%#v\ngot:\n%#v\n", exp, got))
	}
}
