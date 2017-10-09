package seedr

// Relation defines Factory relation.
// Can be defined using one of following functions
// inside Relations map:
//    - BelongsTo
//    - HasMany
//    - HasManyThrough
type Relation struct {
	kind      int
	factory   string
	lfield    string
	rfield    string
	joinTrait string
}

const (
	relationM2M int = iota
	relationChild
	relationParent
)

// BelongsTo defines "belogs to" relation.
// By default joinField is eql to relation name
// (key of this relation in Relations map of Factory)
// Example:
//   Relations{
//     // with custom name
//     "author": BelongsTo("users", "author_id"),
//     // with same name as field
//     "author_id": BelongsTo("users"),
//   },
// It defines author_id column as FK to users.id
func BelongsTo(factory string, joinField ...string) *Relation {
	return &Relation{
		kind:    relationParent,
		factory: factory,
		lfield:  append(joinField, "")[0],
		rfield:  "",
	}
}

// HasMany declares relation.
// Example:
//   // Relations of "users" factory
//   Relations{
//     "articles": HasMany("articles", "author_id"),
//   },
// It will join on articles.author_id = user.id
func HasMany(factory, foreignKey string) *Relation {
	return &Relation{
		kind:    relationChild,
		factory: factory,
		lfield:  foreignKey,
		rfield:  "",
	}
}

// HasManyThrough defines "M2M" relation through joinTrait, where
// lfield is a foreign key for this factory, and rfield
// is a foreignKey for related factory.
// First parameter is actually public trait name
// as opposed to factory name in  BelongsTo and HasMany.
// Example:
//   Relations{
//     "users": HasManyThrough("ClubToUser", "club_id", "user_id"),
//   },
func HasManyThrough(joinTrait, lfield, rfield string) *Relation {
	return &Relation{
		kind:      relationM2M,
		factory:   "", // FIXME: unused (practically)
		joinTrait: joinTrait,
		lfield:    lfield,
		rfield:    rfield,
	}
}

// Relations defines relations of a Factory.
// Example:
//   Relations{
//     "author": BelongsTo("users", "author_id"),
//     "comments": HasMany("comments", "article_id"),
//     "commenters": HasManyThrough("Comment", "article_id", "user_id"),
//   },
type Relations map[string]*Relation

func (r Relations) normalize() Relations {
	for name, rel := range r {
		switch rel.kind {
		case relationParent:
			if rel.lfield == "" {
				rel.lfield = name
			}
		case relationM2M:
			if rel.factory == "" {
				rel.factory = name
			}
		}
	}
	return r
}
