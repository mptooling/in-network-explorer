package explorer

import "time"

// ProspectReport holds the data for a prospect report output.
type ProspectReport struct {
	GeneratedAt time.Time    `json:"generated_at"`
	Prospects   []ReportItem `json:"prospects"`
}

// ReportItem represents a single prospect entry in the report.
type ReportItem struct {
	ProfileURL      string    `json:"profile_url"`
	Name            string    `json:"name"`
	Headline        string    `json:"headline"`
	Location        string    `json:"location"`
	WorthinessScore int       `json:"worthiness_score"`
	CalibratedProb  float64   `json:"calibrated_prob"`
	DraftedMessage  string    `json:"drafted_message"`
	RecentPost      string    `json:"recent_post"`
	ScannedAt       time.Time `json:"scanned_at"`
}
