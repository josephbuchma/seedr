package seedrs

import (
	"time"

	. "github.com/josephbuchma/seedr"
)

var _ = TestSeedr.Add("hellota_fields", Factory{
	FactoryConfig{
		Entity:     "hellota_fields",
		PrimaryKey: "id",
	},
	Relations{},
	Traits{
		"HellotaFieldsTest": {
			"id": Auto(),
			"a":  SequenceInt(),
			"b":  DummyText(400),
			"c":  SequenceString("seq-%s"),
			"d":  time.Now(),
			"e":  true,
			"f":  SequenceInt(),
			"g":  DummyText(400),
			"h":  SequenceString("seq-%s"),
			"i":  time.Now(),
			"j":  true,
			"k":  SequenceInt(),
			"l":  DummyText(400),
			"m":  SequenceString("seq-%s"),
			"n":  time.Now(),
			"o":  true,
			"p":  SequenceInt(),
			"q":  DummyText(400),
			"r":  SequenceString("seq-%s"),
			"s":  time.Now(),
			"t":  true,
			"u":  SequenceInt(),
			"v":  DummyText(400),
			"w":  SequenceString("seq-%s"),
			"x":  time.Now(),
			"y":  true,
			"z":  SequenceInt(),
		},
	},
})
