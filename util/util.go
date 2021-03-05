package util

import (
	"coriolis-ovm-exporter/config"
	"fmt"
	"io"
	"os"
	"path"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// GetLoggingWriter returns a new io.Writer suitable for logging.
func GetLoggingWriter(cfg *config.Config) (io.Writer, error) {
	var writer io.Writer = os.Stdout
	if cfg.LogFile != "" {
		dirname := path.Dir(cfg.LogFile)
		if _, err := os.Stat(dirname); err != nil {
			if os.IsNotExist(err) == false {
				return nil, fmt.Errorf("failed to create log folder")
			}
			if err := os.MkdirAll(dirname, 0o711); err != nil {
				return nil, fmt.Errorf("failed to create log folder")
			}
		}
		writer = &lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   //days
			Compress:   true, // disabled by default
		}
	}
	return writer, nil
}
