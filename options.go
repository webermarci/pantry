package pantry

import "time"

type Options struct {
	CleaningInterval     time.Duration
	PersistenceDirectory string
}
