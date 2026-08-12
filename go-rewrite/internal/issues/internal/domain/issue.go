package domain

import "time"

type Issue struct {
	id 						int64
	contentItemID int64
	scanRule			ScanRule
	contentXPath	string
	status				IssueStatus
	severity 			IssueSeverity
	fixedBy				int64
	fixedAt				time.Time

	// details contains amorphous additional information about the issue
	details				map[string]any

	createdAt			time.Time
	updatedAt			time.Time
}

func NewIssue(contentItemID int64, scanRule ScanRule, contentXPath string, status IssueStatus, severity IssueSeverity, details map[string]any) *Issue {
	return &Issue{
		contentItemID: 	contentItemID,
		scanRule:      	scanRule,
		contentXPath:  	contentXPath,
		status:   	status,
		severity: 	severity,
		details:       	details,
		createdAt:   	 	time.Now(),
		updatedAt:    	time.Now(),
	}
}

func RehydrateIssue(
	id 						int64,
	contentItemID int64,
	scanRule 			ScanRule,
	contentXPath 	string,
	status				IssueStatus,
	severity 			IssueSeverity,
	fixedBy 			int64,
	fixedAt 			time.Time,
	details 			map[string]any,
	createdAt 		time.Time,
	updatedAt 		time.Time,
) *Issue {
	return &Issue{
		id:            id,
		contentItemID: contentItemID,
		scanRule:      scanRule,
		contentXPath:  contentXPath,
		status:        status,
		severity:     severity,
		fixedBy:       fixedBy,
		fixedAt:       fixedAt,
		details:       details,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (i *Issue) ID() int64 {
	return i.id
}

func (i *Issue) ContentItemID() int64 {
	return i.contentItemID
}

func (i *Issue) ScanRule() ScanRule {
	return i.scanRule
}

func (i *Issue) ContentXPath() string {
	return i.contentXPath
}

func (i *Issue) Status() IssueStatus {
	return i.status
}

func (i *Issue) Severity() IssueSeverity {
	return i.severity
}

func (i *Issue) FixedBy() int64 {
	return i.fixedBy
}

func (i *Issue) FixedAt() time.Time {
	return i.fixedAt
}

func (i *Issue) Details() map[string]any {
	return i.details
}

func (i *Issue) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Issue) UpdatedAt() time.Time {
	return i.updatedAt
}