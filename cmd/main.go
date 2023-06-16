package main

import (
	"dev/investing/app"
	"dev/investing/config"
	"github.com/sirupsen/logrus"
)

var cfg config.Config

func init() {
	err := cfg.InitCfg()
	if err != nil {
		logrus.Fatalln("error with parse config file:", err)
	}
}

func main() {
	err := app.Run(cfg)
	if err != nil {
		logrus.Errorln(err.Error())
	}
}
