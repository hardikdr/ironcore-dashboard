package config

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"go.yaml.in/yaml/v3"
)

type SourceRole string

const (
	RoleZonalCAPI       SourceRole = "zonal-capi"
	RoleCentralIronCore SourceRole = "central-ironcore"
	RoleGardener        SourceRole = "gardener"
)

type Config struct {
	Enabled        bool          `yaml:"enabled"`
	PollInterval   time.Duration `yaml:"pollInterval"`
	StaleAfter     time.Duration `yaml:"staleAfter"`
	RequestTimeout time.Duration `yaml:"requestTimeout"`
	QPS            float32       `yaml:"qps"`
	Burst          int           `yaml:"burst"`
	MaxConcurrent  int           `yaml:"maxConcurrent"`
	Sources        []Source      `yaml:"sources"`
}

type Source struct {
	ID                 string     `yaml:"id"`
	Environment        string     `yaml:"environment"`
	Role               SourceRole `yaml:"role"`
	Zone               string     `yaml:"zone,omitempty"`
	InstallationRole   string     `yaml:"installationRole,omitempty"`
	Context            string     `yaml:"context,omitempty"`
	Kubeconfig         string     `yaml:"kubeconfig"`
	ProviderNamespaces []string   `yaml:"providerNamespaces,omitempty"`
}

var safeID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open SRE configuration: %w", err)
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode SRE configuration: %w", err)
	}
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (cfg *Config) applyDefaults() {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 30 * time.Second
	}
	if cfg.StaleAfter == 0 {
		cfg.StaleAfter = 2 * time.Minute
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 15 * time.Second
	}
	if cfg.QPS == 0 {
		cfg.QPS = 5
	}
	if cfg.Burst == 0 {
		cfg.Burst = 10
	}
	if cfg.MaxConcurrent == 0 {
		cfg.MaxConcurrent = 3
	}
}

func (cfg Config) Validate() error {
	if !cfg.Enabled {
		if len(cfg.Sources) != 0 {
			return fmt.Errorf("disabled SRE configuration cannot declare sources")
		}
		return nil
	}
	if cfg.PollInterval <= 0 || cfg.StaleAfter <= 0 || cfg.RequestTimeout <= 0 {
		return fmt.Errorf("pollInterval, staleAfter, and requestTimeout must be positive")
	}
	if cfg.QPS <= 0 || cfg.Burst <= 0 || cfg.MaxConcurrent <= 0 {
		return fmt.Errorf("qps, burst, and maxConcurrent must be positive")
	}
	if len(cfg.Sources) == 0 {
		return fmt.Errorf("enabled SRE configuration requires at least one source")
	}

	seen := make(map[string]struct{}, len(cfg.Sources))
	for index, source := range cfg.Sources {
		if !safeID.MatchString(source.ID) {
			return fmt.Errorf("source %d has an invalid id", index)
		}
		if _, duplicate := seen[source.ID]; duplicate {
			return fmt.Errorf("source %q is declared more than once", source.ID)
		}
		seen[source.ID] = struct{}{}
		if source.Environment == "" {
			return fmt.Errorf("source %q requires an environment", source.ID)
		}
		if source.Kubeconfig == "" {
			return fmt.Errorf("source %q requires a kubeconfig file", source.ID)
		}
		switch source.Role {
		case RoleZonalCAPI:
			if source.Zone == "" {
				return fmt.Errorf("zonal source %q requires a zone", source.ID)
			}
		case RoleCentralIronCore:
			if source.Zone != "" {
				return fmt.Errorf("central source %q cannot declare a single zone", source.ID)
			}
		case RoleGardener:
			if source.InstallationRole == "" {
				return fmt.Errorf("Gardener source %q requires installationRole", source.ID)
			}
		default:
			return fmt.Errorf("source %q has unsupported role %q", source.ID, source.Role)
		}
		if err := validateSourceFile(source); err != nil {
			return err
		}
	}
	return nil
}

func validateSourceFile(source Source) error {
	info, err := os.Stat(source.Kubeconfig)
	if err != nil {
		return fmt.Errorf("source %q kubeconfig is unavailable", source.ID)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("source %q kubeconfig is not a regular file", source.ID)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("source %q kubeconfig permissions are too broad", source.ID)
	}
	return nil
}
