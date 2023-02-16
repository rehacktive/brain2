package brain2

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

type WhisperAPI struct {
	BaseURL string
}

func (w WhisperAPI) SendFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a new multipart writer.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a new form file field and add the file to it.
	fileField, err := writer.CreateFormFile("file", filename)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(fileField, file)
	if err != nil {
		panic(err)
	}

	// Close the multipart writer.
	err = writer.Close()
	if err != nil {
		panic(err)
	}

	log.Println("whisper in progress...")
	req, err := http.NewRequest("POST", w.BaseURL+"/whisper", body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var res Results
	err = json.Unmarshal(content, &res)
	if err != nil {
		return "", err
	}

	return res.Results[0].Transcript, nil
}

type Results struct {
	Results []Result `json:"results"`
}

type Result struct {
	Filename   string `json:"filename"`
	Transcript string `json:"transcript"`
}
