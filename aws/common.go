package aws

import (
	"aws-tutorial/core/config"
	"aws-tutorial/core/logger"

	"go.uber.org/fx"
)

type Params struct {
	fx.In

	Logger logger.Logger
	Config *config.Configuration
}
