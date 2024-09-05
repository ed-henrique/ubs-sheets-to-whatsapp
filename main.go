package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
        spreadsheetReadRange = "NOTIFICAÇÕES (NÃO MEXER)!A2:B"
)

var (
	googleApiKey = os.Getenv("GOOGLE_API_KEY")
        spreadsheetID = os.Getenv("GOOGLE_SPREADSHEET_ID")
)

type pregnantWoman struct {
	name              string
	address           string
	healthAgentName   string
	healthAgentNumber string
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
                                name: row[0].(string),
                                healthAgentNumber: row[1].(string),
                        }
                        pregnantWomen = append(pregnantWomen, pw)
		}
	}
}
