package compliance

import "errors"

var (
	ErrTextMismatch        = errors.New("compliance text mismatch")
	ErrAlreadyAcknowledged = errors.New("compliance already acknowledged")
)
