package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SynologyClient holds the connection details for a Synology Download Station instance.
type SynologyClient struct {
	BaseURL  string
	Username string
	Password string
	SID      string
	HTTP     *http.Client
}

// SynologyResponse is the generic API response envelope from Synology DSM.
type SynologyResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *SynologyError  `json:"error,omitempty"`
}

// SynologyError represents an error returned by the Synology API.
type SynologyError struct {
	Code int `json:"code"`
}

// NewSynologyClient creates a new SynologyClient using environment variables.
func NewSynologyClient() *SynologyClient {
	return &SynologyClient{
		BaseURL:  getEnv("SYNOLOGY_URL", "http://localhost:5000"),
		Username: getEnv("SYNOLOGY_USER", ""),
		Password: getEnv("SYNOLOGY_PASS", ""),
		HTTP: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// getEnv retrieves an environment variable or returns a fallback default.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Login authenticates with the Synology DSM API and stores the session ID.
func (c *SynologyClient) Login() error {
	params := url.Values{}
	params.Set("api", "SYNO.API.Auth")
	params.Set("version", "3")
	params.Set("method", "login")
	params.Set("account", c.Username)
	params.Set("passwd", c.Password)
	params.Set("session", "DownloadStation")
	params.Set("format", "sid")

	respBody, err := c.get("/webapi/auth.cgi", params)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			SID string `json:"sid"`
		} `json:"data"`
		Error *SynologyError `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}
	if !result.Success {
		return fmt.Errorf("login failed with error code: %d", result.Error.Code)
	}

	c.SID = result.Data.SID
	return nil
}

// get performs a GET request against the Synology API endpoint.
func (c *SynologyClient) get(path string, params url.Values) ([]byte, error) {
	rawURL := strings.TrimRight(c.BaseURL, "/") + path + "?" + params.Encode()
	resp, err := c.HTTP.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// LoginHandler is the Gin handler that triggers a Synology DSM login and
// returns the resulting session ID to the caller.
func LoginHandler(c *gin.Context) {
	client := NewSynologyClient()
	if err := client.Login(); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"sid":     client.SID,
	})
}
