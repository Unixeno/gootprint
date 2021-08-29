package main

import (
	"github.com/sirupsen/logrus"
)

func init() {
	formatter := new(logrus.TextFormatter)
	formatter.DisableTimestamp = true
	formatter.DisableLevelTruncation = true
	logrus.SetFormatter(formatter)
}
