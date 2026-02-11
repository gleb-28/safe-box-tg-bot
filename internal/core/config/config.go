package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	TgBotToken            string         `env:"TG_BOT_TOKEN" env-required:"true"`
	LoggerBotToken        string         `env:"LOGGER_BOT_TOKEN"`
	AdminID               int64          `env:"ADMIN_ID" env-required:"true"`
	ActivationKey         string         `env:"ACTIVATION_KEY" env-required:"true"`
	OpenRouterModelApiKey string         `env:"OPENROUTER_API_KEY" env-required:"true"`
	OpenRouterModelName   string         `env:"OPENROUTER_MODEL_NAME" env-default:"openrouter/auto"`
	GroqAPIKey            string         `env:"GROQ_API_KEY"`
	GroqModelName         string         `env:"GROQ_MODEL_NAME" env-default:"groq/compound"`
	PromptPath            string         `env:"PROMPT_PATH" env-default:"data/prompt"`
	IsDebug               bool           `env:"IS_DEBUG" env-default:"false"`
	ForcePreviewFallback  bool           `env:"FORCE_PREVIEW_FALLBACK" env-default:"false"`
	Database              DatabaseConfig `env-required:"true"`
}

type DatabaseConfig struct {
	FileName string `env:"DB_FILE_NAME" env-required:"true"`
}

func MustConfig() *AppConfig {
	config := &AppConfig{}

	if os.Getenv("IS_DOCKERIZED") != "true" {
		if err := cleanenv.ReadConfig(".env", config); err != nil {
			log.Fatal("Error loading .env file (might be missing in non-dockerized env): ", err.Error())
		}
	}

	if err := cleanenv.ReadEnv(config); err != nil {
		log.Fatal("Error reading environment variables: ", err.Error())
	}

	return config
}
