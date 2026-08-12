package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"rewritetest/internal/scanner/internal/domain"
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

// ScanResult represents the structure of the scan results returned by the
// EqualAccess scanning service.
type ScanResult struct {
	Results []ScanIssue `json:"results"`
}

type ScanIssue struct {
	Message 		string 		`json:"message"`
	Path 				ScanPath 	`json:"path"`
	ReasonID 		string 		`json:"reasonId"`
	RuleID			string		`json:"ruleId"`
	Value 			[]string 	`json:"value"`
	MessageArgs []any 	`json:"messageArgs"`
}

type ScanPath struct {
	ARIA string `json:"aria"`
	DOM string `json:"dom"`
}


func (e *EqualAccessScanner) ScanContent(ctx context.Context, items []domain.ScanItem) ([]domain.ScanResult, error) {
	scanResults := make([]domain.ScanResult, 0)

	for _, item := range items {
		body := ScanRequest{
			HTML:         item.HTML,
			GuidelineIDs: nil,
			ReportLevels: nil,
		}

		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		scanResponse, err := e.httpClient.Post(e.baseURL+"/scan", "application/json", bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		defer scanResponse.Body.Close()


		var result ScanResult
		err = json.NewDecoder(scanResponse.Body).Decode(&result)
		if err != nil {
			return nil, err
		}

		// for logging
		{
			respJSONBytes, err := json.Marshal(result)
			if err != nil {
				return nil, err
			}

			var prettyJSON bytes.Buffer
			err = json.Indent(&prettyJSON, respJSONBytes, "", "  ")
			if err != nil {
				return make([]domain.ScanResult, 0), err
			}

			fmt.Printf(prettyJSON.String() + "\n")
		}

		for _, issue := range result.Results {
			domainResult := domain.ScanResult{
				ContentItemID: 	item.ContentItemID,
				ScanRule:      	issue.RuleID,
				ContentXPath:  	issue.Path.DOM,
				Severity:     	e.getScanIssueSeverity(issue),
				Details:       	map[string]any{
					"message":    	issue.Message,
					"messageArgs":	issue.MessageArgs,
				},
			}

			scanResults = append(scanResults, domainResult)
		}

		
	}

	return scanResults, nil
}

func (e *EqualAccessScanner) getScanIssueSeverity(issue ScanIssue) domain.ScanIssueSeverity {

	// TODO: handle "PASS"

	if issue.Value[1] == "MANUAL" {
		return domain.ScanIssueSeverityPotential
	}

	if issue.Value[0] == "VIOLATION" {
		if issue.Value[1] == "FAIL" {
			return domain.ScanIssueSeverityError
		} else {
			return domain.ScanIssueSeverityPotential
		}
	}

	if issue.Value[0] == "RECOMMENDATION" {
		return domain.ScanIssueSeverityPotential
	}


	return domain.ScanIssueSeverityError
}

var _ domain.Scanner = (*EqualAccessScanner)(nil)