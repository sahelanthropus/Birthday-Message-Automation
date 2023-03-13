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
}

// Struct for secret key/value pair
type SecretData struct {
	AccountSID	string `json:"accountSID"`
	AuthToken	string `json:"authToken"`
}

func main() {
	lambda.Start(sendBirthdayMessage)
}


func sendBirthdayMessage() {
	// Get today's date
	today	:= time.Now()
	date	:= fmt.Sprintf("%02d/%02d/%02d", int(today.Month()), today.Day(), today.Year())

	// Open the CSV file
	file, err := os.Open("birthdays2023.csv")
	if err != nil {
		// handle error
		fmt.Println(err)
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
			fmt.Println(err)
			break
		}

		// Parse the date string in the CSV into a time.Time type
		parsedDate, err := time.Parse("1/2/2006", row[0])
		if err != nil {
			// handle error - unable to parse date string
		}
		dateCSV := fmt.Sprintf("%02d/%02d/%02d", int(parsedDate.Month()), parsedDate.Day(), parsedDate.Year())


		if dateCSV == date {
			birthday = Birthdays {
				Date: row[0],
				Name: row[1],
				PhoneNumber: row[2],
				Message: row[3],
			}
			break
		}
	}

	textMessage := "Happy Birthday " + birthday.Name


	// It is recommended to follow best practices for handling secrets in your code,
	// such as storing as environment variables on in a secure configuration.
	
	// Get encrypted API key pairs from Secrets Manager
	secretName := "test/twilio/birthdayAutomation"
    region := "us-east-1"
    
    config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
    if err != nil {
        log.Fatal(err)
    }

    // Create Secrets Manager client
    svc := secretsmanager.NewFromConfig(config)

    input := &secretsmanager.GetSecretValueInput{
        SecretId:     aws.String(secretName),
        VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
    }

    result, err := svc.GetSecretValue(context.TODO(), input)
    if err != nil {
        // For a list of exceptions thrown, see
        // https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
        log.Fatal(err.Error())
    }

    // Initialize struct for twilio Secret
    twilioSecret := SecretData{}

    // Decrypt and store into SecretData struct
    err2 := json.Unmarshal([]byte(*result.SecretString), &twilioSecret)
    if err2 != nil {
        // handle error
    }

    // Store the AccountSID and AuthToken as strings
    accountSID := twilioSecret.AccountSID
    authToken := twilioSecret.AuthToken

	from := os.Getenv("TWILIO_FROM_PHONE_NUMBER")
	to := "+1" + birthday.PhoneNumber
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(from)
	params.SetBody(textMessage)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		response, _ := json.Marshal(*resp)
		fmt.Println("Response: " + string(response))
	}
}
