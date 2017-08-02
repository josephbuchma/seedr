package seedrs

import (
	"time"

	. "github.com/josephbuchma/seedr"
	"github.com/josephbuchma/seedr/driver/sql/mysql/tests/util"

	_ "github.com/go-sql-driver/mysql"
)

var _ = TestSeedr.Add("users", Factory{
	FactoryConfig{
		Entity:     "users",
		PrimaryKey: "id",
	},
	Relations{
		"articles": HasMany("articles", "author_id"),
	},
	Traits{
		"basic": {
			"id":         Auto(),
			"name":       SequenceString("Agent Smith %d"),
			"email":      SequenceString("agentsmith-%d@gmail.com"),
			"active":     true,
			"checkin":    time.Now(),
			"created_at": time.Now(),
		},
		"inactive": {
			"active": false,
		},
		"withTestTime": {
			"created_at": util.TestTime,
			"checkin":    util.TestTime,
		},

		// Public:

		"User": {
			Include: "basic",
		},
		"TestUser": {
			Include: "basic withTestTime",
		},
		"InactiveUser": {
			Include: "TestUser inactive",
		},
		"UserJohn": {
			Include: "TestUser",
			"name":  "John",
		},
		"UserHeavyWriter": {
			Include:    "TestUser withTestTime",
			"articles": CreateRelatedBatch("Article", 2),
		},
	},
})
