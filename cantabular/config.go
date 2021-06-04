package cantabular

import (
	"time"
)

// Config holds the config used to initialise the Cantabular Client
type Config struct{
	Timeout time.Duration
	Host    string
}