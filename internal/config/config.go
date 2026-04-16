// Package config persists llml runtime and discovery cache in human-readable TOML.
// Parameter profiles remain in model-params.json; see [github.com/flyingnobita/llml/internal/tui].
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/flyingnobita/llml/internal/llamacpp"
)

// SchemaVersion is the current on-disk format for config.toml.
const SchemaVersion = 1

// Config is the root document stored at [ConfigPath].
type Config struct {
	SchemaVersion int             `toml:"schema_version"`
	Runtime       RuntimeConfig   `toml:"runtime"`
	Discovery     DiscoveryConfig `toml:"discovery"`
	Models        []ModelEntry    `toml:"models"`
}

// RuntimeConfig mirrors env vars LLAMA_CPP_PATH, VLLM_PATH, VLLM_VENV, and server ports.
// Empty strings mean unset; ports use pointers so zero can mean "omit default in file".
type RuntimeConfig struct {
	LlamaCppPath    string `toml:"llama_cpp_path"`
	VLLMPath        string `toml:"vllm_path"`
	VLLMVenv        string `toml:"vllm_venv"`
	LlamaServerPort *int   `toml:"llama_server_port,omitempty"`
	VLLMServerPort  *int   `toml:"vllm_server_port,omitempty"`
}

// DiscoveryConfig holds extra search roots and the last full filesystem scan time.
type DiscoveryConfig struct {
	ExtraModelPaths []string  `toml:"extra_model_paths"`
	LastScan        time.Time `toml:"last_scan"`
}

// ModelEntry is one cached model row from discovery.
type ModelEntry struct {
	Backend    string    `toml:"backend"`
	Path       string    `toml:"path"`
	Name       string    `toml:"name"`
	Size       int64     `toml:"size"`
	ModTime    time.Time `toml:"mod_time"`
	Parameters string    `toml:"parameters"`
}

// ConfigPath returns {UserConfigDir}/llml/config.toml.
func ConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "llml", "config.toml"), nil
}

// ReadFile reads and parses config.toml, or returns an empty error if the file is missing.
func ReadFile() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
		return Config{}, err
	}
	var c Config
	if _, err := toml.Decode(string(b), &c); err != nil {
		return Config{}, err
	}
	return c, nil
}

// ValidForCache reports whether the file is usable for instant startup (skip filesystem walk).
func (c Config) ValidForCache() bool {
	if c.SchemaVersion != SchemaVersion {
		return false
	}
	if len(c.Models) < 1 {
		return false
	}
	return true
}

// ApplyRuntimeFromConfig sets process environment from [runtime] only where the
// corresponding env var is currently unset (env wins over TOML).
func ApplyRuntimeFromConfig(r *RuntimeConfig) {
	if r == nil {
		return
	}
	applyPathIfUnset(llamacpp.EnvLlamaCppPath, r.LlamaCppPath)
	applyPathIfUnset(llamacpp.EnvVLLMPath, r.VLLMPath)
	applyPathIfUnset(llamacpp.EnvVLLMVenv, r.VLLMVenv)
	if r.LlamaServerPort != nil && os.Getenv(llamacpp.EnvLlamaServerPort) == "" {
		os.Setenv(llamacpp.EnvLlamaServerPort, strconv.Itoa(*r.LlamaServerPort))
	}
	if r.VLLMServerPort != nil && os.Getenv(llamacpp.EnvVLLMServerPort) == "" {
		os.Setenv(llamacpp.EnvVLLMServerPort, strconv.Itoa(*r.VLLMServerPort))
	}
}

func applyPathIfUnset(key, value string) {
	v := strings.TrimSpace(value)
	if v == "" || os.Getenv(key) != "" {
		return
	}
	v = filepath.Clean(llamacpp.ExpandTildePath(v))
	if v == "" || v == "." {
		return
	}
	os.Setenv(key, v)
}

// RuntimeFromEnv builds a RuntimeConfig from the current process environment (for writing).
func RuntimeFromEnv() RuntimeConfig {
	var r RuntimeConfig
	if v := strings.TrimSpace(os.Getenv(llamacpp.EnvLlamaCppPath)); v != "" {
		r.LlamaCppPath = filepath.Clean(llamacpp.ExpandTildePath(v))
	}
	if v := strings.TrimSpace(os.Getenv(llamacpp.EnvVLLMPath)); v != "" {
		r.VLLMPath = filepath.Clean(llamacpp.ExpandTildePath(v))
	}
	if v := strings.TrimSpace(os.Getenv(llamacpp.EnvVLLMVenv)); v != "" {
		r.VLLMVenv = filepath.Clean(llamacpp.ExpandTildePath(v))
	}
	if v := strings.TrimSpace(os.Getenv(llamacpp.EnvLlamaServerPort)); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 && p <= 65535 {
			r.LlamaServerPort = &p
		}
	} else {
		p := llamacpp.ListenPort()
		r.LlamaServerPort = &p
	}
	if v := strings.TrimSpace(os.Getenv(llamacpp.EnvVLLMServerPort)); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 && p <= 65535 {
			r.VLLMServerPort = &p
		}
	} else {
		p := llamacpp.VLLMPort()
		r.VLLMServerPort = &p
	}
	return r
}

