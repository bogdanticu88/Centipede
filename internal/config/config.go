package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config represents the entire CENTIPEDE configuration
type Config struct {
	Cloud     string           `mapstructure:"cloud"`
	Azure     *AzureConfig     `mapstructure:"azure"`
	AWS       *AWSConfig       `mapstructure:"aws"`
	Detection *DetectionConfig `mapstructure:"detection"`
	Honeypots []HoneypotConfig `mapstructure:"honeypots"`
	Tenants   []TenantConfig   `mapstructure:"tenants"`
	Alert     *AlertConfig     `mapstructure:"alert"`
}

// AzureConfig holds Azure-specific settings
type AzureConfig struct {
	APIMName       string `mapstructure:"apim_name"`
	ResourceGroup  string `mapstructure:"resource_group"`
	SubscriptionID string `mapstructure:"subscription_id"`
	TenantID       string `mapstructure:"tenant_id"`
}

// AWSConfig holds AWS-specific settings
type AWSConfig struct {
	APIGatewayID string `mapstructure:"api_gateway_id"`
	Region       string `mapstructure:"region"`
}

// DetectionConfig holds detection rule thresholds
type DetectionConfig struct {
	VolumeThreshold    float64 `mapstructure:"volume_threshold"`
	PayloadThreshold   float64 `mapstructure:"payload_threshold"`
	ErrorRateThreshold float64 `mapstructure:"error_rate_threshold"`
	ScoreWarning       int     `mapstructure:"score_warning"`
	ScoreCritical      int     `mapstructure:"score_critical"`
}

// HoneypotConfig represents a honeypot endpoint
type HoneypotConfig struct {
	Path     string `mapstructure:"path"`
	Severity int    `mapstructure:"severity"`
}

// TenantConfig holds per-tenant configuration
type TenantConfig struct {
	ID           string   `mapstructure:"id"`
	Name         string   `mapstructure:"name"`
	Endpoints    []string `mapstructure:"endpoints"`
	RateLimitRPS int      `mapstructure:"rate_limit_rps"`
}

// AlertConfig holds alerting configuration
type AlertConfig struct {
	Type         string `mapstructure:"type"` // "slack", "webhook", "none"
	SlackWebhook string `mapstructure:"slack_webhook"`
	WebhookURL   string `mapstructure:"webhook_url"`
}

// LoadConfig reads and parses a YAML configuration file with env var overrides
func LoadConfig(path string) (*Config, error) {
	// Allow config path override via environment variable
	if envPath := os.Getenv("CENTIPEDE_CONFIG"); envPath != "" {
		path = envPath
	}

	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()            // Read from env vars
	viper.SetEnvPrefix("CENTIPEDE") // Prefix for env vars (e.g., CENTIPEDE_CLOUD)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply environment variable overrides
	if cloud := os.Getenv("CENTIPEDE_CLOUD"); cloud != "" {
		cfg.Cloud = cloud
	}

	if webhookURL := os.Getenv("CENTIPEDE_SLACK_WEBHOOK"); webhookURL != "" {
		if cfg.Alert == nil {
			cfg.Alert = &AlertConfig{}
		}
		cfg.Alert.SlackWebhook = webhookURL
		cfg.Alert.Type = "slack"
	}

	// Apply defaults for detection config if not provided
	if cfg.Detection == nil {
		cfg.Detection = &DetectionConfig{
			VolumeThreshold:    2.0,
			PayloadThreshold:   3.0,
			ErrorRateThreshold: 10.0,
			ScoreWarning:       2,
			ScoreCritical:      4,
		}
	}

	// Validate all configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Detection: &DetectionConfig{
			VolumeThreshold:    2.0,
			PayloadThreshold:   3.0,
			ErrorRateThreshold: 10.0,
			ScoreWarning:       2,
			ScoreCritical:      4,
		},
	}
}

// Validate validates the configuration for required fields and valid values
func (c *Config) Validate() error {
	// Validate cloud provider
	if c.Cloud == "" {
		return fmt.Errorf("cloud provider must be specified")
	}

	// Validate cloud-specific configuration
	switch c.Cloud {
	case "azure":
		if c.Azure == nil {
			return fmt.Errorf("azure configuration required when cloud=azure")
		}
		if err := c.Azure.Validate(); err != nil {
			return err
		}
	case "aws":
		if c.AWS == nil {
			return fmt.Errorf("aws configuration required when cloud=aws")
		}
	case "gcp":
		// GCP validation would go here
	default:
		return fmt.Errorf("unknown cloud provider: %s (must be 'azure', 'aws', or 'gcp')", c.Cloud)
	}

	// Validate detection config
	if c.Detection != nil {
		if err := c.Detection.Validate(); err != nil {
			return err
		}
	}

	// Validate honeypots
	for i, hp := range c.Honeypots {
		if hp.Path == "" {
			return fmt.Errorf("honeypot %d: path cannot be empty", i)
		}
		if hp.Severity < 0 {
			return fmt.Errorf("honeypot %d: severity cannot be negative", i)
		}
	}

	// Validate tenants
	for i, t := range c.Tenants {
		if t.ID == "" {
			return fmt.Errorf("tenant %d: id cannot be empty", i)
		}
		if t.RateLimitRPS < 0 {
			return fmt.Errorf("tenant %d: rate_limit_rps cannot be negative", i)
		}
	}

	return nil
}

// Validate validates Azure configuration
func (ac *AzureConfig) Validate() error {
	if ac.APIMName == "" {
		return fmt.Errorf("azure.apim_name is required")
	}
	if ac.ResourceGroup == "" {
		return fmt.Errorf("azure.resource_group is required")
	}
	if ac.SubscriptionID == "" {
		return fmt.Errorf("azure.subscription_id is required")
	}

	// Validate subscription ID is UUID format
	if len(ac.SubscriptionID) != 36 || ac.SubscriptionID[8] != '-' {
		return fmt.Errorf("azure.subscription_id does not appear to be a valid UUID")
	}

	return nil
}

// Validate validates detection configuration
func (dc *DetectionConfig) Validate() error {
	if dc.VolumeThreshold < 0 {
		return fmt.Errorf("detection.volume_threshold cannot be negative")
	}
	if dc.PayloadThreshold < 0 {
		return fmt.Errorf("detection.payload_threshold cannot be negative")
	}
	if dc.ErrorRateThreshold < 0 {
		return fmt.Errorf("detection.error_rate_threshold cannot be negative")
	}
	if dc.ScoreWarning < 0 {
		return fmt.Errorf("detection.score_warning cannot be negative")
	}
	if dc.ScoreCritical < 0 {
		return fmt.Errorf("detection.score_critical cannot be negative")
	}
	if dc.ScoreCritical <= dc.ScoreWarning {
		return fmt.Errorf("detection.score_critical must be greater than score_warning")
	}

	return nil
}
