package logging

import (
	"cmp"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
)

const (
	defaultLogFileName           = "harvest.log"
	defaultConsoleLoggingEnabled = true
	defaultFileLoggingEnabled    = false // false to avoid opening many file descriptors for same log file
	DefaultLogMaxMegaBytes       = 10    // 10 MB
	DefaultLogMaxBackups         = 5
	DefaultLogMaxAge             = 7
)

// LogConfig defines the configuration for logging
type LogConfig struct {
	// Enable console logging
	ConsoleLoggingEnabled bool
	// Log Level
	LogLevel slog.Level
	// Prefix
	PrefixKey   string
	PrefixValue string
	// FileLoggingEnabled makes the framework log to a file
	FileLoggingEnabled bool
	// Directory to log to when filelogging is enabled
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

var (
	logger *slog.Logger
	once   sync.Once
)

// Get Returns a logger with default configuration if logger is not initialized yet
// default configuration only writes to console and not to file see param defaultFileLoggingEnabled
func Get() *slog.Logger {
	once.Do(func() {
		if logger == nil {
			defaultPrefixKey := "harvest"
			defaultPrefixValue := "harvest"
			logConfig := LogConfig{
				ConsoleLoggingEnabled: defaultConsoleLoggingEnabled,
				PrefixKey:             defaultPrefixKey,
				PrefixValue:           defaultPrefixValue,
				LogLevel:              slog.LevelInfo,
				FileLoggingEnabled:    defaultFileLoggingEnabled,
				Directory:             GetLogPath(),
				Filename:              defaultLogFileName,
				MaxSize:               DefaultLogMaxMegaBytes,
				MaxBackups:            DefaultLogMaxBackups,
				MaxAge:                DefaultLogMaxAge,
			}
			logger = Configure(logConfig)
		}
	})
	return logger
}

func GetLogPath() string {
	return cmp.Or(os.Getenv("HARVEST_LOGS"), "/var/log/harvest/")
}

// Configure sets up the logging framework
func Configure(config LogConfig) *slog.Logger {

	handlerOptions := &slog.HandlerOptions{
		AddSource: true,
		Level:     config.LogLevel,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source, ok := a.Value.Any().(*slog.Source)
				if !ok {
					return a
				}
				if source != nil {
					source.File = filepath.Base(source.File)
					a.Value = slog.StringValue(source.File + ":" + strconv.Itoa(source.Line))
				}
			}
			return a
		},
	}

	var (
		handlers []slog.Handler
		aLogger  *slog.Logger
	)

	if config.ConsoleLoggingEnabled {
		handlers = append(handlers, slog.NewTextHandler(os.Stderr, handlerOptions))
	}

	if config.FileLoggingEnabled {
		handlers = append(handlers, slog.NewJSONHandler(newRollingFile(config), handlerOptions))
	}

	if len(handlers) == 1 {
		aLogger = slog.New(handlers[0])
	} else {
		aLogger = slog.New(MultiHandler(handlers...))
	}

	if config.PrefixKey != "" {
		aLogger = aLogger.With(slog.String(config.PrefixKey, config.PrefixValue))
	}

	slog.SetDefault(aLogger)

	return aLogger
}

// returns lumberjack writer
func newRollingFile(config LogConfig) io.Writer {
	return &lumberjack.Logger{
		Filename:   path.Join(config.Directory, config.Filename),
		MaxBackups: config.MaxBackups, // files
		MaxSize:    config.MaxSize,    // megabytes
		MaxAge:     config.MaxAge,     // days
		Compress:   true,
	}
}

// GetLogLevel returns log level mapping
func GetLogLevel(logLevel int) slog.Level {
	switch logLevel {
	case 0:
		return slog.LevelDebug
	case 1:
		return slog.LevelDebug
	case 2:
		return slog.LevelInfo
	case 3:
		return slog.LevelWarn
	case 4:
		return slog.LevelError
	case 5:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
