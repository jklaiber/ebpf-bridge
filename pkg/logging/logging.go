package logging

import "github.com/sirupsen/logrus"

var DefaultLogger = InitializeDefaultLogger()

func InitializeDefaultLogger() (logger *logrus.Logger) {
	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	})
	return
}
