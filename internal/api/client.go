package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

type Txt2ImgRequest struct {
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt,omitempty"`
	Steps          int     `json:"steps"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	CFGScale       float64 `json:"cfg_scale"`
	SamplerName    string  `json:"sampler_name,omitempty"`
	Seed           int64   `json:"seed,omitempty"`
}

type Img2ImgRequest struct {
	Txt2ImgRequest
	InitImages        []string `json:"init_images"`
	DenoisingStrength float64  `json:"denoising_strength"`
}

type GenerateResponse struct {
	Images []string `json:"images"`
	Info   string   `json:"info"`
}

type ProgressState struct {
	JobCount      int `json:"job_count"`
	JobNo         int `json:"job_no"`
	SamplingStep  int `json:"sampling_step"`
	SamplingSteps int `json:"sampling_steps"`
}

type ProgressResponse struct {
	Progress    float64       `json:"progress"`
	EtaRelative float64       `json:"eta_relative"`
	State       ProgressState `json:"state"`
}

type Model struct {
	Title     string `json:"title"`
	ModelName string `json:"model_name"`
	Hash      string `json:"hash"`
}

func (c *Client) post(path string, body any) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Post(c.baseURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("connection refused (%s)", c.baseURL)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return resp, nil
}

func (c *Client) get(path string) (*http.Response, error) {
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("connection refused (%s)", c.baseURL)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return resp, nil
}

func (c *Client) Txt2Img(req Txt2ImgRequest) (*GenerateResponse, error) {
	resp, err := c.post("/sdapi/v1/txt2img", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Img2Img(req Img2ImgRequest) (*GenerateResponse, error) {
	resp, err := c.post("/sdapi/v1/img2img", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetProgress() (*ProgressResponse, error) {
	resp, err := c.get("/sdapi/v1/progress")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result ProgressResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListModels() ([]Model, error) {
	resp, err := c.get("/sdapi/v1/sd-models")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var models []Model
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, err
	}
	return models, nil
}

func (c *Client) SetModel(modelName string) error {
	body := map[string]string{"sd_model_checkpoint": modelName}
	resp, err := c.post("/sdapi/v1/options", body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
