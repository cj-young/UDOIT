package domain

// Mirrors the IssueStatus in the issues module. Like ScanIssueSeverity, it is
// likely bad duplication and needs to be reevaluated.
type ScanIssueStatus string

const (
	ScanIssueStatusActive           ScanIssueStatus = "active"
	ScanIssueStatusFixed            ScanIssueStatus = "fixed"
	ScanIssueStatusMarkedAsResolved ScanIssueStatus = "marked_as_resolved"
)

func (s ScanIssueStatus) String() string {
	return string(s)
}
