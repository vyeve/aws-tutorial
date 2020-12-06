package aws

import (
	"math/rand"
	"time"

	"go.uber.org/fx"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var Module = fx.Provide(
	New,
)
