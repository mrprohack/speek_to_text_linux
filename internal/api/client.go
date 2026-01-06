// Package api provides Groq API integration for speech-to-text
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"VoiceType/pkg/errors"
	"VoiceType/pkg/wav"
)

// Client represents the Groq API client
type Client struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
	errHandler *errors.Handler
}

// NewClient creates a new API client
func NewClient(apiKey string, errHandler *errors.Handler) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.groq.com/openai/v1",
		model:   "whisper-large-v3",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		errHandler: errHandler,
	}
}

// Response represents the API response structure
type Response struct {
	ID       string    `json:"id"`
	Object   string    `json:"object"`
	Created  int64     `json:"created"`
	Model    string    `json:"model"`
	Task     string    `json:"task"`
	Audio    AudioData `json:"audio"`
	Text     string    `json:"text"`
	Language string    `json:"language"`
	Duration float64   `json:"duration"`
	Segments []Segment `json:"segments"`
}

// AudioData contains audio-related response data
type AudioData struct {
}

// Segment represents a segment of transcribed audio
type Segment struct {
	ID         int     `json:"id"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

// Transcribe sends audio data to the API for transcription
func (c *Client) Transcribe(ctx context.Context, audioData []byte) (string, error) {
	if len(audioData) == 0 {
		return "", errors.ErrAudioTooShort
	}

	// Encode audio as WAV
	wavData, err := wav.Encode(audioData, 16000, 1, 16)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeAPI, "failed to encode WAV")
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add audio file
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeAPI, "failed to create form file")
	}

	if _, err := io.Copy(part, bytes.NewReader(wavData)); err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeAPI, "failed to write audio data")
	}

	// Add other fields
	_ = writer.WriteField("model", c.model)
	_ = writer.WriteField("temperature", "0")
	_ = writer.WriteField("response_format", "verbose_json")
	// Add instruction prompt for better flow, punctuation, and cleanup (Wispr Flow style)
	_ = writer.WriteField("prompt", "Transcribe the audio accurately. Add appropriate punctuation and capitalization. Remove filler words like 'um', 'uh', 'ah'. Ensure the output is natural and professional.")

	if err := writer.Close(); err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeAPI, "failed to close form writer")
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/audio/transcriptions", body)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeAPI, "failed to create request")
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeNetwork, "request failed")
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", c.handleErrorResponse(resp)
	}

	// Parse response
	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeAPI, "failed to decode response")
	}

	return result.Text, nil
}

// handleErrorResponse handles API error responses
func (c *Client) handleErrorResponse(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case 401:
		return errors.ErrAPIKeyInvalid
	case 429:
		return errors.ErrRateLimited
	case 400:
		return fmt.Errorf("bad request: %s", string(body))
	case 500, 502, 503, 504:
		return fmt.Errorf("server error: %s", string(body))
	default:
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}
}

// SetModel sets the transcription model
func (c *Client) SetModel(model string) {
	c.model = model
}

// GetModel returns the current model
func (c *Client) GetModel() string {
	return c.model
}

// HealthCheck checks if the API is accessible
func (c *Client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/models", nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeNetwork, "failed to create health check request")
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeNetwork, "health check failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// Cleanup closes the client and releases resources
func (c *Client) Cleanup() {
	c.httpClient.CloseIdleConnections()
}

// Import os for file operations
var _ = os.Stdin
