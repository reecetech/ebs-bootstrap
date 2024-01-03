package layer

import (
	"log"
	"math"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/config"
)

const (
	DisabledWarning = ""
)

type Layer interface {
	From(config *config.Config) error
	Modify(c *config.Config) ([]action.Action, error)
	Validate(config *config.Config) error
	Warning() string
}

type LayerExecutor interface {
	Execute(layers []Layer) error
}

type ExponentialBackoffLayerExecutor struct {
	backoff        backoff.BackOff
	actionExecutor action.ActionExecutor
	config         *config.Config
}

type ExponentialBackoffParameters struct {
	InitialInterval time.Duration
	Multiplier      uint32
	MaxRetries      uint32
}

func DefaultExponentialBackoffParameters() *ExponentialBackoffParameters {
	return &ExponentialBackoffParameters{
		InitialInterval: 200 * time.Millisecond,
		Multiplier:      2,
		MaxRetries:      3,
	}
}

func NewExponentialBackoffLayerExecutor(c *config.Config, ae action.ActionExecutor, ebp *ExponentialBackoffParameters) *ExponentialBackoffLayerExecutor {
	// Cast Multiplier and MaxRetries to float64 for use in the backoff calculation
	m := float64(ebp.Multiplier)
	mr := float64(ebp.MaxRetries)

	// Create Backoff
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = ebp.InitialInterval
	// Disable randomisation of the calculation of the next desired backoff duration
	bo.RandomizationFactor = 0
	// Set the multiplier for the exponential backoff, using the square root of the provided Multiplier.
	bo.Multiplier = math.Sqrt(m)
	// Calculate the maximum elapsed time for the backoff strategy based on the provided MaxRetries and InitialInterval.
	// This formula calculates the maximum elapsed time as a geometric series sum for a given number of retries and interval.
	bo.MaxElapsedTime = time.Duration((math.Pow(m, mr)-1)/(m-1)) * ebp.InitialInterval
	bo.MaxInterval = ebp.InitialInterval * time.Duration(math.Pow(m, mr-1))
	return &ExponentialBackoffLayerExecutor{
		backoff:        bo,
		actionExecutor: ae,
		config:         c,
	}
}

func (le *ExponentialBackoffLayerExecutor) Execute(layers []Layer) error {
	for _, layer := range layers {
		err := layer.From(le.config)
		if err != nil {
			return err
		}
		actions, err := layer.Modify(le.config)
		if err != nil {
			return err
		}
		// Only print warning if actions are detected and a valid warning
		// message is provided
		if warning := layer.Warning(); len(actions) > 0 && warning != DisabledWarning {
			log.Printf("🟠 %s", warning)
		}
		err = le.actionExecutor.Execute(actions)
		if err != nil {
			return err
		}
		// Reset exponential backoff timer
		le.backoff.Reset()
		err = backoff.Retry(func() error {
			return le.validate(layer)
		}, le.backoff)
		if err != nil {
			return err
		}
	}
	log.Println("🟢 Passed all validation checks")
	return nil
}

func (le *ExponentialBackoffLayerExecutor) validate(layer Layer) error {
	// Any potential errors that arise from ingesting the configuration
	// are most likely persistent. Therefore, we wrap any errors produced
	// from layer.From() as a backoff.Permanent so that it can bypass the
	// exponential backoff algorithm
	err := layer.From(le.config)
	if err != nil {
		return backoff.Permanent(err)
	}
	return layer.Validate(le.config)
}
