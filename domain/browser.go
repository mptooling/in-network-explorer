package domain

import "context"

// BlockType classifies the kind of LinkedIn access restriction detected.
type BlockType int

const (
	// BlockNone means no restriction was detected.
	BlockNone BlockType = iota
	// BlockAuthwall means LinkedIn redirected to a login/auth page.
	BlockAuthwall
	// BlockChallenge means a CAPTCHA or identity-verification challenge appeared.
	BlockChallenge
	// BlockSoftEmpty means the profile or feed page rendered as empty (soft block).
	BlockSoftEmpty
)

// ProfileData is the raw data extracted from a LinkedIn profile page.
type ProfileData struct {
	URL         string
	Name        string
	Headline    string
	Location    string
	About       string
	RecentPosts []string
	Slug        string
}

// BrowserClient abstracts the headless browser (go-rod) interactions with
// LinkedIn. Implementations live in adapter/linkedin.
type BrowserClient interface {
	// VisitProfile navigates to profileURL and extracts structured profile data.
	VisitProfile(ctx context.Context, profileURL string) (ProfileData, error)

	// LikeRecentPost finds and likes the most recent post on profileURL.
	LikeRecentPost(ctx context.Context, profileURL string) error

	// SearchProfiles queries LinkedIn for profiles matching query in location and
	// returns up to limit profile URLs.
	SearchProfiles(ctx context.Context, query, location string, limit int) ([]string, error)

	// CheckBlock detects whether the current browser session is being restricted
	// by LinkedIn and classifies the restriction type.
	CheckBlock(ctx context.Context) (BlockType, error)

	// Close releases all browser resources.
	Close() error
}
