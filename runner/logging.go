package runner

import (
	"io"
	"net/http"
	"os"
	"path"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	// Enable console logging
	ConsoleLoggingEnabled bool
	// Enable file logging
	FileLoggingEnabled bool
	// Directory to log to to when filelogging is enabled
	Directory string
	// Filename is the name of the logfile which will be placed inside the directory
	FileName string
}

type Logger struct {
	*zerolog.Logger
}

// This type implements the http.RoundTripper interface
type LoggingRoundTripper struct {
	Proxied http.RoundTripper
}

var Zlog *Logger

func ConfigureLogger(config *LogConfig) *Logger {
	var writers []io.Writer
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	if config.ConsoleLoggingEnabled {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if config.FileLoggingEnabled {
		writers = append(writers, newRollingFile(config))
	}
	mw := io.MultiWriter(writers...)

	logger := zerolog.New(mw).With().Timestamp().Logger()

	logger.Info().
		Str("logDir", config.Directory).
		Str("logFile", config.FileName).
		Msg("Logging configured")

	return &Logger{
		Logger: &logger,
	}
}

func newRollingFile(config *LogConfig) io.Writer {
	return &lumberjack.Logger{
		Filename:   path.Join(config.Directory, config.FileName),
		MaxBackups: 0, // files
		MaxAge:     0, // days
	}
}

func (lrt LoggingRoundTripper) RoundTrip(req *http.Request) (res *http.Response, e error) {
	// Do "before sending requests" actions here.
	Zlog.Info().Msgf("Sending request to %v\n", req.URL)

	// Send the request, get the response (or the error)
	res, e = lrt.Proxied.RoundTrip(req)

	// Handle the result.
	if e != nil {
		Zlog.Error().Msgf("Error: %v", e)
	} else {
		Zlog.Info().Msgf("Received %v response\n", res.Status)
	}

	return
}
