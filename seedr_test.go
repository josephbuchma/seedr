package seedr

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestTrait_buildPublics(t *testing.T) {
	testCases := []struct {
		FactoryTraits        Traits
		ExpectedBuilt        Traits
		ExpectedPublicTraits []string
	}{
		{
			Traits{
				"User": {
					"name": "Jon",
					"age":  20,
				},
			},
			Traits{
				"User": {
					"name": "Jon",
					"age":  20,
				},
			},
			[]string{"User"},
		},
		{
			Traits{
				"basic": {
					"name": "Jon",
					"age":  20,
				},
				"User": {
					Include: "basic",
					"age":   21,
				},
			},
			Traits{
				"basic": {
					"name": "Jon",
					"age":  20,
				},
				"User": {
					"name": "Jon",
					"age":  21,
				},
			},
			[]string{"User"},
		},
		{
			Traits{
				"basic": {
					"name": "Jon",
					"age":  20,
				},
				"male": {
					"sex": "male",
				},
				"User": {
					Include: "basic male",
				},
			},
			Traits{
				"basic": {
					"name": "Jon",
					"age":  20,
				},
				"male": {
					"sex": "male",
				},
				"User": {
					"name": "Jon",
					"age":  20,
					"sex":  "male",
				},
			},
			[]string{"User"},
		},
		{
			Traits{
				"basic": {
					"name": "Jon",
					"age":  20,
				},
				"male": {
					"sex": "male",
				},
				"User": {
					Include: "basic male",
					"name":  "Walter",
				},
			},
			Traits{
				"basic": {
					"name": "Jon",
					"age":  20,
				},
				"male": {
					"sex": "male",
				},
				"User": {
					"name": "Walter",
					"age":  20,
					"sex":  "male",
				},
			},
			[]string{"User"},
		},
		{
			Traits{
				"basic": {
					"name": "Jon",
					"age":  20,
					"sex":  "male",
				},
				"old": {
					Include: "basic",
					"age":   80,
				},
				"User": {
					Include: "basic old",
					"name":  "Oliver",
				},
			},
			Traits{
				"basic": {
					"name": "Jon",
					"age":  20,
					"sex":  "male",
				},
				"old": {
					Include: "basic",
					"age":   80,
				},
				"User": {
					"name": "Oliver",
					"age":  80,
					"sex":  "male",
				},
			},
			[]string{"User"},
		},
		{
			Traits{
				"basic": {
					"name": "Jon",
					"age":  20,
					"sex":  "male",
				},
				"old": {
					Include: "basic",
					"age":   80,
				},
				"withPhone": {
					"phone": "+3802939524",
				},
				"female": {
					"sex": "female",
				},
				"oldWoman": {
					Include: "basic old female",
					"name":  "Ann",
				},
				"User": {
					Include:  "basic oldWoman withPhone",
					"active": true,
				},
			},
			Traits{
				"basic": {
					"name": "Jon",
					"age":  20,
					"sex":  "male",
				},
				"old": {
					Include: "basic",
					"age":   80,
				},
				"withPhone": {
					"phone": "+3802939524",
				},
				"female": {
					"sex": "female",
				},
				"oldWoman": {
					Include: "basic old female",
					"name":  "Ann",
				},
				"User": {
					"name":   "Ann",
					"age":    80,
					"sex":    "female",
					"phone":  "+3802939524",
					"active": true,
				},
			},
			[]string{"User"},
		},
	}

	for _, tc := range testCases {
		publics := tc.FactoryTraits.buildPublics()
		if !reflect.DeepEqual(publics, tc.ExpectedPublicTraits) {
			t.Errorf("Expected Tst public trait built")
		}

		if !reflect.DeepEqual(tc.FactoryTraits, tc.ExpectedBuilt) {
			t.Errorf("Expected built:\n%#v,\ngot\n%#v", tc.ExpectedBuilt, tc.FactoryTraits)
		}
	}
}

func Test_resolveDependentFields(t *testing.T) {
	t.Run("Basic dependent field", func(t *testing.T) {
		pub := &publicTrait{
			trait: Trait{
				"first_name": SequenceString("Jon-%d"),
				"last_name":  SequenceString("Snow-%d"),
				"full_name": DependsOn("first_name", "last_name").Generate(func(t Trait) interface{} {
					return fmt.Sprintf("%s %s", t["first_name"], t["last_name"])
				}),
			},
		}

		rt := pub.next(1, nil)
		if rt.data[0]["first_name"].(string) != "Jon-1" {
			t.Fatalf("Invalid first_name value")
		}
		if d, ok := rt.dependent["full_name"]; !ok || !stringSice(d.fields).contains("first_name") {
			t.Fatalf("Invalid dependent: %#v", rt.dependent)
		}
		resolveDependentFields(rt.data[0], rt.dependent)
		if fn, ok := rt.data[0]["full_name"].(string); !ok || fn != "Jon-1 Snow-1" {
			t.Fatalf("invalid full_name, expected %s, got %s", "Jon-1 Snow-1", fn)
		}
	})

	t.Run("Must panic on circular dependency", func(t *testing.T) {
		defer func() {
			r := recover().(string)
			if !strings.HasPrefix(r, "Circular field dependency:") {
				t.Fatalf("no panic")
			}
		}()
		pub := &publicTrait{
			trait: Trait{
				"a": DependsOn("b").Generate(func(t Trait) interface{} { return nil }),
				"b": DependsOn("c").Generate(func(t Trait) interface{} { return nil }),
				"c": DependsOn("d").Generate(func(t Trait) interface{} { return nil }),
				"d": DependsOn("b").Generate(func(t Trait) interface{} { return nil }),
			},
		}

		rt := pub.next(1, nil)
		resolveDependentFields(rt.data[0], rt.dependent)
	})
}
