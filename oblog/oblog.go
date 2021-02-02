package oblog

import (
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

// Log global log object
var log = logrus.New()

// Tracef Tracef
func Tracef(format string, args ...interface{}) {
	log.Logf(logrus.TraceLevel, format, args...)
}

// Debugf Debugf
func Debugf(format string, args ...interface{}) {
	log.Logf(logrus.DebugLevel, format, args...)
}

// Infof Infof
func Infof(format string, args ...interface{}) {
	log.Logf(logrus.InfoLevel, format, args...)
}

// Warnf Warnf
func Warnf(format string, args ...interface{}) {
	log.Logf(logrus.WarnLevel, format, args...)
}

// Errorf Errorf
func Errorf(format string, args ...interface{}) {
	log.Logf(logrus.ErrorLevel, format, args...)
}

// Fatalf Fatalf
func Fatalf(format string, args ...interface{}) {
	log.Logf(logrus.FatalLevel, format, args...)
	log.Exit(1)
}

// Panicf Panicf
func Panicf(format string, args ...interface{}) {
	log.Logf(logrus.PanicLevel, format, args...)
}

func newLfsHook(logLevel string, maxRemainCnt uint) logrus.Hook {
	writer, err := rotatelogs.New(
		"logs/vgpay-%Y%m%d.log",
		rotatelogs.WithLinkName("vgpay.log"),

		rotatelogs.WithRotationTime(time.Hour*24),

		// WithRotationCount.
		rotatelogs.WithRotationCount(maxRemainCnt),
	)

	if err != nil {
		log.Errorf("config local file system for logger error: %v", err)
	}

	level, err := logrus.ParseLevel(logLevel)

	if err == nil {
		log.SetLevel(level)
	} else {
		log.SetLevel(logrus.WarnLevel)
	}

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, &logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: false,
	})

	return lfsHook
}

// Init init
func Init() {
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006.01.02 03:04:05",
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)

	hook := newLfsHook("trace", 365*3)
	log.AddHook(hook)
}
