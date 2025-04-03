package config

// GetConfig returns config
func GetConfig() *Config {
	return &config
}

// GetName returns app name
func GetName() string {
	return config.Name
}

// GetVersion returns app version
func GetVersion() string {
	return config.Version
}

// GetPostgresHost returns postgres host
func GetPostgresHost() string {
	return config.Postgres.Host
}

// GetPostgresPort returns postgres port
func GetPostgresPort() uint16 {
	return config.Postgres.Port
}

// GetPostgresUsername returns postgres user
func GetPostgresUsername() string {
	return config.Postgres.Username
}

// GetPostgresPassword returns postgres user password
func GetPostgresPassword() string {
	return config.Postgres.Password
}

// GetPostgresDatabase returns postgres database to connect to
func GetPostgresDatabase() string {
	return config.Postgres.Database
}

// GetLoggingPath returns logging path
func GetLoggingPath() string {
	return config.Logging.Path
}
