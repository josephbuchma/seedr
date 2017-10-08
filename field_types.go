package seedr

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

// Generator provides a way to generate dynamic field values
// for every instance created by seedr.
type Generator interface {
	// Next must return any value of supported type:
	//    - any numeric type
	//    - time.Time
	//    - sql.Scanner
	//    - nil value (interface{}(nil))
	Next() interface{}
}

// Func allows to convert func()interface{} to Generator
type Func func() interface{}

// Next implements Generator.Next
func (g Func) Next() interface{} {
	return g()
}

// Dependent holds field dependencies
type Dependent struct {
	fields []string
	do     func(t Trait) interface{}
}

// DependsOn allows to initialize field
// based on other fields of this trait.
// Seedr will ensure that listed fields are initialized
// before this one. It will panic on circular dependency.
// Must be followed by #Generate.
// Example:
//
//   "full_name": DependsOn("first_name", "last_name").Generate(func(this Trait)interface{}{
//     return fmt.Sprintf("%s %s", this["first_name"], this["last_name"])
//   },
func DependsOn(fields ...string) Dependent {
	return Dependent{fields: fields}
}

// Generate allows to compute value for field based on other fields of this trait.
// See DependsOn
func (d Dependent) Generate(f func(t Trait) interface{}) Generator {
	d.do = f
	return dependentField{d}
}

type dependentField struct {
	Dependent
}

func (d dependentField) Next() interface{} {
	return d.Dependent
}

// SequenceFunc creates generator from given func. On n'th .Next call it calls given func
// with startFrom + n argument. By default startFrom == 1
func SequenceFunc(f func(int) interface{}, startFrom ...int) Generator {
	cur := append(startFrom, 1)[0] - 1
	return Func(func() interface{} {
		cur++
		return f(cur)
	})
}

// ChainInt creates Generator that yields result of `f`
// with result of previous call to `f` as parameter (prev).
// For initial call `init` value is used.
func ChainInt(init int, f func(prev int) int) Generator {
	return Func(func() interface{} {
		init = f(init)
		return init
	})
}

// SequenceInt generates sequence of ints starting from `startFrom` (1 by default)
func SequenceInt(startFrom ...int) Generator {
	cur := append(startFrom, 1)[0] - 1
	return Func(func() interface{} {
		cur++
		return cur
	})
}

// SequenceString generates strings by given template `fmtStr`.
// Example: `SequenceString("MyString-%d")` => "MyString-1", "MyString-2",...
func SequenceString(fmtStr string, startFrom ...int) Generator {
	cur := append(startFrom, 1)[0] - 1
	return Func(func() interface{} {
		cur++
		return fmt.Sprintf(fmtStr, cur)
	})
}

// PickRandom picks value by random index from given slice/array/string.
func PickRandom(slice interface{}) Generator {
	v := reflect.ValueOf(slice)
	l := v.Len()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return Func(func() interface{} {
		return v.Index(r.Intn(l)).Interface()
	})
}

// Loop returns values from given slice sequentially.
// When last element is reached, it starts from beginning.
func Loop(slice interface{}) Generator {
	v := reflect.ValueOf(slice)
	l := v.Len()
	i := -1
	return Func(func() interface{} {
		i++
		if i == l {
			i = 0
		}
		return v.Index(i).Interface()
	})
}

// mnSliceSeq yields every element of given slice n times
func mnSliceSeq(s []interface{}, n int) Generator {
	i, j := 0, -1
	return Func(func() (ret interface{}) {
		j++
		if j < n {
			return s[i]
		}
		j = 0
		i++
		return s[i]
	})
}

type auto struct{}

// Auto is a special kind of Generator.
// "Auto" field must be initialized by Driver on Create.
// May be useful for fields like "id", "created_at", etc.
func Auto() Generator {
	return Func(func() interface{} {
		return auto{}
	})
}

// DummyText returns meaningless text up to `limit` bytes
// Limit 0 means 'no limit' (returns full length of hardcoded "Lorem Ipsum ....")
func DummyText(limit int) Generator {
	return Func(func() interface{} {
		if limit == 0 {
			return string(loremipsum)
		}
		if limit > len(loremipsum) {
			panicf("The maximum length of DummyText is %d, if you need longer make your own.", len(loremipsum))
		}
		return string(loremipsum[:limit])
	})
}

// RelationField instructs to create
// related record(s) for given Trait's field
type relationField struct {
	kind      int
	traitName string
	lfield    string
	rfield    string
	n         int
	override  Trait
}

// CreateRelated is a special Generator that will create related trait
func CreateRelated(traitName string) Generator {
	return Func(func() interface{} {
		return &relationField{traitName: traitName, n: 1}
	})
}

// CreateRelatedBatch is a special Generator that will create a batch of related traits
func CreateRelatedBatch(traitName string, n int) Generator {
	return Func(func() interface{} {
		return &relationField{traitName: traitName, n: n}
	})
}

// CreateRelatedCustom is a special Generator that will create related trait
// with additional changes.
func CreateRelatedCustom(traitName string, override Trait) Generator {
	return Func(func() interface{} {
		return &relationField{traitName: traitName, n: 1, override: override}
	})
}

// CreateRelatedCustomBatch is a special Generator that will create
// a batch of related traits with additional changes.
func CreateRelatedCustomBatch(traitName string, n int, override Trait) Generator {
	return Func(func() interface{} {
		return &relationField{traitName: traitName, n: n, override: override}
	})
}
