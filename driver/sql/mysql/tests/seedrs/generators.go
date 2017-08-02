package seedrs

import "time"

// TimeSequence implements seedr.Generator,
// and it is an example of custom generator
type TimeSequence struct {
	current time.Time
}

func (ts *TimeSequence) Next() interface{} {
	ts.current = ts.current.Add(time.Hour)
	return ts.current.Add(-time.Hour).UTC()
}
