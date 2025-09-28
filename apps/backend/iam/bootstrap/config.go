package bootstrap

import (
	"os"
	"path/filepath"
	"strings"

	envFetcher "github.com/adverax/metacrm/pkg/access/fetchers/maps/env"
	yamlConfig "github.com/adverax/metacrm/pkg/configs/formats/yaml"
	"github.com/adverax/metacrm/pkg/database/sql"
	"github.com/jackc/pgx/v5/pgconn"
)

type ApiConfig struct {
	Port int
}

type DbConfig struct {
	sql.DSN
	Dsn string `yaml:"dsn" json:"dsn"` // Data Source Name
}

func (that *DbConfig) Init() error {
	if that.Dsn == "" {
		return nil
	}

	dsn, err := pgconn.ParseConfig(that.Dsn)
	if err != nil {
		return err
	}
	that.DSN = sql.DSN{
		Host:     dsn.Host,
		Port:     dsn.Port,
		User:     dsn.User,
		Password: dsn.Password,
		Database: dsn.Database,
	}

	return nil
}

type LogConfig struct {
	Level  string `yaml:"level" json:"level"`   // Log level
	Output string `yaml:"output" json:"output"` // Log output destination (e.g., "stdout", "stderr")
	Format string `yaml:"format" json:"format"` // Log format (e.g., "json", "text")
}

type Config struct {
	Env string    `yaml:"env" json:"env"` // Application environment (e.g., "development", "production", etc.)
	DB  DbConfig  `yaml:"db" json:"db"`
	Api ApiConfig `yaml:"api" json:"api"`
	Log LogConfig `yaml:"log" json:"log"`
}

func (that *Config) IsDevEnv() bool {
	env := strings.ToLower(that.Env)
	return env == "dev" || env == "development"
}

func (that *Config) Load() error {
	globalConfig, err := getConfigPath()
	if err != nil {
		return err
	}

	if globalConfig != "" {
		ext := filepath.Ext(globalConfig)
		localConfig := strings.TrimSuffix(globalConfig, ext) + ".local" + ext
		loader, err := yamlConfig.NewFileLoaderBuilder().
			WithFile(globalConfig, false).
			WithFile(localConfig, false).
			WithSource(
				envFetcher.New(
					envFetcher.NewPrefixGuard("META_"),
					envFetcher.NewKeyPathAccumulator("_"),
				),
			).
			Build()
		if err != nil {
			return err
		}

		err = loader.Load(that)
		if err != nil {
			return err
		}
	}

	return nil
}

func (that *Config) Init() error {
	err := that.DB.Init()
	if err != nil {
		return err
	}

	return nil
}

func (that *Config) Validate() error {
	return nil
}

func getConfigPath() (string, error) {
	path := os.Getenv("META_CONFIG_PATH")
	if path != "" {
		return path, nil
	}

	const filename = "config.yaml"

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		configPath := filepath.Join(dir, filename)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}

		dir = parent
	}
}

func DefaultConfig() *Config {
	return &Config{
		Env: "development",
		Api: ApiConfig{
			Port: 8080,
		},
		DB: DbConfig{
			DSN: sql.DSN{
				Host:     "localhost",
				Port:     5439,
				User:     "postgres",
				Password: "",
				Database: "iam",
			},
		},
		Log: LogConfig{
			Level:  "info",
			Output: "stdout",
			Format: "text",
		},
	}
}
