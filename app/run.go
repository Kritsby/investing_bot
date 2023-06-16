package app

import (
	"dev/investing/bot"
	"dev/investing/config"
	"dev/investing/driver"
	"dev/investing/repository"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func Run(cfg config.Config) error {
	//Init postgresql
	pool, err := driver.NewPostgres(cfg.PSQL)
	if err != nil {
		return fmt.Errorf("error with init repository %w", err)
	}
	defer pool.Close()

	//Init controller Api
	botApi, err := tgbotapi.NewBotAPI(cfg.TgBot.Token)
	if err != nil {
		return fmt.Errorf("can't create controller api: %w", err)
	}
	logrus.Infof("Authorized on account %s\n", botApi.Self.UserName)
	defer botApi.StopReceivingUpdates()

	//Init inject
	db := repository.New(pool)
	TgBot := controller.NewTgBot(botApi, db)

	go func() {
		TgBot.Start()
	}()

	//Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Info("\nApp Shutting Down")

	return nil
}
