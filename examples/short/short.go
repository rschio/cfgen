package short

import (
	"time"
)

type Config struct {
	APIKey      string
	APIEndpoint string
	Headers     map[string]string
	NowFn       func() time.Time
}
