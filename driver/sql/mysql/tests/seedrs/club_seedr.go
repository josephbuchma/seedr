package seedrs

import . "github.com/josephbuchma/seedr"

var _ = TestSeedr.Add("clubs", Factory{
	FactoryConfig{
		Entity:     "clubs",
		PrimaryKey: "id",
	},
	Relations{
		"users": HasManyThrough("ClubToUser", "club_id", "user_id"),
	},
	Traits{
		"basic": {
			"id":   Auto(),
			"name": SequenceString("Club-%d"),
		},

		"withUsers": {
			Include: "basic",
		},

		"Club": {
			Include: "basic",
		},

		"ClubWithUsers": {
			Include: "basic",
			"users": CreateRelatedBatch("TestUser", 2),
		},
	},
}).Add("clubs_to_users", Factory{
	FactoryConfig{
		Entity:     "clubs_to_users",
		PrimaryKey: "id",
	},
	Relations{
		"club_id": BelongsTo("clubs", "id"),
		"user_id": BelongsTo("users", "id"),
	},
	Traits{
		"ClubToUser": {
			"club_id": CreateRelated("Club"),
			"user_id": CreateRelated("User"),
		},
	},
})
