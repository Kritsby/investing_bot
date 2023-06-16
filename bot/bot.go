package controller

import (
	"context"
	"dev/investing/repository"
	"fmt"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"github.com/vodolaz095/go-investAPI/investapi"
)

var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("help", "help"),
		tgbotapi.NewInlineKeyboardButtonData("register", "register"),
		tgbotapi.NewInlineKeyboardButtonData("token", "token"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("auto", "auto"),
		tgbotapi.NewInlineKeyboardButtonData("sum in rubles", "total_sum_rub"),
		tgbotapi.NewInlineKeyboardButtonData("sum in usds", "total_sum_usd"),
	),
)

type Bot struct {
	repository *repository.Postgres
	tgBot      *tgbotapi.BotAPI
}

func NewTgBot(bot *tgbotapi.BotAPI, repo *repository.Postgres) *Bot {
	return &Bot{
		tgBot:      bot,
		repository: repo,
	}
}

func (b *Bot) Start() {
	updates := b.initUpdate(0)
	b.handleUpdate(updates)
}

func (b *Bot) initUpdate(offset int) tgbotapi.UpdatesChannel {
	b.tgBot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return b.tgBot.GetUpdatesChan(u)
}

type portfolioSum struct {
	totalSum    float64
	procentPlus float64
	sumPlus     float64
	currency    string
}

func (b *Bot) handleUpdate(updates tgbotapi.UpdatesChannel) error {
	for update := range updates {
		if update.Message != nil {
			switch update.Message.Command() {
			case "menu":
				b.replyMenuCommand(update)
			case "start":
				b.replyStartCommand(update)
			default:
				b.replyDefault(update)
			}
		} else if update.CallbackQuery != nil {
			switch update.CallbackQuery.Data {
			case "help":
				b.replyHelpCommand(update)
			case "register":
				b.replyRegisterCommand(update)
			case "token":
				b.replyTokenCommand(update)
			case "total_sum_rub":
				b.replyTotalRubCommand(update)
			case "total_sum_usd":
				b.replyTotalUsdCommand(update)
			case "auto":
				b.replyAutoCommand(update)
			}
		}
	}

	return nil
}

func (b *Bot) replyMenuCommand(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Menu")
	msg.ReplyMarkup = numericKeyboard
	_, err := b.tgBot.Send(msg)
	if err != nil {
		logrus.Errorf("error send message %s\n", err)
	}
	return nil
}

func (b *Bot) replyStartCommand(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "hello")
	_, err := b.tgBot.Send(msg)
	if err != nil {
		logrus.Errorf("error send message %s\n", err)
	}
	return nil
}

func (b *Bot) replyDefault(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "can't understand try /menu")
	msg.ReplyToMessageID = update.Message.MessageID
	_, err := b.tgBot.Send(msg)
	if err != nil {
		logrus.Errorf("error send message %s\n", err)
	}
	return nil
}

func (b *Bot) replyHelpCommand(update tgbotapi.Update) error {
	text := fmt.Sprintln("u can use this commands:\n/register - for registration\n/token - for change ur token\n/total_sum_rub - for receiving the amount in rub\n/total_sum_usd - for receiving the amount in usd")
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
	_, err := b.tgBot.Send(msg)
	if err != nil {
		logrus.Errorf("error send message %s\n", err)
	}

	return nil
}

func (b *Bot) replyRegisterCommand(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "write ur token")
	msg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
	_, err := b.tgBot.Send(msg)
	if err != nil {
		logrus.Errorf("error send message %s\n", err)
	}

	b.newUpdateUser(update)

	msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "г was registered")
	_, err = b.tgBot.Send(msg)
	if err != nil {
		logrus.Errorf("error send message %s\n", err)
	}

	return nil
}

func (b *Bot) newUpdateUser(update tgbotapi.Update) {
	updates2 := b.initUpdate(update.UpdateID + 1)

	for update2 := range updates2 {
		if update2.Message != nil {
			err := b.repository.NewUser(context.Background(), update.CallbackQuery.Message.Chat.ID, update2.Message.Text)
			if err != nil {
				logrus.Errorf("can't create user")
			}
			break
		}
	}
}

func (b *Bot) replyTokenCommand(update tgbotapi.Update) error {
	err := b.repository.SaveToken(context.Background(), update.Message.Text, update.Message.Chat.ID)
	if err != nil {
		logrus.Errorf("can't save token %s", err)
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "can't save token")
		_, err := b.tgBot.Send(msg)
		if err != nil {
			logrus.Errorf("error send message %s\n", err)
			return err
		}
	}
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "token was saved")
	_, err = b.tgBot.Send(msg)
	if err != nil {
		logrus.Errorf("error send message %s\n", err)
		return err
	}
	return nil
}

