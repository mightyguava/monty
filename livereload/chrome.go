package livereload

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
)

// Chrome is a live-reload chrome client, backed by the Chrome Developer Protocol
type Chrome struct {
	URL string
	cdp *chromedp.CDP
}

// NewChrome allocates and initializes a new Chrome client
func NewChrome(url string) (*Chrome, error) {
	// create chrome instance
	cdp, err := chromedp.New(context.Background(), chromedp.WithErrorf(log.Printf))
	if err != nil {
		return nil, err
	}
	return &Chrome{
		URL: url,
		cdp: cdp,
	}, nil
}

// Open navigates the Chrome browser to the given URL.
func (c *Chrome) Open() error {
	return c.cdp.Run(context.Background(), chromedp.Tasks{
		chromedp.Navigate(c.URL),
	})
}

// Reload triggers the Chrome browser to reload
func (c *Chrome) Reload() error {
	return c.cdp.Run(context.Background(), chromedp.Tasks{
		chromedp.Reload(),
	})
}

// Close disconnects and closes the Chrome browser
func (c *Chrome) Close() error {
	return c.cdp.Shutdown(context.Background())
}
