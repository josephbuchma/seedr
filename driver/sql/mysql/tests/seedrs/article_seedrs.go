package seedrs

import (
	"time"

	. "github.com/josephbuchma/seedr"
	"github.com/josephbuchma/seedr/driver/sql/mysql/tests/util"
)

var _ = TestSeedr.Add("articles", Factory{
	FactoryConfig{
		Entity:     "articles",
		PrimaryKey: "id",
	},
	Relations{
		"author": BelongsTo("users", "author_id"),
	},
	Traits{
		"basic": {
			"id":         Auto(),
			"author_id":  nil,
			"title":      SequenceString("Awesome Title %d"),
			"body":       "test body value",
			"created_at": time.Now(),
		},
		"withTestAuthor": {
			"author": CreateRelated("TestUser"),
		},

		"withTestTime": {
			"created_at": util.TestTime,
		},

		"Article": {
			Include: "basic",
		},

		"TestArticle": {
			Include: "basic withTestAuthor withTestTime",
		},
	},
})
