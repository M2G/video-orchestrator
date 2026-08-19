package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func New(level logrus.Level) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(level)
	return log
}
