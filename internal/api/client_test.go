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
		InitImages:      []string{"base64inputimage"},
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
				JobCount:     1,
				JobNo:        0,
				SamplingStep: 10,
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

func TestConnectionError(t *testing.T) {
	client := api.NewClient("http://localhost:1")
	_, err := client.ListModels()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
