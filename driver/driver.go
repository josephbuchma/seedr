package driver

// Payload for Driver.MustCreate
type Payload struct {
	// Entity is a name of database schema/table, index, etc.
	// It may not be required, depends on backend / driver implementation.
	Entity string
	// PrimaryKey is optional, depends on backend / driver implementation.
	PrimaryKey string
	// InsertFields is a list of fields that must be stored
	// (e.g. in SQL database, it could be all field except `id`, which is assigned automatically)
	InsertFields []string
	// ReturnFields is a list of fields that will be returned by Create
	// (e.g. in SQL database, it would probably be all fields, including `id` that was autogenerated)
	ReturnFields []string
	// Data is a list of objects to be created
	Data []map[string]interface{}
}

// Driver holds implementation of particular storage behind
type Driver interface {
	// Create stores data and returns results that are containing ReturnFields only
	// if something goes wrong it should return meaningful error.
	Create(Payload) (results []map[string]interface{}, err error)
}
