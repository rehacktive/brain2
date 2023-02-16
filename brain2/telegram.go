package brain2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	baseURL         = "https://api.telegram.org/bot"
	baseDownloadURL = "https://api.telegram.org/file/bot"

	// path
	getUpdatesParam  = "/getUpdates"
	sendMessageParam = "/sendMessage"
	sendImageParam   = "/sendPhoto"
	getFile          = "/getFile?file_id="

	filesFolder = "./files/"
)

type TelegramManager struct {
	botKey  string
	cmdChan chan Command
}

type Command struct {
	Cmd          string
	From         string
	VoiceCmd     string
	lastUpdateID int
}

func NewTelegramManager(botKey string) TelegramManager {
	return TelegramManager{
		botKey:  botKey,
		cmdChan: make(chan Command, 1),
	}
}

func (t TelegramManager) ListenForCommands() <-chan Command {
	c := make(chan Command)
	go t.getUpdates(0, c)

	go func() {
		for command := range c {
			if command.Cmd != "empty" {
				t.cmdChan <- command
			}
			go func(last int) {
				time.Sleep(5 * time.Second)
				t.getUpdates(last+1, c)
			}(command.lastUpdateID)
		}
	}()

	return t.cmdChan
}

func (t TelegramManager) SendMessage(destinationID string, message string) error {
	url := baseURL + t.botKey + sendMessageParam
	values := map[string]string{"chat_id": destinationID, "text": message}
	json, err := json.Marshal(values)
	if err != nil {
		return err
	}
	_, err = http.Post(url, "application/json", bytes.NewBuffer(json))
	return err
}

type updateResponse struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateID int `json:"update_id"`
		Message  struct {
			MessageID int `json:"message_id"`
			From      struct {
				ID           int    `json:"id"`
				IsBot        bool   `json:"is_bot"`
				FirstName    string `json:"first_name"`
				Username     string `json:"username"`
				LanguageCode string `json:"language_code"`
			} `json:"from"`
			Chat struct {
				ID        int    `json:"id"`
				FirstName string `json:"first_name"`
				Username  string `json:"username"`
				Type      string `json:"type"`
			} `json:"chat"`
			Date  int    `json:"date"`
			Text  string `json:"text"`
			Voice struct {
				Duration     int    `json:"duration"`
				MimeType     string `json:"mime_type"`
				FileID       string `json:"file_id"`
				FileUniqueID string `json:"file_unique_id"`
				FileSize     int    `json:"file_size"`
			} `json:"voice"`
		} `json:"message"`
	} `json:"result"`
}

func (t TelegramManager) getUpdates(lastUpdate int, c chan Command) {
	url := baseURL + t.botKey + getUpdatesParam + "?offset=" + strconv.Itoa(lastUpdate)
	res, err := http.Get(url)
	if err == nil {
		var data updateResponse
		decoder := json.NewDecoder(res.Body)
		decoder.Decode(&data)
		fmt.Println(data)
		if len(data.Result) > 0 {
			lastItem := data.Result[len(data.Result)-1]

			fromID := fmt.Sprint(lastItem.Message.From.ID)
			last := lastItem.UpdateID
			var cmd, voiceCmd string

			if lastItem.Message.Text != "" {
				log.Println("received a text command")
				cmd = lastItem.Message.Text
			}

			if lastItem.Message.Voice.FileID != "" {
				log.Println("received a voice command")
				file, err := t.downloadFile(lastItem.Message.Voice.FileID)
				if err != nil {
					log.Println(err)
					cmd = "error on downloading file"
				}
				voiceCmd = file
			}
			c <- Command{Cmd: cmd, From: fromID, VoiceCmd: voiceCmd, lastUpdateID: last}
			return
		}
	} else {
		log.Println("error on getting telegram updates:", err)
	}
	c <- Command{"empty", "", "", 0}
}

type downloadFile struct {
	Ok     bool `json:"ok"`
	Result struct {
		FileID       string `json:"file_id"`
		FileUniqueID string `json:"file_unique_id"`
		FileSize     int    `json:"file_size"`
		FilePath     string `json:"file_path"`
	} `json:"result"`
}

func (t TelegramManager) downloadFile(FileID string) (fullPathFile string, err error) {
	url := baseURL + t.botKey + getFile + FileID
	fmt.Println(url)

	fullPathFile = filesFolder + FileID

	file, err := os.Create(fullPathFile)
	if err != nil {
		return
	}
	client := http.Client{}

	resp, err := client.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var data downloadFile
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&data)

	url = baseDownloadURL + t.botKey + "/" + data.Result.FilePath
	resp, err = client.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)

	defer file.Close()

	log.Printf("Downloaded a file %s with size %d", fullPathFile, size)
	return
}
