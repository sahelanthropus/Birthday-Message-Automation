package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"context"
	"log"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/aws/aws-lambda-go/lambda"
)

// Struct for Birthdays in CSV file
type Birthdays struct {
	Date		string
	Name		string
	PhoneNumber	string
	Message		string
	Discord		string
}

// Struct for Twilio secret key/value pair
type SecretData struct {
	AccountSID	string `json:"accountSID"`
	AuthToken	string `json:"authToken"`
}

// Struct for Discord secret key/value pair
type DiscordSecretData struct {
    AuthToken string `json:"authToken"`
}

// Main function
func main() {
	lambda.Start(sendBirthdayMessage)
}

// Function for sending Happy Birthday Message
func sendBirthdayMessage() {
	// Get today's date
	today	:= time.Now()
	date	:= fmt.Sprintf("%02d/%02d/%02d", int(today.Month()), today.Day(), today.Year())

	// Open the CSV file
	file, err := os.Open("birthdays2023.csv")
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer file.Close()

	// Define CSV reader
	reader := csv.NewReader(file)

	// Define birthday variable
	var birthday Birthdays

	// Read the CSV rows
	for {
		row, err := reader.Read()
		if err != nil {
			// handle error
			log.Fatal(err)
			break
		}

		// Parse the date string in the CSV into a time.Time type
		parsedDate, err := time.Parse("1/2/2006", row[0])
		if err != nil {
			// handle error - unable to parse date string
		}
		dateCSV := fmt.Sprintf("%02d/%02d/%02d", int(parsedDate.Month()), parsedDate.Day(), parsedDate.Year())

		// check the date in CSV for today's date.
		if dateCSV == date {
			birthday = Birthdays {
				Date: row[0],
				Name: row[1],
				PhoneNumber: row[2],
				Message: row[3],
				Discord: row[4],
			}
			break
		}
	}

	// Check if birthday.Name is empty
	if birthday.Name == "" {
		return
	}

	// Prepare message
	var message string
	if birthday.Message == "" {
		message = "Happy Birthday " + birthday.Name
	} else {
		message = birthday.Message
	}

	// Determine if Discord or Twilio message
	if len(birthday.Discord) > 0 {
		err := sendDiscord(birthday, message)
	} else {
		if err != nil {
			log.Fatal(err)
		}
		err := sendTwilio(birthday, message)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Function to get Secret from AWS Secrets Manager
func getSecretString(secretName, region string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", err
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:	 aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(context.Background(), input)
	if err != nil {
		// For a list of exceptions thrown, see
		// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
		return "", err
	}

	// Return secret
	return *result.SecretString, nil
}

// Function to send text message using Twilio
func sendTwilio(birthday Birthdays, textMessage string) error {
	// Get twilio API key pairs from Secrets Manager
	twilioSecret := SecretData{}
	twilioSecretName := "test/twilio/birthdayAutomation"
	region := "us-east-1"

	twilioSecretString, err := getSecretString(twilioSecretName, region)
	if err != nil {
		return err
	}

	err2 := json.Unmarshal([]byte(twilioSecretString), &twilioSecret)
	if err2 != nil{
		return err
	}

	// Store the AccountSID and AuthToken as strings
	accountSID := twilioSecret.AccountSID
	authToken := twilioSecret.AuthToken

	// Set up Twilio client
	from := os.Getenv("TWILIO_FROM_PHONE_NUMBER")
	to := "+1" + birthday.PhoneNumber
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})

	// Create message params
	params := &twilioApi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(from)
	params.SetBody(textMessage)

	// Send message
	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		return err
	} else {
		response, _ := json.Marshal(*resp)
		fmt.Println("Response: " + string(response))
	}

	return nil
}

// Function to send a specified message to a discord channel
func sendDiscord(birthday Birthdays, message string) error {
	url := os.Getenv("DISCORD_FRIENDZONE_CHANNEL")
	payload := map[string]string{"content": message}
	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))

	// Get authentication token from Secrets Manager
	discordSecret := DiscordSecretData{}
	discordSecretName := "test/discord/birthdayAutomation"
	region := "us-east-1"

	// Pull discord secret
	discordSecretJSON, err := getSecretString(discordSecretName, region)
	if err != nil{
		return err
	}

	// Unmarshal JSON secret
	err = json.Unmarshal([]byte(discordSecretJSON), &discordSecret)
	if err != nil {
		return err
	}
	
	// Set POST headers
	req.Header.Set("Authorization", discordSecret.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("sendDiscord failed with status code: %d", resp.StatusCode)
	}

	return nil
}
