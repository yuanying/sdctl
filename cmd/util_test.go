package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveImages_Single(t *testing.T) {
	dir := t.TempDir()
	b64 := "aGVsbG8=" // base64("hello")

	paths, err := saveImages([]string{b64}, dir)
	if err != nil {
		t.Fatalf("saveImages failed: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}
	if _, err := os.Stat(paths[0]); err != nil {
		t.Errorf("file does not exist: %s", paths[0])
	}
}

func TestSaveImages_MultipleWithDir(t *testing.T) {
	dir := t.TempDir()
	b64 := "aGVsbG8="

	paths, err := saveImages([]string{b64, b64, b64}, dir)
	if err != nil {
		t.Fatalf("saveImages failed: %v", err)
	}
	if len(paths) != 3 {
		t.Fatalf("expected 3 paths, got %d", len(paths))
	}
	for i, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("file %d does not exist: %s", i, p)
		}
	}
	if filepath.Base(paths[0]) == filepath.Base(paths[1]) {
		t.Errorf("expected different filenames, got same: %s", paths[0])
	}
}

func TestSaveImages_MultipleWithFilePath(t *testing.T) {
	dir := t.TempDir()
	b64 := "aGVsbG8="
	filePath := filepath.Join(dir, "output.png")

	paths, err := saveImages([]string{b64, b64, b64}, filePath)
	if err != nil {
		t.Fatalf("saveImages failed: %v", err)
	}
	if len(paths) != 3 {
		t.Fatalf("expected 3 paths, got %d", len(paths))
	}
	expected := []string{
		filepath.Join(dir, "output.0.png"),
		filepath.Join(dir, "output.1.png"),
		filepath.Join(dir, "output.2.png"),
	}
	for i, p := range paths {
		if p != expected[i] {
			t.Errorf("path[%d]: expected %s, got %s", i, expected[i], p)
		}
		if _, err := os.Stat(p); err != nil {
			t.Errorf("file does not exist: %s", p)
		}
	}
}

func TestSaveImages_MultipleWithFilePath_ZeroPadded(t *testing.T) {
	dir := t.TempDir()
	b64 := "aGVsbG8="
	filePath := filepath.Join(dir, "output.png")

	images := make([]string, 11)
	for i := range images {
		images[i] = b64
	}

	paths, err := saveImages(images, filePath)
	if err != nil {
		t.Fatalf("saveImages failed: %v", err)
	}
	if len(paths) != 11 {
		t.Fatalf("expected 11 paths, got %d", len(paths))
	}
	if filepath.Base(paths[0]) != "output.00.png" {
		t.Errorf("expected output.00.png, got %s", filepath.Base(paths[0]))
	}
	if filepath.Base(paths[10]) != "output.10.png" {
		t.Errorf("expected output.10.png, got %s", filepath.Base(paths[10]))
	}
}

func TestSaveImages_MultipleDefaultPath(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	b64 := "aGVsbG8="
	paths, err := saveImages([]string{b64, b64}, "")
	if err != nil {
		t.Fatalf("saveImages failed: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("file does not exist: %s", p)
		}
	}
}
