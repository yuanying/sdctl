package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yuanying/sdctl/internal/api"
)

func TestTxt2Img(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sdapi/v1/txt2img" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		resp := api.GenerateResponse{
			Images: []string{"base64encodedimage"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	req := api.Txt2ImgRequest{
		Prompt:         "a cat",
		NegativePrompt: "",
		Steps:          20,
		Width:          512,
		Height:         512,
		CFGScale:       7.0,
		SamplerName:    "Euler a",
	}
	resp, err := client.Txt2Img(req)
	if err != nil {
		t.Fatalf("Txt2Img failed: %v", err)
	}
	if len(resp.Images) != 1 {
		t.Errorf("expected 1 image, got %d", len(resp.Images))
	}
}

func TestImg2Img(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sdapi/v1/img2img" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		resp := api.GenerateResponse{
			Images: []string{"base64encodedimage"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	req := api.Img2ImgRequest{
		Txt2ImgRequest: api.Txt2ImgRequest{
			Prompt: "a dog",
			Steps:  20,
			Width:  512,
			Height: 512,
		},
		InitImages:        []string{"base64inputimage"},
		DenoisingStrength: 0.75,
	}
	resp, err := client.Img2Img(req)
	if err != nil {
		t.Fatalf("Img2Img failed: %v", err)
	}
	if len(resp.Images) != 1 {
		t.Errorf("expected 1 image, got %d", len(resp.Images))
	}
}

func TestGetProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sdapi/v1/progress" || r.Method != http.MethodGet {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		resp := api.ProgressResponse{
			Progress:    0.5,
			EtaRelative: 3.0,
			State: api.ProgressState{
				JobCount:      1,
				JobNo:         0,
				SamplingStep:  10,
				SamplingSteps: 20,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	resp, err := client.GetProgress()
	if err != nil {
		t.Fatalf("GetProgress failed: %v", err)
	}
	if resp.Progress != 0.5 {
		t.Errorf("expected progress 0.5, got %f", resp.Progress)
	}
}

func TestListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sdapi/v1/sd-models" || r.Method != http.MethodGet {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		resp := []api.Model{
			{Title: "v1-5-pruned [abc123]", ModelName: "v1-5-pruned", Hash: "abc123"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	models, err := client.ListModels()
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if len(models) != 1 {
		t.Errorf("expected 1 model, got %d", len(models))
	}
	if models[0].ModelName != "v1-5-pruned" {
		t.Errorf("unexpected model name: %s", models[0].ModelName)
	}
}

func TestSetModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sdapi/v1/options" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["sd_model_checkpoint"] != "v1-5-pruned" {
			t.Errorf("unexpected model: %s", body["sd_model_checkpoint"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	err := client.SetModel("v1-5-pruned")
	if err != nil {
		t.Fatalf("SetModel failed: %v", err)
	}
}

func TestTxt2ImgWithBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		if v, ok := body["n_iter"].(float64); !ok || v != 3 {
			t.Errorf("expected n_iter 3, got %v", body["n_iter"])
		}
		if v, ok := body["batch_size"].(float64); !ok || v != 2 {
			t.Errorf("expected batch_size 2, got %v", body["batch_size"])
		}
		images := make([]string, 6)
		for i := range images {
			images[i] = "base64image"
		}
		resp := api.GenerateResponse{Images: images}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	req := api.Txt2ImgRequest{
		Prompt:     "a cat",
		Steps:      20,
		Width:      512,
		Height:     512,
		BatchCount: 3,
		BatchSize:  2,
	}
	resp, err := client.Txt2Img(req)
	if err != nil {
		t.Fatalf("Txt2Img failed: %v", err)
	}
	if len(resp.Images) != 6 {
		t.Errorf("expected 6 images, got %d", len(resp.Images))
	}
}

func TestListSamplers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sdapi/v1/samplers" || r.Method != http.MethodGet {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		resp := []api.Sampler{
			{Name: "Euler a", Aliases: []string{"k_euler_a"}},
			{Name: "DPM++ 2M", Aliases: []string{"k_dpmpp_2m"}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	samplers, err := client.ListSamplers()
	if err != nil {
		t.Fatalf("ListSamplers failed: %v", err)
	}
	if len(samplers) != 2 {
		t.Errorf("expected 2 samplers, got %d", len(samplers))
	}
	if samplers[0].Name != "Euler a" {
		t.Errorf("unexpected sampler name: %s", samplers[0].Name)
	}
}

func TestListSchedulers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sdapi/v1/schedulers" || r.Method != http.MethodGet {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		resp := []api.Scheduler{
			{Name: "automatic", Label: "Automatic"},
			{Name: "karras", Label: "Karras"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	schedulers, err := client.ListSchedulers()
	if err != nil {
		t.Fatalf("ListSchedulers failed: %v", err)
	}
	if len(schedulers) != 2 {
		t.Errorf("expected 2 schedulers, got %d", len(schedulers))
	}
	if schedulers[0].Name != "automatic" {
		t.Errorf("unexpected scheduler name: %s", schedulers[0].Name)
	}
}

func TestTxt2ImgWithScheduler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		if v, ok := body["scheduler"].(string); !ok || v != "karras" {
			t.Errorf("expected scheduler karras, got %v", body["scheduler"])
		}
		resp := api.GenerateResponse{Images: []string{"base64image"}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	req := api.Txt2ImgRequest{
		Prompt:        "a cat",
		Steps:         20,
		Width:         512,
		Height:        512,
		SchedulerName: "karras",
	}
	resp, err := client.Txt2Img(req)
	if err != nil {
		t.Fatalf("Txt2Img failed: %v", err)
	}
	if len(resp.Images) != 1 {
		t.Errorf("expected 1 image, got %d", len(resp.Images))
	}
}

func TestConnectionError(t *testing.T) {
	client := api.NewClient("http://localhost:1")
	_, err := client.ListModels()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
