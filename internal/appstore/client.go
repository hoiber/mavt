package appstore

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/thomas/mavt/pkg/models"
)

const (
	lookupURL = "https://itunes.apple.com/lookup"
	searchURL = "https://itunes.apple.com/search"
)

// Client handles communication with the App Store API
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new App Store API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// iTunesResponse represents the response from iTunes API
type iTunesResponse struct {
	ResultCount int          `json:"resultCount"`
	Results     []iTunesApp  `json:"results"`
}

// iTunesApp represents an app in the iTunes API response
type iTunesApp struct {
	TrackID              int64     `json:"trackId"`
	BundleID             string    `json:"bundleId"`
	TrackName            string    `json:"trackName"`
	Version              string    `json:"version"`
	CurrentVersionReleaseDate string `json:"currentVersionReleaseDate"`
	ReleaseNotes         string    `json:"releaseNotes"`
	ArtistName           string    `json:"artistName"`
	MinimumOsVersion     string    `json:"minimumOsVersion"`
	FileSizeBytes        string    `json:"fileSizeBytes"`
	Price                float64   `json:"price"`
	Currency             string    `json:"currency"`
}

// LookupByBundleID fetches app information by bundle ID
func (c *Client) LookupByBundleID(bundleID string) (*models.AppInfo, error) {
	params := url.Values{}
	params.Add("bundleId", bundleID)
	params.Add("entity", "software")
	params.Add("country", "us")

	resp, err := c.httpClient.Get(lookupURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch app info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var itunesResp iTunesResponse
	if err := json.NewDecoder(resp.Body).Decode(&itunesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if itunesResp.ResultCount == 0 {
		return nil, fmt.Errorf("app not found: %s", bundleID)
	}

	return c.convertToAppInfo(itunesResp.Results[0])
}

// LookupByTrackID fetches app information by track ID
func (c *Client) LookupByTrackID(trackID int64) (*models.AppInfo, error) {
	params := url.Values{}
	params.Add("id", fmt.Sprintf("%d", trackID))
	params.Add("entity", "software")
	params.Add("country", "us")

	resp, err := c.httpClient.Get(lookupURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch app info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var itunesResp iTunesResponse
	if err := json.NewDecoder(resp.Body).Decode(&itunesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if itunesResp.ResultCount == 0 {
		return nil, fmt.Errorf("app not found: %d", trackID)
	}

	return c.convertToAppInfo(itunesResp.Results[0])
}

// SearchApps searches for apps by name/term
func (c *Client) SearchApps(term string, limit int) ([]*models.AppInfo, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	params := url.Values{}
	params.Add("term", term)
	params.Add("entity", "software")
	params.Add("country", "us")
	params.Add("limit", fmt.Sprintf("%d", limit))

	resp, err := c.httpClient.Get(searchURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to search apps: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var itunesResp iTunesResponse
	if err := json.NewDecoder(resp.Body).Decode(&itunesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var apps []*models.AppInfo
	for _, itunesApp := range itunesResp.Results {
		app, err := c.convertToAppInfo(itunesApp)
		if err != nil {
			continue
		}
		apps = append(apps, app)
	}

	return apps, nil
}

// convertToAppInfo converts iTunes API response to internal model
func (c *Client) convertToAppInfo(app iTunesApp) (*models.AppInfo, error) {
	releaseDate, err := time.Parse(time.RFC3339, app.CurrentVersionReleaseDate)
	if err != nil {
		releaseDate = time.Now()
	}

	var fileSize int64
	fmt.Sscanf(app.FileSizeBytes, "%d", &fileSize)

	return &models.AppInfo{
		BundleID:        app.BundleID,
		TrackID:         app.TrackID,
		TrackName:       app.TrackName,
		Version:         app.Version,
		ReleaseDate:     releaseDate,
		ReleaseNotes:    app.ReleaseNotes,
		ArtistName:      app.ArtistName,
		MinOSVersion:    app.MinimumOsVersion,
		FileSizeBytes:   fileSize,
		Price:           app.Price,
		Currency:        app.Currency,
		LastChecked:     time.Now(),
		FirstDiscovered: time.Now(),
	}, nil
}
