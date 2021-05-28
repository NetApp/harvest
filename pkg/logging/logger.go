package logging

import (
	"github.com/rs/zerolog/pkgerrors"
	"goharvest2/pkg/conf"
	"io"
	"os"
	"path"
	"sync"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultLogFileName           string = "harvest.log"
	defaultLogLevel                     = zerolog.InfoLevel
	defaultConsoleLoggingEnabled bool   = true
	// DefaultFileLoggingEnabled is intentionally false to avoid opening many file descriptors for same log file
	defaultFileLoggingEnabled bool = false
	DefaultLogMaxMegaBytes    int  = 10 // 10MB
	DefaultLogMaxBackups      int  = 10
	DefaultLogMaxAge          int  = 30
)

// Configuration for logging
type LogConfig struct {
	// Enable console logging
	ConsoleLoggingEnabled bool
	// Log Level
	LogLevel zerolog.Level
	// Prefix
	PrefixKey   string
	PrefixValue string
	// FileLoggingEnabled makes the framework log to a file
	FileLoggingEnabled bool
	// Directory to log to to when filelogging is enabled
	Directory string
	// Filename is the name of the logfile which will be placed inside the directory
	Filename string
	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int
	// MaxBackups the max number of rolled files to keep
	MaxBackups int
	// MaxAge the max age in days to keep a logfile
	MaxAge int
}

var logger *Logger

var once sync.Once

type Logger struct {
	*zerolog.Logger
}

// Get Returns a logger with default configuration if logger is not initialized yet
// default configuration only writes to console and not to file see param defaultFileLoggingEnabled
func Get() *Logger {
	once.Do(func() {
		if logger == nil {
			defaultPrefixKey := "harvest"
			defaultPrefixValue := "harvest"
			logConfig := LogConfig{ConsoleLoggingEnabled: defaultConsoleLoggingEnabled,
				PrefixKey:          defaultPrefixKey,
				PrefixValue:        defaultPrefixValue,
				LogLevel:           defaultLogLevel,
				FileLoggingEnabled: defaultFileLoggingEnabled,
				Directory:          conf.GetHarvestLogPath(),
				Filename:           defaultLogFileName,
				MaxSize:            DefaultLogMaxMegaBytes,
				MaxBackups:         DefaultLogMaxBackups,
				MaxAge:             DefaultLogMaxAge}
			logger = Configure(logConfig)
		}
	})
	return logger
}

// SubLogger adds the field key with val as a string to the logger context and returns sublogger
func SubLogger(key string, value string) *Logger {
	logger := Get().With().Str(key, value).Logger()
	subLogger := &Logger{
		Logger: &logger,
	}
	return subLogger
}

// Configure sets up the logging framework
func Configure(config LogConfig) *Logger {
	var writers []io.Writer

	if config.ConsoleLoggingEnabled {
		// writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
		writers = append(writers, os.Stdout)
	}
	if config.FileLoggingEnabled {
		writers = append(writers, newRollingFile(config))
	}
	multiWriters := zerolog.MultiLevelWriter(writers...)

	zerolog.SetGlobalLevel(config.LogLevel)
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zeroLogger := zerolog.New(multiWriters).With().Str(config.PrefixKey, config.PrefixValue).Timestamp().Logger()

	zeroLogger.Debug().
		Bool("consoleLoggingEnabled", config.ConsoleLoggingEnabled).
		Bool("fileLogging", config.FileLoggingEnabled).
		Str("loglevel", config.LogLevel.String()).
		Str("prefixKey", config.PrefixKey).
		Str("prefixValue", config.PrefixValue).
		Str("logDirectory", config.Directory).
		Str("fileName", config.Filename).
		Int("maxSizeMB", config.MaxSize).
		Int("maxBackups", config.MaxBackups).
		Int("maxAgeInDays", config.MaxAge).
		Msg("logging configured")

	logger = &Logger{
		Logger: &zeroLogger,
	}
	return logger
}

//returns lumberjack writer
func newRollingFile(config LogConfig) io.Writer {
	return &lumberjack.Logger{
		Filename:   path.Join(config.Directory, config.Filename),
		MaxBackups: config.MaxBackups, // files
		MaxSize:    config.MaxSize,    // megabytes
		MaxAge:     config.MaxAge,     // days
		Compress:   true,
	}
}

// GetZerologLevel returns log level mapping
func GetZerologLevel(logLevel int) zerolog.Level {
	switch logLevel {
	case 0:
		return zerolog.TraceLevel
	case 1:
		return zerolog.DebugLevel
	case 2:
		return zerolog.InfoLevel
	case 3:
		return zerolog.WarnLevel
	case 4:
		return zerolog.ErrorLevel
	case 5:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}
