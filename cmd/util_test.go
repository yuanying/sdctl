package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestResolveUniqueFilePath_NoFile_NotBatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.png")

	got := resolveUniqueFilePath(path, false)
	if got != path {
		t.Errorf("expected %s, got %s", path, got)
	}
}

func TestResolveUniqueFilePath_NoFile_Batch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.png")

	got := resolveUniqueFilePath(path, true)
	expected := filepath.Join(dir, "output.0001.png")
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestResolveUniqueFilePath_FileExists_NotBatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.png")
	os.WriteFile(path, []byte("x"), 0644)

	got := resolveUniqueFilePath(path, false)
	expected := filepath.Join(dir, "output.0001.png")
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestResolveUniqueFilePath_FileAndIndexedExist_NotBatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.png")
	os.WriteFile(path, []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, "output.0001.png"), []byte("x"), 0644)

	got := resolveUniqueFilePath(path, false)
	expected := filepath.Join(dir, "output.0002.png")
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestResolveUniqueFilePath_Batch_ExistingIndexed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.png")
	os.WriteFile(filepath.Join(dir, "output.0001.png"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, "output.0003.png"), []byte("x"), 0644)

	got := resolveUniqueFilePath(path, true)
	expected := filepath.Join(dir, "output.0004.png")
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestResolveUniqueFilePath_IndexPaddingExpands(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.png")
	// Simulate max index at 9999
	os.WriteFile(filepath.Join(dir, "output.9999.png"), []byte("x"), 0644)

	got := resolveUniqueFilePath(path, true)
	expected := filepath.Join(dir, "output.10000.png")
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestSaveImages_Single_NoConflict(t *testing.T) {
	dir := t.TempDir()
	b64 := "aGVsbG8="
	filePath := filepath.Join(dir, "output.png")

	paths, err := saveImages([]string{b64}, filePath)
	if err != nil {
		t.Fatalf("saveImages failed: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}
	if paths[0] != filePath {
		t.Errorf("expected %s, got %s", filePath, paths[0])
	}
}

func TestSaveImages_Single_WithConflict(t *testing.T) {
	dir := t.TempDir()
	b64 := "aGVsbG8="
	filePath := filepath.Join(dir, "output.png")
	os.WriteFile(filePath, []byte("existing"), 0644)

	paths, err := saveImages([]string{b64}, filePath)
	if err != nil {
		t.Fatalf("saveImages failed: %v", err)
	}
	expected := filepath.Join(dir, "output.0001.png")
	if paths[0] != expected {
		t.Errorf("expected %s, got %s", expected, paths[0])
	}
}

func TestSaveImages_Multiple_FilePath(t *testing.T) {
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
		filepath.Join(dir, "output.0001.png"),
		filepath.Join(dir, "output.0002.png"),
		filepath.Join(dir, "output.0003.png"),
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

func TestSaveImages_Multiple_FilePath_ExistingFiles(t *testing.T) {
	dir := t.TempDir()
	b64 := "aGVsbG8="
	filePath := filepath.Join(dir, "output.png")
	os.WriteFile(filepath.Join(dir, "output.0001.png"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, "output.0002.png"), []byte("x"), 0644)

	paths, err := saveImages([]string{b64, b64}, filePath)
	if err != nil {
		t.Fatalf("saveImages failed: %v", err)
	}
	expected := []string{
		filepath.Join(dir, "output.0003.png"),
		filepath.Join(dir, "output.0004.png"),
	}
	for i, p := range paths {
		if p != expected[i] {
			t.Errorf("path[%d]: expected %s, got %s", i, expected[i], p)
		}
	}
}

func TestSaveImages_Single_DirOutput(t *testing.T) {
	dir := t.TempDir()
	b64 := "aGVsbG8="

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

func TestMergeMap_BothNil(t *testing.T) {
	if got := mergeMap(nil, nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestMergeMap_BaseOnly(t *testing.T) {
	base := map[string]any{"a": 1}
	got := mergeMap(base, nil)
	if !reflect.DeepEqual(got, base) {
		t.Errorf("expected %v, got %v", base, got)
	}
}

func TestMergeMap_OverrideOnly(t *testing.T) {
	override := map[string]any{"a": 2}
	got := mergeMap(nil, override)
	if !reflect.DeepEqual(got, override) {
		t.Errorf("expected %v, got %v", override, got)
	}
}

func TestMergeMap_OverrideTakesPrecedence(t *testing.T) {
	base := map[string]any{"a": 1, "b": 2}
	override := map[string]any{"b": 99, "c": 3}
	got := mergeMap(base, override)
	expected := map[string]any{"a": 1, "b": 99, "c": 3}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestBuildAdditionalModules_BothSet(t *testing.T) {
	got := buildAdditionalModules("/path/to/vae.safetensors", "/path/to/te.safetensors")
	expected := map[string]any{
		"forge_additional_modules": []string{"/path/to/vae.safetensors", "/path/to/te.safetensors"},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestBuildAdditionalModules_VAEOnly(t *testing.T) {
	got := buildAdditionalModules("/path/to/vae.safetensors", "")
	expected := map[string]any{
		"forge_additional_modules": []string{"/path/to/vae.safetensors"},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestBuildAdditionalModules_TextEncoderOnly(t *testing.T) {
	got := buildAdditionalModules("", "/path/to/te.safetensors")
	expected := map[string]any{
		"forge_additional_modules": []string{"/path/to/te.safetensors"},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestBuildAdditionalModules_NoneSet(t *testing.T) {
	got := buildAdditionalModules("", "")
	if got != nil {
		t.Errorf("expected nil, got %v", got)
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
