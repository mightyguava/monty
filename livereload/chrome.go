package livereload

import (
	"context"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// Chrome is a live-reload chrome client, backed by the Chrome Developer Protocol
type Chrome struct {
	url  string
	host string
	cdp  *chromedp.CDP
}

// NewChrome allocates and initializes a new Chrome client
func NewChrome(urlToOpen string) (*Chrome, error) {
	if !strings.Contains(urlToOpen, "://") {
		urlToOpen = "http://" + urlToOpen
	}
	u, err := url.Parse(urlToOpen)
	if err != nil {
		return nil, err
	}
	// create chrome instance
	cdp, err := chromedp.New(
		context.Background(),
		chromedp.WithErrorf(log.Printf),
	)
	if err != nil {
		return nil, err
	}
	return &Chrome{
		url:  urlToOpen,
		cdp:  cdp,
		host: u.Host,
	}, nil
}

func getContext() (context.Context, func()) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func (c *Chrome) waitForReachability() error {
	log.Printf("Waiting for %v to be reachable", c.host)
	ctx, cancel := getContext()
	defer cancel()
	dialer := &net.Dialer{}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.Tick(10 * time.Millisecond):
			_, err := dialer.DialContext(ctx, "tcp", c.host)
			if err == nil {
				return nil
			}
		}
	}
}

// Open navigates the Chrome browser to the given URL.
func (c *Chrome) Open() error {
	err := c.waitForReachability()
	if err != nil {
		log.Println("could not connect to server: ", err)
	}
	log.Println("reachable")
	ctx, cancel := getContext()
	defer cancel()
	return c.cdp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(c.url),
	})
}

// Reload triggers the Chrome browser to reload
func (c *Chrome) Reload() error {
	err := c.waitForReachability()
	if err != nil {
		log.Println("could not connect to server: ", err)
	}
	log.Println("reachable")
	ctx, cancel := getContext()
	defer cancel()
	return c.cdp.Run(ctx, chromedp.Tasks{
		chromedp.Reload(),
	})
}

// Close disconnects and closes the Chrome browser
func (c *Chrome) Close() error {
	ctx, cancel := getContext()
	defer cancel()
	return c.cdp.Shutdown(ctx)
}
