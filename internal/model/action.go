package model

import (
	"fmt"
)

type Mode string

const (
	Empty       Mode = ""
	Healthcheck Mode = "healthcheck"
	Prompt      Mode = "prompt"
	Force       Mode = "force"
)

func ParseMode(s string) (Mode, error) {
	m := Mode(s)
	switch m {
	case Empty, Healthcheck, Prompt, Force:
		return m, nil
	default:
		return m, fmt.Errorf("🔴 Mode '%s' is not supported", s)
	}
}
