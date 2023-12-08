package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var defaultTelegram atomic.Value

func init() {
	defaultTelegram.Store(DefaultTelegram())
}

func Default() *Telegram { return defaultTelegram.Load().(*Telegram) }

func SetDefault(l *Telegram) {
	defaultTelegram.Store(l)
}

func DefaultTelegram() *Telegram {
	t := &Telegram{
		enabled:  viper.GetBool(TelegramEnabled),
		database: viper.GetString(PostgresDatabase),
		apiUrl:   viper.GetString(TelegramApiUrl),
		token:    viper.GetString(TelegramToken),
		chatId:   viper.GetString(TelegramChatId),
	}

	return t
}

type Telegram struct {
	enabled  bool
	database string
	apiUrl   string
	token    string
	chatId   string
	status   status
	err      error
}

func OK() *Telegram {
	return Default().OK()
}

func (t *Telegram) OK() *Telegram {
	return t.SetStatus(StatusOK)
}

func Err(err error) *Telegram {
	return Default().Err(err)
}

func (t *Telegram) Err(err error) *Telegram {
	t.err = err
	return t.SetStatus(StatusErr)
}

func SetStatus(status status) *Telegram {
	return Default().SetStatus(status)
}

func (t *Telegram) SetStatus(status status) *Telegram {
	t.status = status
	return t
}

func Msg(msg string, a ...any) error {
	return Default().Msg(msg, a...)
}

func (t *Telegram) Msg(msg string, a ...any) error {
	return t.action(SendMessage, fmt.Sprintf(msg, a...))
}

func (t *Telegram) action(method, msg string) error {
	if !t.enabled || t.token == "" || t.chatId == "" {
		return nil
	}

	if t.err != nil {
		msg = fmt.Sprintf("\n- Error: %v\n%s", t.err, msg)
		defer func(t *Telegram) {
			t.err = nil
		}(t)
	}

	if t.database != "" {
		msg = fmt.Sprintf("[%s] - %s", strings.ToUpper(t.database), msg)
	}

	switch t.status {
	case StatusOK:
		msg = fmt.Sprintf("🟢 %s", msg)
	case StatusErr:
		msg = fmt.Sprintf("🔴 %s", msg)
	}

	data := map[string]interface{}{
		"chat_id": t.chatId,
		"text":    msg,
	}

	// Chuyển requestBody thành JSON
	body, err := sonic.Marshal(data)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Tạo request
	url := fmt.Sprintf("%s/bot%s/%s", t.apiUrl, t.token, method)

	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Err(err).Msg("Error create http.NewRequest")
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	// Gửi request
	client := http.DefaultClient
	response, err := client.Do(request)
	if err != nil {
		log.Err(err).Msg("Error sending message")
		return err
	}
	defer response.Body.Close()

	return nil
}
