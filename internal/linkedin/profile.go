package linkedin

import (
	"context"
	"fmt"
	"strings"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/jitter"
)

// VisitProfile navigates to profileURL, applies human-like dwell time and
// scrolling, then extracts structured profile data from the DOM.
func (c *Client) VisitProfile(ctx context.Context, profileURL string) (explorer.ProfileData, error) {
	c.log.InfoContext(ctx, "visiting profile", "url", profileURL)

	if err := c.navigate(ctx, profileURL); err != nil {
		return explorer.ProfileData{}, fmt.Errorf("navigate to profile: %w", err)
	}

	if err := c.dwellAndScroll(ctx); err != nil {
		return explorer.ProfileData{}, err
	}

	data, err := c.extractProfileData()
	if err != nil {
		return explorer.ProfileData{}, err
	}
	data.URL = profileURL
	data.Slug = slugFromURL(profileURL)
	return data, nil
}

func (c *Client) dwellAndScroll(ctx context.Context) error {
	dwell := jitter.DwellDuration(200) // assume ~200 words on a profile
	if err := jitter.TimeSleeper(ctx, dwell); err != nil {
		return err
	}
	return jitter.HumanScroll(ctx, c.page.Mouse, scrollPixels, jitter.TimeSleeper)
}

func (c *Client) extractProfileData() (explorer.ProfileData, error) {
	result, err := c.page.Eval(extractProfileJS)
	if err != nil {
		return explorer.ProfileData{}, fmt.Errorf("extract profile data: %w", err)
	}

	obj := result.Value
	var posts []string
	for _, v := range obj.Get("posts").Arr() {
		if s := v.Str(); s != "" {
			posts = append(posts, s)
		}
	}

	return explorer.ProfileData{
		Name:        strings.TrimSpace(obj.Get("name").Str()),
		Headline:    strings.TrimSpace(obj.Get("headline").Str()),
		Location:    strings.TrimSpace(obj.Get("location").Str()),
		About:       strings.TrimSpace(obj.Get("about").Str()),
		RecentPosts: posts,
	}, nil
}