// ExtraModelPathsFromEnv returns comma-separated LLML_MODEL_PATHS entries.
func ExtraModelPathsFromEnv() []string {
	v := strings.TrimSpace(os.Getenv("LLML_MODEL_PATHS"))
	if v == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(v, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// MergeExtraRoots combines discovery extra paths from config with env-only extras for Discover options.
// Config file paths are merged with env in [llamacpp.MergeSearchRoots] via Options.ExtraRoots.
func MergeExtraRoots(discoveryExtra, envExtra []string) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(p string) {
		p = filepath.Clean(llamacpp.ExpandTildePath(strings.TrimSpace(p)))
		if p == "" || p == "." {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	for _, p := range discoveryExtra {
		add(p)
	}
	for _, p := range envExtra {
		add(p)
	}
	return out
}

// ModelEntryFromFile converts a discovered model to a cache entry.
func ModelEntryFromFile(f llamacpp.ModelFile) ModelEntry {
	be := "llama"
	if f.Backend == llamacpp.BackendVLLM {
		be = "vllm"
	}
	return ModelEntry{
		Backend:    be,
		Path:       f.Path,
		Name:       f.Name,
		Size:       f.Size,
		ModTime:    f.ModTime,
		Parameters: f.Parameters,
	}
}

// ToModelFile converts a cache entry to [llamacpp.ModelFile].
func (e ModelEntry) ToModelFile() (llamacpp.ModelFile, error) {
	var be llamacpp.ModelBackend
	switch strings.ToLower(strings.TrimSpace(e.Backend)) {
	case "llama", "":
		be = llamacpp.BackendLlama
	case "vllm":
		be = llamacpp.BackendVLLM
	default:
		return llamacpp.ModelFile{}, fmt.Errorf("unknown backend %q", e.Backend)
	}
	path := filepath.Clean(e.Path)
	if path == "" || path == "." {
		return llamacpp.ModelFile{}, errors.New("empty model path")
	}
	return llamacpp.ModelFile{
		Backend:    be,
		Path:       path,
		Name:       e.Name,
		Size:       e.Size,
		ModTime:    e.ModTime,
		Parameters: e.Parameters,
	}, nil
}

// ModelFilesFromEntries converts cache entries to model files, skipping invalid rows.
func ModelFilesFromEntries(entries []ModelEntry) []llamacpp.ModelFile {
	var out []llamacpp.ModelFile
	for _, e := range entries {
		f, err := e.ToModelFile()
		if err != nil {
			continue
		}
		out = append(out, f)
	}
	return out
}

// FilterExistingPaths keeps only models whose path still exists on disk.
func FilterExistingPaths(files []llamacpp.ModelFile) []llamacpp.ModelFile {
	var out []llamacpp.ModelFile
	for _, f := range files {
		if _, err := os.Stat(f.Path); err != nil {
			continue
		}
		out = append(out, f)
	}
	return out
}

// BuildConfig builds a full Config for writing from runtime, discovery, and models.
func BuildConfig(runtime RuntimeConfig, discovery DiscoveryConfig, files []llamacpp.ModelFile) Config {
	c := Config{
		SchemaVersion: SchemaVersion,
		Runtime:       runtime,
		Discovery:     discovery,
	}
	for _, f := range files {
		c.Models = append(c.Models, ModelEntryFromFile(f))
	}
	return c
}

// DiscoveryConfigForWrite merges extra model paths from a previous on-disk config with
// current LLML_MODEL_PATHS so hand-edited TOML entries are preserved across writes.
func DiscoveryConfigForWrite(prev *Config, lastScan time.Time) DiscoveryConfig {
	var fromFile []string
	if prev != nil {
		fromFile = prev.Discovery.ExtraModelPaths
	}
	return DiscoveryConfig{
		ExtraModelPaths: MergeExtraRoots(fromFile, ExtraModelPathsFromEnv()),
		LastScan:        lastScan,
	}
}

// WriteFile writes config.toml atomically (write temp + rename).
func WriteFile(c Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	c.SchemaVersion = SchemaVersion
	var buf strings.Builder
	if err := toml.NewEncoder(&buf).Encode(c); err != nil {
		return err
	}
	data := []byte(buf.String())
	tmp, err := os.CreateTemp(filepath.Dir(path), "llml-config-*.toml")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Chmod(tmpPath, 0o644); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}
