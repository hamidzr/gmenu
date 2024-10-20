package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

type StderrHook struct{}

func (h *StderrHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel}
}

func (h *StderrHook) Fire(entry *logrus.Entry) error {
	entry.Logger.Out = os.Stderr
	return nil
}

func SetupLogger() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.TraceLevel)
	hook := &StderrHook{}
	logrus.AddHook(hook)
}
