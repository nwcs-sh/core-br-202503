package config

import "errors"

var (
	// logging
	ErrLoggingLevelInvalid    = errors.New("logging.level is invalid, values: debug, info, warn, error, panic")
	ErrLoggingEncodingInvalid = errors.New("logging.encoding must be either json or console")
)
