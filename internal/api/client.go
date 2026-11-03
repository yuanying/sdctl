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
	token      string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithToken sends "Authorization: Bearer <token>" with every request.
// An empty token sends no Authorization header.
func WithToken(token string) Option {
	return func(c *Client) {
		c.token = token
	}
}

func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			// A redirect to another host is a login page, not the API: stop there
			// so the caller reports it instead of failing to decode its HTML.
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if req.URL.Host != via[0].URL.Host {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Txt2ImgRequest struct {
	Prompt                            string         `json:"prompt"`
	NegativePrompt                    string         `json:"negative_prompt,omitempty"`
	Steps                             int            `json:"steps"`
	Width                             int            `json:"width"`
	Height                            int            `json:"height"`
	CFGScale                          float64        `json:"cfg_scale"`
	DistilledCFGScale                 *float64       `json:"distilled_cfg_scale,omitempty"`
	SamplerName                       string         `json:"sampler_name,omitempty"`
	SchedulerName                     string         `json:"scheduler,omitempty"`
	Seed                              int64          `json:"seed,omitempty"`
	BatchCount                        int            `json:"n_iter,omitempty"`
	BatchSize                         int            `json:"batch_size,omitempty"`
	OverrideSettings                  map[string]any `json:"override_settings,omitempty"`
	OverrideSettingsRestoreAfterwards *bool          `json:"override_settings_restore_afterwards,omitempty"`
	AlwaysonScripts                   map[string]any `json:"alwayson_scripts,omitempty"`
	EnableHR                          bool           `json:"enable_hr,omitempty"`
	HRScale                           float64        `json:"hr_scale,omitempty"`
	HRUpscaler                        string         `json:"hr_upscaler,omitempty"`
	HRSecondPassSteps                 int            `json:"hr_second_pass_steps,omitempty"`
	DenoisingStrength                 float64        `json:"denoising_strength,omitempty"`
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

type Sampler struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}

type Scheduler struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type SDModule struct {
	ModelName string `json:"model_name"`
	Filename  string `json:"filename"`
}

type Upscaler struct {
	Name      string  `json:"name"`
	ModelName *string `json:"model_name"`
	Scale     float64 `json:"scale"`
}

func (c *Client) post(path string, body any) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

func (c *Client) get(path string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection refused (%s)", c.baseURL)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		if loc := resp.Header.Get("Location"); loc != "" && resp.StatusCode >= 300 && resp.StatusCode < 400 {
			return nil, fmt.Errorf("API error %d: redirected to %s (the WebUI may be behind a login; set SDCTL_TOKEN or token in the config file)", resp.StatusCode, loc)
		}
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

func (c *Client) ListSamplers() ([]Sampler, error) {
	resp, err := c.get("/sdapi/v1/samplers")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var samplers []Sampler
	if err := json.NewDecoder(resp.Body).Decode(&samplers); err != nil {
		return nil, err
	}
	return samplers, nil
}

func (c *Client) ListSchedulers() ([]Scheduler, error) {
	resp, err := c.get("/sdapi/v1/schedulers")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var schedulers []Scheduler
	if err := json.NewDecoder(resp.Body).Decode(&schedulers); err != nil {
		return nil, err
	}
	return schedulers, nil
}

func (c *Client) ListSDModules() ([]SDModule, error) {
	resp, err := c.get("/sdapi/v1/sd-modules")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var modules []SDModule
	if err := json.NewDecoder(resp.Body).Decode(&modules); err != nil {
		return nil, err
	}
	return modules, nil
}

func (c *Client) ListUpscalers() ([]Upscaler, error) {
	resp, err := c.get("/sdapi/v1/upscalers")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var upscalers []Upscaler
	if err := json.NewDecoder(resp.Body).Decode(&upscalers); err != nil {
		return nil, err
	}
	return upscalers, nil
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
