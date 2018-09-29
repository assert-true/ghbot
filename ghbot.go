package main

import (
	"fmt"
	"github.com/assert-true/ghbot/gh"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/go-playground/webhooks.v5/github"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var githubRepo string
var githubUser string
var githubToken string

var webhookURL string
var webhookName string
var botSecret string
var telegramToken string

func requireEnv(name string) string {
	result := os.Getenv(name)
	if result == "" {
		log.Fatalf("Environment variable '%s' must be set!\n", name)
	}
	return result
}

func main() {
	githubRepo = requireEnv("REPO")
	githubUser = requireEnv("OWNER")
	githubToken = requireEnv("GITHUB_TOKEN")
	webhookURL = requireEnv("WEBHOOK_URL")
	webhookName = requireEnv("WEBHOOK_NAME")
	botSecret = requireEnv("BOT_SECRET")
	telegramToken = requireEnv("TELEGRAM_TOKEN")

	config := &gh.GitHubClientConfig{
		Repo:githubRepo,
		User:githubUser,
		Token:githubToken,
	}

	send := make(chan string, 10)

	client := gh.NewGitHubClient(config)
	client.SetupHook(webhookName, webhookURL)

	hook, _ := github.New(github.Options.Secret(""))

	parsedWebhookURL, err := url.Parse(webhookURL)
	if err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc(parsedWebhookURL.Path, func (w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.PushEvent)
		log.Println("Here")
		if err != nil {
			if err == github.ErrEventNotFound {
				log.Println(err)
			}
		}

		switch payload.(type) {
		case github.PushPayload:
			push := payload.(github.PushPayload)
			send <- fmt.Sprintf("%s pushed new commits to %s", push.Pusher.Name, push.Repository.Name)
		}
	})

	runBot(send)
	http.ListenAndServe(":3000", nil)
}

var registeredChat int64 = -1

func runBot(send chan string) {
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}

			if strings.HasPrefix(update.Message.Text, "/start ") {
				if update.Message.Text[len("/start "):] == botSecret {
					registeredChat = update.Message.Chat.ID
					message := tgbotapi.NewMessage(registeredChat, "Sälü zäme. Ig bi eh Bot 🤖")
					bot.Send(message)
				}
				continue
			}

		}
	}()

	go func() {
		for message := range send {
			if registeredChat != -1 {
				m := tgbotapi.NewMessage(registeredChat, message)
				bot.Send(m)
			}
		}
	}()


}
