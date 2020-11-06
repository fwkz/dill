package operations

import log "github.com/sirupsen/logrus"

type utcFormatter struct {
	log.Formatter
}

func (u utcFormatter) Format(e *log.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}

func SetupLogging() {
	log.SetFormatter(utcFormatter{&log.JSONFormatter{}})
	log.SetLevel(log.DebugLevel)
}
