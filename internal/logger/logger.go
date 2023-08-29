package logger

import "github.com/sirupsen/logrus"


func ConfigLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	SetLogLevel(logLevel, logger)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableQuote:     true,
		TimestampFormat: "15:04:05",
		FullTimestamp:   true, 
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyMsg: "msg", // Переименовываем поле сообщения в "message"
		},
	})
	return logger
}

func SetLogLevel(level string, logger *logrus.Logger) {
    switch level {
    case "debug":
        logger.SetLevel(logrus.DebugLevel)
    case "info":
        logger.SetLevel(logrus.InfoLevel)
    case "warn":
        logger.SetLevel(logrus.WarnLevel)
    case "error":
        logger.SetLevel(logrus.ErrorLevel)
    default:
        logger.SetLevel(logrus.InfoLevel) // По умолчанию
    }
}