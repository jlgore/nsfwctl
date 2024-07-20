package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EnsureDirectory creates a directory if it doesn't exist
func EnsureDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// FileExists checks if a file exists and is not a directory
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// GetHomeDir returns the user's home directory
func GetHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting user home directory: %v", err)
	}
	return home, nil
}

// GetAppDir returns the application's directory (usually ~/.nsfwctl)
func GetAppDir() (string, error) {
	home, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".nsfwctl"), nil
}

// FormatDuration formats a duration in a human-readable format
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	parts := []string{}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
	}
	if m > 0 {
		parts = append(parts, fmt.Sprintf("%dm", m))
	}
	if s > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", s))
	}
	return strings.Join(parts, " ")
}

// TruncateString truncates a string to a specified length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ParseKeyValuePairs parses a string of key-value pairs (e.g., "key1=value1,key2=value2")
func ParseKeyValuePairs(s string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return result
}

// IsValidBranchName checks if a given string is a valid Git branch name
func IsValidBranchName(name string) bool {
	// Git branch naming rules:
	// - Must not contain spaces
	// - Must not contain two consecutive dots
	// - Must not contain ".."
	// - Must not end with "/"
	// - Must not contain "~", "^", ":", "?", "*", "[", "@{", "\"
	// - Must not start with "-"
	if strings.Contains(name, " ") ||
		strings.Contains(name, "..") ||
		strings.HasSuffix(name, "/") ||
		strings.ContainsAny(name, "~^:?*[\\@{") ||
		strings.HasPrefix(name, "-") {
		return false
	}
	return true
}
