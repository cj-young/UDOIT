package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"rewritetest/internal/lms"
)

type EqualAccessScanner struct {
	httpClient *http.Client
	baseURL string
}

func NewEqualAccessScanner(httpClient *http.Client, baseURL string) *EqualAccessScanner {
	return &EqualAccessScanner{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

type ScanRequest struct {
	HTML string `json:"html"`
	GuidelineIDs []string `json:"guidelineIds"`
	ReportLevels []string `json:"reportLevels"`
}

type ScanResult struct {
	Results []ScanIssue `json:"results"`
}

type ScanIssue struct {
	Message 	string 		`json:"message"`
	Path 			ScanPath 	`json:"path"`
	ReasonID 	string 		`json:"reasonId"`
	RuleID		string		`json:"ruleId"`
	Value 		[]string 	`json:"value"`
}

type ScanPath struct {
	ARIA string `json:"aria"`
	DOM string `json:"dom"`
}

func (e *EqualAccessScanner) ScanContent(ctx context.Context, items []lms.ContentItemDTO) error {

	for _, item := range items {
		body := ScanRequest{
			HTML:         item.HTML,
			GuidelineIDs: nil,
			ReportLevels: nil,
		}

		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}

		resp, err := e.httpClient.Post(e.baseURL+"/scan", "application/json", bytes.NewReader(payload))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var result ScanResult
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return err
		}

		// for logging
		{
			respJSONBytes, err := json.Marshal(result)
			if err != nil {
				return err
			}

			var prettyJSON bytes.Buffer
			err = json.Indent(&prettyJSON, respJSONBytes, "", "  ")
			if err != nil {
				return err
			}

			fmt.Printf(prettyJSON.String() + "\n")
		}

		
	}

	return nil
}