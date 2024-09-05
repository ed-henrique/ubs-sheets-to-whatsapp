package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	whapiBaseURL         = "https://gate.whapi.cloud/messages/text"
	spreadsheetReadRange = "NOTIFICAÇÕES (NÃO MEXER)!A2:C"
)

var (
	whapiApiKey   = os.Getenv("WHAPI_API_KEY")
	googleApiKey  = os.Getenv("GOOGLE_API_KEY")
	spreadsheetID = os.Getenv("GOOGLE_SPREADSHEET_ID")

	errInvalidNumber       = errors.New("The number must be either 10 or 11 characters long")
	errCouldNotSendMessage = errors.New("Unable to send message")

	currentPregnantWomen = make([]pregnantWoman, 0)
)

type pregnantWoman struct {
	name              string
	address           string
	healthAgentName   string
	healthAgentNumber string
}

func validateAndFormatNumber(num string) ([]string, error) {
	var formattedNum string

	for _, c := range num {
		if c >= 48 && c <= 57 {
			formattedNum += string(c)
		}
	}

	nums := make([]string, 2)

	nums[0] = "55" + formattedNum
	if len(formattedNum) == 10 {
		nums[1] = "55959" + formattedNum[2:]
	} else if len(formattedNum) == 11 {
		nums[1] = "5595" + formattedNum[3:]
	} else {
		log.Printf("The number %s is invalid\n", num)
		return nil, errInvalidNumber
	}

	return nums, nil
}

func getPregnantWomen(s *sheets.Service) []pregnantWoman {
	pregnantWomen := make([]pregnantWoman, 0)

	res, err := s.Spreadsheets.Values.Get(spreadsheetID, spreadsheetReadRange).Do()
	if err != nil {
		log.Printf("Unable to retrieve data from sheet: %v\n", err)
	}

	if len(res.Values) == 0 {
		log.Print("No data found.\n")
	} else {
		for _, row := range res.Values {
			pw := pregnantWoman{
				name:              row[0].(string),
				healthAgentNumber: row[1].(string),
				address:           row[2].(string),
			}

			pregnantWomen = append(pregnantWomen, pw)
		}
	}

	return pregnantWomen
}

func sendMessage(to []string, message string) error {
	for _, num := range to {
		body := struct {
			To   string `json:"to"`
			Body string `json:"body"`
		}{
			To:   num,
			Body: message,
		}

		bodyRaw, err := json.Marshal(body)
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodPost, whapiBaseURL, bytes.NewBuffer(bodyRaw))
		if err != nil {
			return err
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+whapiApiKey)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
			if res.ContentLength < 0 {
				log.Printf("Unable to send message: %d", res.StatusCode)
				return errCouldNotSendMessage
			}

			resBody := make([]byte, res.ContentLength)

			_, err = res.Body.Read(resBody)
			if err != nil {
				return err
			}

			defer res.Body.Close()

			log.Printf("Unable to send message: %s", string(resBody))
			return errCouldNotSendMessage
		}
	}

	return nil
}

func alertMessage(name, address string) string {
	return fmt.Sprintf(
		"Gestante: *%s*\nEndereço: *%s*\n\nNão compareceu à consulta, verifique, por favor.",
		name,
		address,
	)
}

func main() {
	ctx := context.Background()
	s, err := sheets.NewService(ctx, option.WithAPIKey(googleApiKey))
	if err != nil {
		log.Fatalf("Unable to start Google Sheets service: %v\n", err)
	}

	for {
		pw := getPregnantWomen(s)

		for _, p := range pw {
			if !slices.Contains(currentPregnantWomen, p) {
				nums, err := validateAndFormatNumber(p.healthAgentNumber)
				if err != nil {
					break
				}
				sendMessage(nums, alertMessage(p.name, p.address))
			}
		}

		currentPregnantWomen = pw

		time.Sleep(2 * time.Minute)
	}
}
