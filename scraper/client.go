package scraper

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	baseURL    = "https://www.linkedin.com/voyager/api"
	userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

type Client struct {
	http       *http.Client
	liAt       string
	jsessionid string
	csrfToken  string
}

func NewClient() (*Client, error) {
	liAt := os.Getenv("LI_AT")
	jsessionid := os.Getenv("JSESSIONID")

	if liAt == "" || jsessionid == "" {
		return nil, fmt.Errorf("LI_AT and JSESSIONID environment variables are required")
	}

	// JSESSIONID is used as the CSRF token (LinkedIn's convention)
	// Strip surrounding quotes if present
	csrfToken := jsessionid
	if len(csrfToken) > 2 && csrfToken[0] == '"' {
		csrfToken = csrfToken[1 : len(csrfToken)-1]
	}

	return &Client{
		http: &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		liAt:       liAt,
		jsessionid: jsessionid,
		csrfToken:  csrfToken,
	}, nil
}

func (c *Client) newRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	cookieHeader := fmt.Sprintf(`li_at=%s; JSESSIONID="%s"`, c.liAt, c.csrfToken)
	req.Header.Set("Cookie", cookieHeader)
	req.Header.Set("Csrf-Token", c.csrfToken)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.linkedin.normalized+json+2.1")
	req.Header.Set("X-RestLi-Protocol-Version", "2.0.0")
	req.Header.Set("X-Li-Lang", "en_US")
	req.Header.Set("X-Li-Track", `{"clientVersion":"1.13.5","mpVersion":"1.13.5","osName":"web","timezoneOffset":-5,"timezone":"America/New_York","deviceFormFactor":"DESKTOP","mpName":"voyager-web","displayDensity":1,"displayWidth":1920,"displayHeight":1080}`)

	return req, nil
}
