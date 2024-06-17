package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v2"
)

type QuizResponse struct {
	Question      string            `json:"Question"`
	Options       []string          `json:"Options"`
	CorrectAnswer map[string]string `json:"CorrectAnswer"`
}

func main() {
	loadEnv()

	client := getOpenAIClient()

	app := &cli.App{
		Name:  "QuizApp",
		Usage: "A CLI app to get dynamic multiple-choice questions generated using OpenAI API",
		Action: func(c *cli.Context) error {
			reader := bufio.NewReader(os.Stdin)
			var topic string

			for {
				fmt.Print("Enter a topic (must not be empty and contain only letters, spaces, and hyphens): ")
				topic, _ = reader.ReadString('\n')
				topic = strings.TrimSpace(topic)
				if isValidTopic(topic) {
					break
				}
				fmt.Println("Invalid topic. Please enter again.")
			}

			for {
				question, options, correctAnswer := fetchQuestion(client, topic)
				fmt.Println("\nQuestion:")
				fmt.Println(question)
				for n, option := range options {
					fmt.Printf("%d. %s \n", n+1, option)
				}
				fmt.Print("\nEnter your answer number (1,2,3,4) or 'q' to quit: ")
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(answer)

				if answer == "q" {
					fmt.Println("Exiting the app. Goodbye!")
					break
				}

				if answer == correctAnswer["Number"] {
					fmt.Println("Correct! Well done!")
				} else {
					fmt.Printf("Incorrect. The correct answer is [ %s. %s ]\n", correctAnswer["Number"], correctAnswer["Answer"])
				}

				fmt.Print("\nPress Enter to continue...")
				reader.ReadString('\n')
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func getOpenAIClient() *openai.Client {
	OPENAI_API_KEY := os.Getenv("OPENAI_API_KEY")
	if OPENAI_API_KEY == "" {
		log.Fatal("OPENAI_API_KEY is not set in .env file")
	}
	return openai.NewClient(OPENAI_API_KEY)
}

func fetchQuestion(client *openai.Client, topic string) (string, []string, map[string]string) {
	message := fmt.Sprintf(`Generate a distinct multiple-choice question about %s with 4 options.
	Format the response as a JSON object with the following structure:
	{
	  "Question": "string",
	  "Options": ["option1", "option2", "option3", "option4"],
	  "CorrectAnswer": {"Number": "string", "Answer": "string"}
	}
	Ensure that the options are plain strings without any prefixes like "a.", "b.", "c.", "d." or "1.", "2.", "3.", "4.".`, topic)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: message,
			},
		},
		MaxTokens: 150,
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		log.Fatalf("Completion error: %v", err)
	}

	//fmt.Println(resp.Choices[0].Message.Content)

	var quizResponse QuizResponse
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &quizResponse)
	if err != nil {
		log.Fatalf("Failed to parse JSON response: %v", err)
	}

	return quizResponse.Question, quizResponse.Options, quizResponse.CorrectAnswer
}

func isValidTopic(topic string) bool {
	if topic == "" {
		return false
	}
	match, _ := regexp.MatchString("^[a-zA-Z\\s-]+$", topic)
	return match
}
