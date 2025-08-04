package eventmanager

import (
	"os"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	validator "github.com/go-playground/validator/v10"
	"github.com/spf13/cast"
)

var (
	app *logmanager.Application
	val *validator.Validate

	channels           = 1
	maxChannels        = 10
	delay              = 2 * time.Second
	prefetchCount      = 10
	refreshDelay       = time.Hour
	maxRetry      uint = 5
	retryDelay         = time.Minute
)

func setVars() {
	if val == nil {
		val = validator.New(validator.WithRequiredStructEnabled())
	}
	if os.Getenv("RABBITMQ_CHANNELS") != "" {
		channels = min(cast.ToInt(os.Getenv("RABBITMQ_CHANNELS")), maxChannels)
	}
	if os.Getenv("RABBITMQ_RECONNECT_DELAY") != "" {
		delayDuration, _ := time.ParseDuration(os.Getenv("RABBITMQ_RECONNECT_DELAY"))
		if delayDuration > 0 {
			delay = delayDuration
		}
	}
	if os.Getenv("RABBITMQ_PREFETCH_COUNT") != "" {
		prefetchCount = cast.ToInt(os.Getenv("RABBITMQ_PREFETCH_COUNT"))
	}
	if os.Getenv("RABBITMQ_REFRESH_DELAY") != "" {
		refreshDelayDuration, _ := time.ParseDuration(os.Getenv("RABBITMQ_REFRESH_DELAY"))
		if refreshDelayDuration > 0 {
			refreshDelay = refreshDelayDuration
		}
	}
	if os.Getenv("RABBITMQ_MAX_RETRY") != "" {
		maxRetry = cast.ToUint(os.Getenv("RABBITMQ_MAX_RETRY"))
	}
	if os.Getenv("RABBITMQ_RETRY_DELAY") != "" {
		retryDelayDuration, _ := time.ParseDuration(os.Getenv("RABBITMQ_RETRY_DELAY"))
		if retryDelayDuration > 0 {
			retryDelay = retryDelayDuration
		}
	}
}