func (b *Bot) replyTotalRubCommand(update tgbotapi.Update) error {
	token, err := b.repository.TakeToken(context.Background(), update.CallbackQuery.Message.Chat.ID)
	if err != nil {
		return err
	}

	portfolioSum, err := b.takeProfileSum(token, investapi.PortfolioRequest_RUB)
	if err != nil {
		logrus.Errorln("error with taking portfolio sum", err)
		return err
	}

	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("Информация по Брокерскому счету:\n\tTotal Sum: %v %s\n\tTotal plus in procent: %v\n\tTotal plus sum: %v", portfolioSum.totalSum, portfolioSum.currency, portfolioSum.procentPlus, portfolioSum.sumPlus))

	_, err = b.tgBot.Send(msg)
	if err != nil {
		logrus.Errorf("error send message %s\n", err)
	}
	return nil
}

func (b *Bot) replyTotalUsdCommand(update tgbotapi.Update) error {
	token, err := b.repository.TakeToken(context.Background(), update.CallbackQuery.Message.Chat.ID)
	if err != nil {
		return err
	}

	portfolioSum, err := b.takeProfileSum(token, investapi.PortfolioRequest_USD)
	if err != nil {
		logrus.Errorln("error with taking portfolio sum", err)
		return err
	}

	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("Информация по Брокерскому счету:\n\tTotal Sum: %v %s\n\tTotal plus in procent: %v\n\tTotal plus sum: %v", portfolioSum.totalSum, portfolioSum.currency, portfolioSum.procentPlus, portfolioSum.sumPlus))

	_, err = b.tgBot.Send(msg)
	if err != nil {
		logrus.Errorf("error send message %s\n", err)
	}
	return nil
}

func (b *Bot) replyAutoCommand(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "write time in hour or comand /stop")
	msg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
	_, err := b.tgBot.Send(msg)
	if err != nil {
		logrus.Errorf("error send message %s\n", err)
	}

	updates2 := b.initUpdate(update.UpdateID + 1)

	var t int

	for update2 := range updates2 {
		if update2.Message != nil {
			switch update2.Message.Command() {
			case "stop":

			}
			t, err = strconv.Atoi(update2.Message.Text)
			if err != nil {
				logrus.Errorf("can't convert time(string) to time(int) - %s", err)
			}
			logrus.Info(t)
			break
		}
	}

	// Create ticker that ticks every minute
	ticker := time.NewTicker(time.Duration(t) * time.Second)
	defer ticker.Stop()

	// Loop over ticker channel
	for range ticker.C {
		// Create new message with desired chat ID and text
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Hello, world!")

		// Send message using bot
		_, err := b.tgBot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (b *Bot) takeProfileSum(token string, currency investapi.PortfolioRequest_CurrencyRequest) (portfolioSum, error) {
	client, err := investapi.New(token)
	if err != nil {
		return portfolioSum{}, err
	}

	id, err := b.takeIdAccount(client)

	portfolio, err := client.OperationsServiceClient.GetPortfolio(context.Background(), &investapi.PortfolioRequest{
		AccountId: id,
		Currency:  currency,
	})
	if err != nil {
		logrus.Error(err)
	}

	portfolioMoney, err := b.profiles(portfolio)

	return portfolioMoney, err
}

func (b *Bot) takeIdAccount(client *investapi.Client) (string, error) {
	res, err := client.UsersServiceClient.GetAccounts(context.Background(), &investapi.GetAccountsRequest{})
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	var id string

	for _, r := range res.Accounts {
		if r.Name == "Брокерский счёт" {
			id = r.Id
			break
		}
	}

	return id, nil
}

func (b *Bot) profiles(money *investapi.PortfolioResponse) (portfolioSum, error) {
	procentPlus, err := strconv.ParseFloat(fmt.Sprintf("%v.%v", money.ExpectedYield.Units, money.GetExpectedYield().Nano), 64)
	if err != nil {
		logrus.Errorf("can't convert string to float %s\n", err)
		return portfolioSum{}, err
	}

	allSumFloat, err := strconv.ParseFloat(fmt.Sprintf("%v.%v", money.TotalAmountPortfolio.Units, money.TotalAmountPortfolio.Nano), 64)
	if err != nil {
		logrus.Errorf("can't convert string to float %s\n", err)
		return portfolioSum{}, err
	}

	sumPlus := (allSumFloat / 100) * procentPlus

	portfolio := portfolioSum{
		totalSum:    allSumFloat,
		procentPlus: procentPlus,
		sumPlus:     sumPlus,
		currency:    money.TotalAmountPortfolio.Currency,
	}

	return portfolio, nil
}
