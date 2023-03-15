package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-lambda-go/lambda"
)

// Struct for Birthdays in CSV file
type Birthdays struct {
	Date        string
	Name        string
	PhoneNumber string
	Message     string
	Discord     string
}

// Struct for Twilio secret key/value pair
type SecretData struct {
	AccountSID string `json:"accountSID"`
	AuthToken  string `json:"authToken"`
}

// Struct for Discord secret key/value pair
type DiscordSecretData struct {
	Authorization string `json:"authorization"`
}

// Main function
func main() {
	// The entry point for AWS Lambda.
	lambda.Start(sendBirthdayMessage)
}

/*
sendBirthdayMessage sends a happy birthday message to someone via either Twilio or Discord.
Using context due to dealing with external services to avoid long wait times and potential timeouts.
*/
func sendBirthdayMessage(ctx context.Context) error {
	// Get today's date
	today := time.Now()
	date := fmt.Sprintf("%02d/%02d/%02d", int(today.Month()), today.Day(), today.Year())

	// Open the CSV file
	birthdayFile, err := os.Open("birthdays2023.csv")
	if err != nil {
		return fmt.Errorf("failed to open birthday file: %v", err)
	}
	defer birthdayFile.Close()

	// Read the CSV rows and find a matching birthday
	birthday, err := findMatchingBirthday(ctx, date, birthdayFile)
	if err != nil {
		return fmt.Errorf("failed to find matching birthday: %v", err)
	}

	// Check if birthday.Name is empty
	if birthday.Name == "" {
		return nil
	}

	// Prepare message
	message := prepareMessage(birthday)

	// Determine if Discord or Twilio message
	if len(birthday.Discord) > 0 {
		err := sendDiscord(ctx, birthday, message)
		if err != nil {
			return fmt.Errorf("failed to send Discord message: %v", err)
		}
	} else {
		err := sendTwilio(ctx, birthday, message)
		if err != nil {
			return fmt.Errorf("failed to send Twilio message: %v", err)
		}
	}

	return nil
}

/*
findMatchingBirthday reads the CSV rows and finds a matching birthday based on the current date.
Returns a Birthdays struct containing the matched birthday or an empty struct if no match found.
*/
func findMatchingBirthday(ctx context.Context, date string, birthdayFile *os.File) (Birthdays, error) {
	// Define CSV reader
	reader := csv.NewReader(birthdayFile)

	// Define birthday variable
	var birthday Birthdays

	// Read the CSV rows
	var rowCounter int
	for {
		row, err := reader.Read()
		if err != nil {
			if err.Error() != "EOF" {
				// handle error
				return birthday, fmt.Errorf("failed to read CSV row: %v", err)
			}
			break
		}

		// Skip header row
		if rowCounter == 0 {
			rowCounter++
			continue
		}

		// Parse the date string in the CSV into a time.Time type
		parsedDate, err := time.Parse("1/2/2006", row[0])
		if err != nil {
			// handle error - unable to parse date string
			log.Printf("error parsing date string '%s': %v", row[0], err)
			continue
		}

		dateCSV := fmt.Sprintf("%02d/%02d/%02d", int(parsedDate.Month()), parsedDate.Day(), parsedDate.Year())
	
		// check the date in CSV for today's date.
		if dateCSV == date {
			birthday = Birthdays{
				Date:        row[0],
				Name:        row[1],
				PhoneNumber: row[2],
				Message:     row[3],
				Discord:     row[4],
			}
			break
		}
		rowCounter++
	}
	
	return birthday, nil
}

/*
prepareMessage prepares the birthday message using the name and message provided in the Birthdays struct.
If message is empty, it will use a default message.
*/
func prepareMessage(birthday Birthdays) string {
	var message string
	if birthday.Message == "" {
		message = "Happy Birthday " + birthday.Name
	} else {
		message = birthday.Message
	}
	return message
}

/*
getSecretString gets a secret string value from AWS Secrets Manager given the secret name and region.
*/
func getSecretString(ctx context.Context, secretName, region string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return "", err
	}
	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(cfg)
	
	input := &secretsmanager.GetSecretValueInput{
		SecretId:      aws.String(secretName),
		VersionStage:  aws.String("AWSCURRENT"),
	}
	
	result, err := svc.GetSecretValue(ctx, input)
	if err != nil {
		// For a list of exceptions thrown, see
		// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
		return "", err
	}
	
	// Return secret
	return *result.SecretString, nil
}

/*
sendTwilio sends a text message using Twilio given a Birthdays struct and a message.
*/
func sendTwilio(ctx context.Context, birthday Birthdays, textMessage string) error {
	// Get twilio API key pairs from Secrets Manager
	twilioSecret := SecretData{}
	twilioSecretName := os.Getenv("TWILIO_SECRET_NAME")
	region := "us-east-1"

	twilioSecretString, err := getSecretString(ctx, twilioSecretName, region)
	if err != nil {
		return err
	}
	
	err = json.Unmarshal([]byte(twilioSecretString), &twilioSecret)
	if err != nil {
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
	}
	response, _ := json.Marshal(*resp)
	fmt.Println("Response: " + string(response))
	
	return nil
}

/*
sendDiscord sends a specified message to a discord channel.
Returns an error if there was a problem sending the message.
*/
func sendDiscord(ctx context.Context, birthday Birthdays, message string) error {
	// Discord webhook URL
	url := os.Getenv("DISCORD_FRIENDZONE_CHANNEL")

	// Prepare POST request payload
	payload := map[string]string{"content": message}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord payload: %v", err)
	}
	
	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create Discord request: %v", err)
	}
	
	// Get authentication token from Secrets Manager
	discordSecret := DiscordSecretData{}
	discordSecretName := os.Getenv("DISCORD_SECRET_NAME")
	region := "us-east-1"
	
	// Pull discord secret
	discordSecretJSON, err := getSecretString(ctx, discordSecretName, region)
	if err != nil {
		return fmt.Errorf("failed to get Discord secret: %v", err)
	}
	
	// Unmarshal secret
	err = json.Unmarshal([]byte(discordSecretJSON), &discordSecret)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Discord secret: %v", err)
	}
	
	// Set POST headers
	req.Header.Set("Authorization", "Bot " + discordSecret.Authorization)
	req.Header.Set("Content-Type", "application/json")
	
	// Send request and check response status
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Discord request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Discord request failed with status code: %d", resp.StatusCode)
	}
	
	return nil
}
