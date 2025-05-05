package sentimentAnalysis

import (
	"encoding/json"
	"fmt"
	"github.com/gage-technologies/mistral-go"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Analysis"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
)

var client *mistral.MistralClient

func InitSentimentAnalysis(apiKey string) {
	client = mistral.NewMistralClientDefault(apiKey)
}

func SentimentAnalysis(feedback Feedback.Feedback) (Analysis.Analysis, error) {
	// Example: Using Chat Completions
	maxRetries := 5
	retryCount := 0
	retry := true
	var chatRes *mistral.ChatCompletionResponse
	for retry && retryCount < maxRetries {
		retryCount++
		retry = false
		var err error
		chatRes, err = client.Chat("mistral-large-latest", []mistral.ChatMessage{
			{Content: "I will give you in input sentence, you need to analyse this sentence, give it a score between -1 and +1 (negative,neutral and positive), also give it a topic in french among thoese : 'Support Client,Tarifs et Valeur,Interface Utilisateur,Bugs et Problèmes Techniques,Documentation,Performance,Personnalisation,Processus d’Inscription,Fonctionnalités Avancées,Expérience Utilisateur Générale' and output this in the json format like : \n{\n    \"topic\":\"the theme of the sentence\",\n    \"sentiment_score\":0\n}", Role: mistral.RoleSystem},
			{Content: "Sure, please provide the sentence you'd like me to analyze.", Role: mistral.RoleAssistant},
			{Content: feedback.Text, Role: mistral.RoleUser},
		}, // make the output json
			&mistral.ChatRequestParams{
				Temperature:    1,
				TopP:           1,
				RandomSeed:     42069,
				MaxTokens:      4000,
				SafePrompt:     false,
				ResponseFormat: mistral.ResponseFormatJsonObject,
			})
		fmt.Println("error", err)
		fmt.Println("chatRes", chatRes)
		fmt.Println("text", feedback.Text)
		if err != nil {
			retry = true
		}
	}
	if chatRes == nil {
		return Analysis.Analysis{}, fmt.Errorf("no response from Mistral API after %d retries", maxRetries)
	}
	// Parse the response to get the sentiment score and topic
	var analysis Analysis.Analysis
	//remove the ```json from the start and ``` from the end
	jsonOutput := chatRes.Choices[0].Message.Content
	err := json.Unmarshal([]byte(jsonOutput), &analysis)
	if err != nil {
		return Analysis.Analysis{}, err
	}
	analysis.FeedbackID = feedback.Id
	return analysis, nil
}
