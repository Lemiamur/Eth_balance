package log

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("Error loading .env file: %v", err)
	}

	Logger = logrus.New()

	if os.Getenv("ENV") == "production" {
		Logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		Logger.Warnf("Invalid LOG_LEVEL value: %v. Defaulting to InfoLevel.", err)
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	Logger.SetOutput(os.Stdout)
}
