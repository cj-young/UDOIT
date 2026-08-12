package domain

// This enum essentially mirrors the IssueSeverity enum in the issues module.
// This is probably bad duplication and needs to be reevaluated.
type ScanIssueSeverity string

const (
	ScanIssueSeverityError       ScanIssueSeverity = "error"
	ScanIssueSeverityPotential   ScanIssueSeverity = "potential"
	ScanIssueSeveritySuggestion  ScanIssueSeverity = "suggestion"
)

func (s ScanIssueSeverity) String() string {
	return string(s)
}