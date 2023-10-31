package state

import (
	"ebs-bootstrap/internal/config"
)

type State interface {
	Pull() (error)
	Diff(c *config.Config) (error)
	Push(c *config.Config) (error)
}
