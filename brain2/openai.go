package brain2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type OpenAI struct {
	apiKey    string
	maxTokens int
}

type Query struct {
	Model       string `json:"model"`
	Prompt      string `json:"prompt"`
	Temperature int    `json:"temperature"`
	MaxTokens   int    `json:"max_tokens"`
}

type Response struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string      `json:"text"`
		Index        int         `json:"index"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func NewOpenAI(apiKey string) OpenAI {
	return OpenAI{
		apiKey:    apiKey,
		maxTokens: 250,
	}
}

func (o OpenAI) DoRequest(prompt string) (ret string, err error) {
	// Build the request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/completions", nil)
	if err != nil {
		return
	}

	// Set the API key in the request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	// Set the request body to the JSON-encoded query
	query := Query{
		Model:       "text-davinci-003",
		Prompt:      prompt,
		Temperature: 0,
		MaxTokens:   o.maxTokens,
	}
	jsonBody, err := json.Marshal(query)
	if err != nil {
		return
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(jsonBody))

	// Send the request and get the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read and parse the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var response Response
	json.Unmarshal(body, &response)

	fmt.Println(response)

	ret = response.Choices[0].Text
	return
}
