package telegram

import (
	"errors"
	"strings"
	"sync"

	"github.com/lack-io/cli"
	tgbotapi "gopkg.in/telegram-bot-api.v4"

	"github.com/lack-io/vine/agent/input"
)

type telegramInput struct {
	sync.Mutex

	debug     bool
	token     string
	whitelist []string

	api *tgbotapi.BotAPI
}

type ChatType string

const (
	Private    ChatType = "private"
	Group      ChatType = "group"
	Supergroup ChatType = "supergroup"
)

func init() {
	input.Inputs["telegram"] = &telegramInput{}
}

func (ti *telegramInput) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "telegram-debug",
			EnvVars: []string{"VINE_TELEGRAM_DEBUG"},
			Usage:   "Telegram debug output",
		},
		&cli.StringFlag{
			Name:    "telegram-token",
			EnvVars: []string{"VINE_TELEGRAM_TOKEN"},
			Usage:   "Telegram token",
		},
		&cli.StringFlag{
			Name:    "telegram-whitelist",
			EnvVars: []string{"VINE_TELEGRAM_WHITELIST"},
			Usage:   "Telegram bot's users (comma-separated values)",
		},
	}
}

func (ti *telegramInput) Init(ctx *cli.Context) error {
	ti.debug = ctx.Bool("telegram-debug")
	ti.token = ctx.String("telegram-token")

	whitelist := ctx.String("telegram-whitelist")

	if whitelist != "" {
		ti.whitelist = strings.Split(whitelist, ",")
	}

	if len(ti.token) == 0 {
		return errors.New("missing telegram token")
	}

	return nil
}

func (ti *telegramInput) Stream() (input.Conn, error) {
	ti.Lock()
	defer ti.Unlock()

	return newConn(ti)
}

func (ti *telegramInput) Start() error {
	ti.Lock()
	defer ti.Unlock()

	api, err := tgbotapi.NewBotAPI(ti.token)
	if err != nil {
		return err
	}

	ti.api = api

	api.Debug = ti.debug

	return nil
}

func (ti *telegramInput) Stop() error {
	return nil
}

func (p *telegramInput) String() string {
	return "telegram"
}
