package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type Birthdays struct {
	Date string
	Name string
	PhoneNumber string
	Message string
}

func main() {
	// Get today's date
	today := time.Now()
	date := fmt.Sprintf("%02d/%02d/%02d", int(today.Month()), today.Day(), today.Year())

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

	text := "Happy Birthday " + birthday.Name
	fmt.Println(text)

	// Grab the phone number and name for the date

	// Send text message
	// It is recommended to follow best practices for handling secrets in your code, such as storing as environment variables on in a secure configuration.
	accountSID := "" //"ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	authToken := "" //"YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY"
	from := ""
	to := "+1" + birthday.PhoneNumber
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
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

	// Translate to Lambda