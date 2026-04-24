package linkedin

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pavlomaksymov/in-network-explorer/internal/jitter"
)

// SearchProfiles queries LinkedIn for people matching query in location and
// returns up to limit profile URLs.
func (c *Client) SearchProfiles(ctx context.Context, query, location string, limit int) ([]string, error) {
	searchURL := buildSearchURL(query, location)
	c.log.InfoContext(ctx, "searching profiles", "url", searchURL)

	if err := c.navigate(ctx, searchURL); err != nil {
		return nil, fmt.Errorf("navigate to search: %w", err)
	}

	if err := jitter.TimeSleeper(ctx, jitter.Jitter(navigationDelay)); err != nil {
		return nil, err
	}

	result, err := c.page.Eval(extractSearchResultsJS)
	if err != nil {
		return nil, fmt.Errorf("extract search results: %w", err)
	}

	var urls []string
	for _, v := range result.Value.Arr() {
		if u := v.Str(); u != "" {
			urls = append(urls, u)
		}
		if len(urls) >= limit {
			break
		}
	}
	return urls, nil
}

// buildSearchURL constructs a LinkedIn people search URL with query and
// location parameters. Uses Berlin's geo URN as the default location filter.
func buildSearchURL(query, location string) string {
	params := url.Values{
		"keywords": {query},
		"origin":   {"SWITCH_SEARCH_VERTICAL"},
	}
	if location != "" {
		params.Set("geoUrn", fmt.Sprintf(`["%s"]`, berlinGeoUrn))
	}
	return baseURL + searchPath + "?" + params.Encode()
}
