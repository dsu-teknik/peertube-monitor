package peertube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	username   string
	password   string
	token      string
	httpClient *http.Client
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type clientCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type uploadResponse struct {
	Video struct {
		ID   int    `json:"id"`
		UUID string `json:"uuid"`
		Name string `json:"name"`
	} `json:"video"`
}

type VideoAttributes struct {
	Name            string
	Category        int
	Licence         int
	Language        string
	Privacy         int
	Description     string
	Tags            []string
	DownloadEnabled bool
	CommentsEnabled bool
	WaitTranscoding bool
	NSFW            bool
}

func NewClient(baseURL, username, password string) *Client {
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Minute, // Long timeout for large uploads
		},
	}
}

func (c *Client) Authenticate() error {
	// First get client credentials
	resp, err := c.httpClient.Get(c.baseURL + "/api/v1/oauth-clients/local")
	if err != nil {
		return fmt.Errorf("getting oauth clients: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("oauth clients request failed: %s - %s", resp.Status, string(body))
	}

	var creds clientCredentials
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return fmt.Errorf("decoding client credentials: %w", err)
	}

	// Now get access token
	tokenData := fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=password&response_type=code&username=%s&password=%s",
		creds.ClientID, creds.ClientSecret, c.username, c.password)

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/users/token", strings.NewReader(tokenData))
	if err != nil {
		return fmt.Errorf("creating token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed: %s - %s", resp.Status, string(body))
	}

	var auth authResponse
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return fmt.Errorf("decoding auth response: %w", err)
	}

	c.token = auth.AccessToken
	return nil
}

func (c *Client) Upload(videoPath string, attrs VideoAttributes) (*uploadResponse, error) {
	if c.token == "" {
		if err := c.Authenticate(); err != nil {
			return nil, fmt.Errorf("authentication required: %w", err)
		}
	}

	file, err := os.Open(videoPath)
	if err != nil {
		return nil, fmt.Errorf("opening video file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add video file
	part, err := writer.CreateFormFile("videofile", filepath.Base(videoPath))
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copying file data: %w", err)
	}

	// Add metadata fields
	fields := map[string]string{
		"name":            attrs.Name,
		"category":        strconv.Itoa(attrs.Category),
		"licence":         strconv.Itoa(attrs.Licence),
		"language":        attrs.Language,
		"privacy":         strconv.Itoa(attrs.Privacy),
		"downloadEnabled": strconv.FormatBool(attrs.DownloadEnabled),
		"waitTranscoding": strconv.FormatBool(attrs.WaitTranscoding),
		"nsfw":            strconv.FormatBool(attrs.NSFW),
	}

	if attrs.Description != "" {
		fields["description"] = attrs.Description
	}

	if !attrs.CommentsEnabled {
		fields["commentsPolicy"] = "2" // DISABLED = 2
	}

	for key, val := range fields {
		if err := writer.WriteField(key, val); err != nil {
			return nil, fmt.Errorf("writing field %s: %w", key, err)
		}
	}

	// Add tags
	for _, tag := range attrs.Tags {
		if err := writer.WriteField("tags[]", tag); err != nil {
			return nil, fmt.Errorf("writing tag: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/videos/upload", body)
	if err != nil {
		return nil, fmt.Errorf("creating upload request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed: %s - %s", resp.Status, string(body))
	}

	var result uploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding upload response: %w", err)
	}

	return &result, nil
}
