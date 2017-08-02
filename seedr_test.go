package seedr

import (
	"reflect"
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
