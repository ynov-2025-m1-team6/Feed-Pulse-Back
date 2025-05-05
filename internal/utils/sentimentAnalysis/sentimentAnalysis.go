package sentimentAnalysis

import (
	"encoding/json"
	"fmt"
	"github.com/gage-technologies/mistral-go"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Analysis"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	gomail "gopkg.in/mail.v2"
)

var client *mistral.MistralClient
var dialer *gomail.Dialer

func InitSentimentAnalysis(apiKey string, emailPass string) {
	client = mistral.NewMistralClientDefault(apiKey)
	dialer = gomail.NewDialer("mail.lucamorgado.com", 465, "noreply-feedpulse@lucamorgado.com", emailPass)
}

func SentimentAnalysis(feedback Feedback.Feedback, userEmail string) (Analysis.Analysis, error) {
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
	// if the score is below -0,5 send an email
	if analysis.SentimentScore <= -0.5 {
		// Send an email to the user
		m := gomail.NewMessage()
		m.SetHeader("From", "noreply-feedpulse@lucamorgado.com")
		m.SetHeader("To", userEmail)
		m.SetHeader("Subject", "Negative Feedback Alert")
		m.SetBody("text/plain", fmt.Sprintf("Dear User,\n\nWe noticed that one of your submitted feedback has a negative sentiment score of %.2f.\n\nFeedback: %s\n\nPlease take a moment to review it.\n\nBest regards,\nFeedPulse Team", analysis.SentimentScore, feedback.Text))

		// Send the email
		if err := dialer.DialAndSend(m); err != nil {
			fmt.Printf("Failed to send email: %v\n", err)
		}
	}
	return analysis, nil
}
