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

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	whapiBaseURL         = "https://gate.whapi.cloud/messages/text"
	spreadsheetReadRange = "NOTIFICAÇÕES (NÃO MEXER)!A2:B"
)

var (
	whapiApiKey   = os.Getenv("WHAPI_API_KEY")
	googleApiKey  = os.Getenv("GOOGLE_API_KEY")
	spreadsheetID = os.Getenv("GOOGLE_SPREADSHEET_ID")
)

type pregnantWoman struct {
	name              string
	address           string
	healthAgentName   string
	healthAgentNumber string
}

func sendMessage(to, message string) error {
	body := struct {
		To   string `json:"to"` 
		Body string `json:"body"` 
	}{
		To: "55" + to,
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
			log.Fatalf("Unable to send message: %d", res.StatusCode)
			return errors.New("Unable to send message")
		}

		resBody := make([]byte, res.ContentLength)

		_, err = res.Body.Read(resBody)
		if err != nil {
			return err
		}

		defer res.Body.Close()

		log.Fatalf("Unable to send message: %s", string(resBody))

		return errors.New("Unable to send message")
	}

	return nil
}

func main() {
	ctx := context.Background()
	pregnantWomen := make([]pregnantWoman, 0)

	srv, err := sheets.NewService(ctx, option.WithAPIKey(googleApiKey))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, spreadsheetReadRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		for _, row := range resp.Values {
			pw := pregnantWoman{
				name:              row[0].(string),
				healthAgentNumber: row[1].(string),
			}
			pregnantWomen = append(pregnantWomen, pw)
		}
	}
}
