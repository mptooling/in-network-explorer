package linkedin

import (
	"context"
	"fmt"

	"github.com/pavlomaksymov/in-network-explorer/internal/jitter"
)

// LikeRecentPost navigates to the prospect's recent activity page, finds the
// most recent post, and clicks the like button with human-like timing.
func (c *Client) LikeRecentPost(ctx context.Context, profileURL string) error {
	activityURL := profileURL + activitySuffix
	c.log.InfoContext(ctx, "liking recent post", "url", activityURL)

	if err := c.navigate(ctx, activityURL); err != nil {
		return fmt.Errorf("navigate to activity: %w", err)
	}

	if err := jitter.TimeSleeper(ctx, jitter.Jitter(navigationDelay)); err != nil {
		return err
	}

	// Scroll to make the like button visible.
	if err := jitter.HumanScroll(ctx, c.page.Mouse, 400, jitter.TimeSleeper); err != nil {
		return err
	}

	hasLike, err := c.page.Eval(findLikeButtonJS)
	if err != nil {
		return fmt.Errorf("find like button: %w", err)
	}
	if !hasLike.Value.Bool() {
		c.log.WarnContext(ctx, "no like button found", "url", activityURL)
		return nil
	}

	if err := jitter.TimeSleeper(ctx, jitter.Jitter(500_000_000)); err != nil {
		return err
	}

	clicked, err := c.page.Eval(clickLikeButtonJS)
	if err != nil {
		return fmt.Errorf("click like button: %w", err)
	}
	if !clicked.Value.Bool() {
		c.log.WarnContext(ctx, "like click did not register", "url", activityURL)
	}
	return nil
}
