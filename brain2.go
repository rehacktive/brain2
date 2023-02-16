package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/namsral/flag"
	"github.com/rehacktive/brain2/brain2"
)

const (
	START_MESSAGE = `
	____ ____ ____ ____ ____ ____ 
	||B |||r |||a |||i |||n |||2 ||
	||__|||__|||__|||__|||__|||__||
	|/__\|/__\|/__\|/__\|/__\|/__\|
	`
)

type Brain2 struct {
	telegram brain2.TelegramManager
	openai   brain2.OpenAI
	whisper  brain2.WhisperAPI
}

func (b *Brain2) processCommand(cmd brain2.Command) {
	log.Println("received command: ", cmd)
	b.telegram.SendMessage(cmd.From, "thinking...")

	var response string
	var err error
	if cmd.Cmd != "" {
		response, err = b.openai.DoRequest(cmd.Cmd)
	}
	if cmd.VoiceCmd != "" {
		transcript, err2 := b.whisper.SendFile(cmd.VoiceCmd)
		if err2 != nil {
			log.Println(err)
		}
		response, err = b.openai.DoRequest(transcript)
	}
	if err != nil {
		b.telegram.SendMessage(cmd.From, "error : "+err.Error())
	} else {
		b.telegram.SendMessage(cmd.From, response)
	}
}

func main() {
	var botAPIKey, whisperURL, openAIKey string
	flag.String(flag.DefaultConfigFlagname, "brain2.conf", "path to config file")
	flag.StringVar(&botAPIKey, "TELEGRAM_BOT_KEY", "", "telegram bot key")
	flag.StringVar(&whisperURL, "WHISPER_URL", "", "whisper service url")
	flag.StringVar(&openAIKey, "OPENAI_KEY", "", "OpenAI bot key")

	flag.Parse()

	if len(botAPIKey) == 0 {
		log.Fatal("telegram bot key required")
	}
	if len(whisperURL) == 0 {
		log.Fatal("whisper url required")
	}
	if len(openAIKey) == 0 {
		log.Fatal("OpenAI key required")
	}

	telegramManager := brain2.NewTelegramManager(botAPIKey)
	openAI := brain2.NewOpenAI(openAIKey)
	whisper := brain2.WhisperAPI{
		BaseURL: whisperURL,
	}

	brain2 := Brain2{
		telegram: telegramManager,
		openai:   openAI,
		whisper:  whisper,
	}

	brain2.Init()

	log.Println(START_MESSAGE)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	log.Println("best regards.")
}

func (b *Brain2) Init() {
	go func() {
		for cmd := range b.telegram.ListenForCommands() {
			b.processCommand(cmd)
		}
	}()
}
