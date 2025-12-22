package system

import (
	"runtime"
	"time"

	"QuanPhotos/internal/config"
	"QuanPhotos/internal/pkg/database"
)

// Version info (can be set via ldflags at build time)
var (
	Version   = "1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

// HealthStatus represents the health check result
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// SystemInfo represents system information
type SystemInfo struct {
	App     AppInfo     `json:"app"`
	Runtime RuntimeInfo `json:"runtime"`
	Storage StorageInfo `json:"storage"`
	I18n    I18nInfo    `json:"i18n"`
}

// AppInfo represents application info
type AppInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Env       string `json:"env"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
}

// RuntimeInfo represents runtime info
type RuntimeInfo struct {
	GoVersion  string `json:"go_version"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	CPUs       int    `json:"cpus"`
	Goroutines int    `json:"goroutines"`
	Uptime     string `json:"uptime"`
}

// StorageInfo represents storage info
type StorageInfo struct {
	SupportedFormats []string `json:"supported_formats"`
	MaxUploadSize    int64    `json:"max_upload_size"`
}

// I18nInfo represents i18n info
type I18nInfo struct {
	SupportedLanguages []string `json:"supported_languages"`
	DefaultLanguage    string   `json:"default_language"`
}

// Service provides system-related operations
type Service struct {
	config *config.Config
}

// NewService creates a new system service
func NewService(cfg *config.Config) *Service {
	return &Service{
		config: cfg,
	}
}

// GetHealth returns the health status of the system
func (s *Service) GetHealth() (*HealthStatus, bool) {
	checks := make(map[string]string)

	// Check database connection
	dbStatus := "ok"
	db := database.GetDB()
	if db == nil {
		dbStatus = "disconnected"
	} else if err := db.Ping(); err != nil {
		dbStatus = "unhealthy"
	}
	checks["database"] = dbStatus

	// Determine overall status
	status := "ok"
	healthy := true
	if dbStatus != "ok" {
		status = "degraded"
		healthy = false
	}

	return &HealthStatus{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}, healthy
}

// GetSystemInfo returns system information
func (s *Service) GetSystemInfo() *SystemInfo {
	return &SystemInfo{
		App: AppInfo{
			Name:      s.config.App.Name,
			Version:   Version,
			Env:       s.config.App.Env,
			BuildTime: BuildTime,
			GitCommit: GitCommit,
		},
		Runtime: RuntimeInfo{
			GoVersion:  runtime.Version(),
			OS:         runtime.GOOS,
			Arch:       runtime.GOARCH,
			CPUs:       runtime.NumCPU(),
			Goroutines: runtime.NumGoroutine(),
			Uptime:     time.Since(startTime).String(),
		},
		Storage: StorageInfo{
			SupportedFormats: s.config.Storage.AllowedTypes,
			MaxUploadSize:    s.config.Storage.MaxSize,
		},
		I18n: I18nInfo{
			SupportedLanguages: []string{"zh-CN", "en-US"},
			DefaultLanguage:    "zh-CN",
		},
	}
}
