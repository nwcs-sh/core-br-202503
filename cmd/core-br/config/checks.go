package config

// checkConfig checks the config for validity.
func checkConfig() error {
	switch config.Logging.Level {
	case "debug", "info", "warn", "error", "fatal":
		// no-op
	default:
		return ErrLoggingLevelInvalid
	}

	switch config.Logging.Encoding {
	case "json", "console":
		// no-op
	default:
		return ErrLoggingEncodingInvalid
	}

	return nil
}
