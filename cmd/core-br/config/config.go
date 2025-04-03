package config

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

var config Config

// Config is the CLI options wrapped in a struct
type Config struct {
	Name       string
	Version    string
	ConfigFile string

	Postgres Postgres `yaml:"postgres" json:"postgres"`
	Logging  Logging  `yaml:"logging" json:"logging"`
}

// Postgres postgres configuration options
type Postgres struct {
	Host     string `yaml:"host" json:"host"`
	Port     uint16 `yaml:"port" json:"port"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Database string `yaml:"database" json:"database"`
}

// Logging logging definition
type Logging struct {
	Level    string `yaml:"level" json:"level"`
	Path     string `yaml:"path" json:"path"`
	Encoding string `yaml:"encoding" json:"encoding"`
}

// NewConfig returns a Config struct that can be used to reference configuration
// NewConfig does the following:
//   - Runs initCLI (sets and read CLI switches)
//   - Runs initConfig (reads config from files)
func NewConfig(name, version string) error {
	config.Name = name
	config.Version = version

	// Setup CLI flags
	initCLI(&config)

	// initialize environment
	if err := initConf(config.ConfigFile); err != nil {
		return fmt.Errorf("config.NewConfig: %w", err)
	}

	// initialize logging
	logfile := fmt.Sprintf("%s/core-br.log", GetLoggingPath())

	if err := initLogging(logfile, true); err != nil {
		return err
	}

	// check config
	if err := checkConfig(); err != nil {
		return err
	}

	return nil
}

// initCLI initializes CLI switches
func initCLI(config *Config) {
	flag.StringVar(&config.ConfigFile, "config", "./config.yaml", "Path for the config.yaml configuration file")
	flag.Parse()
}

// initConf initializes the configuration
func initConf(cfgFile string) error {
	guestletData, err := os.ReadFile(cfgFile)
	if err != nil {
		return fmt.Errorf("os.ReadFile failed: %w", err)
	}

	if err1 := yaml.Unmarshal(guestletData, &config); err1 != nil {
		return fmt.Errorf("yaml.Unmarshal failed: %w", err1)
	}

	config.ConfigFile = cfgFile

	return nil
}

// initLogging initializes logging
func initLogging(logfile string, stdout bool) error {
	var stdoutPaths []string
	var stderrPaths []string

	if stdout {
		stdoutPaths = []string{logfile, "stdout"}
		stderrPaths = []string{logfile, "stderr"}
	} else {
		stdoutPaths = []string{logfile}
		stderrPaths = []string{logfile}
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	zapConfig := &zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Encoding:         config.Logging.Encoding,
		OutputPaths:      stdoutPaths,
		ErrorOutputPaths: stderrPaths,
		InitialFields: map[string]interface{}{
			"hostname": hostname,
			"pid":      os.Getpid(),
		},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "name",
			CallerKey:      "caller",
			FunctionKey:    "function",
			MessageKey:     "message",
			StacktraceKey:  "stack",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	switch config.Logging.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "fatal":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	}

	zapLogger := zap.Must(zapConfig.Build())
	zap.ReplaceGlobals(zapLogger)

	return nil
}

// GetPGConn gets pg connection (you must close it)
func GetPGConn() (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=disable", config.Postgres.Username, config.Postgres.Password, config.Postgres.Host, config.Postgres.Port)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
