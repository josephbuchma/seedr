package seedr

// Relation defines Factory relation.
// Can be defined using one of:
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
func BelongsTo(factory string, joinField ...string) *Relation {
	return &Relation{
		kind:    relationParent,
		factory: factory,
		lfield:  append(joinField, "")[0],
		rfield:  "",
	}
}

// HasMany defines "has many" relation.
func HasMany(factory, foreignKey string) *Relation {
	return &Relation{
		kind:    relationChild,
		factory: factory,
		lfield:  foreignKey,
		rfield:  "",
	}
}

// HasManyThrough defines "M2M" relation through joinFactory, where
// lfield is a foreign key on this factory, and rfield
// is a foreignKey on related factory.
// By default relatedFactory is eql to relation name
// (key of this relation in Relations map of Factory)
func HasManyThrough(joinTrait, lfield, rfield string, relatedFactory ...string) *Relation {
	return &Relation{
		kind:      relationM2M,
		factory:   append(relatedFactory, "")[0],
		joinTrait: joinTrait,
		lfield:    lfield,
		rfield:    rfield,
	}
}

// Relations defines relations of the Factory
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
