package config

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port           string
	JWTSecret      string
	AdminEmail     string
	AdminPassword  string
	AllowedOrigins []string
	Database       DatabaseConfig
	Supabase       SupabaseConfig
	S3             S3Config
}

type DatabaseConfig struct {
	URL string
}

type SupabaseConfig struct {
	URL            string
	JWKSURL        string
	KeyID          string
	PublicKeyX     string
	PublicKeyY     string
	PublishableKey string
	SecretKey      string
}

func (s *SupabaseConfig) IsAuthEnabled() bool {
	return s.JWKSURL != "" || (s.PublicKeyX != "" && s.PublicKeyY != "")
}

type S3Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicURL       string
	UseSSL          bool
	ForcePathStyle  bool
}

func (s *S3Config) IsConfigured() bool {
	return s.AccessKeyID != "" && s.SecretAccessKey != "" && s.BucketName != ""
}

func Load() *Config {
	loadEnvFile(".env", "../.env")

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		AdminEmail:     getEnv("ADMIN_EMAIL", ""),
		AdminPassword:  getEnv("ADMIN_PASSWORD", ""),
		AllowedOrigins: getEnvList("ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://127.0.0.1:3000"}),
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", ""),
		},
		Supabase: SupabaseConfig{
			URL:            getEnv("SUPABASE_URL", ""),
			JWKSURL:        getEnv("SUPABASE_JWKS_URL", ""),
			KeyID:          getEnv("SUPABASE_JWT_KEY_ID", ""),
			PublicKeyX:     getEnv("SUPABASE_JWT_PUBLIC_KEY_X", ""),
			PublicKeyY:     getEnv("SUPABASE_JWT_PUBLIC_KEY_Y", ""),
			PublishableKey: getEnv("SUPABASE_PUBLISHABLE_KEY", ""),
			SecretKey:      getEnv("SUPABASE_SECRET_KEY", ""),
		},
		S3: S3Config{
			Endpoint:        getEnv("S3_ENDPOINT", ""),
			Region:          getEnv("S3_REGION", "auto"),
			AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", ""),
			BucketName:      getEnv("S3_BUCKET_NAME", ""),
			PublicURL:       getEnv("S3_PUBLIC_URL", ""),
			UseSSL:          getEnvBool("S3_USE_SSL", true),
			ForcePathStyle:  getEnvBool("S3_FORCE_PATH_STYLE", false),
		},
	}

	if cfg.S3.IsConfigured() {
		log.Printf("S3 configured: bucket=%s endpoint=%s public_url=%s", cfg.S3.BucketName, cfg.S3.Endpoint, cfg.S3.PublicURL)
	} else {
		log.Println("S3 not fully configured. Set S3_ACCESS_KEY_ID, S3_SECRET_ACCESS_KEY, and S3_BUCKET_NAME to enable cloud uploads.")
	}

	if cfg.Database.URL != "" {
		log.Println("PostgreSQL database URL configured")
	} else {
		log.Println("DATABASE_URL not set. Running with local in-memory store.")
	}

	if cfg.Supabase.IsAuthEnabled() {
		log.Printf("Supabase ECC Auth configured: discovery=%s kid=%s", cfg.Supabase.JWKSURL, cfg.Supabase.KeyID)
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}

func getEnvList(key string, defaultVal []string) []string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	items := strings.Split(val, ",")
	for i := range items {
		items[i] = strings.TrimSpace(items[i])
	}
	return items
}

func loadEnvFile(paths ...string) {
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, `"'`)
				if _, exists := os.LookupEnv(key); !exists {
					_ = os.Setenv(key, val)
				}
			}
		}
		break
	}
}
