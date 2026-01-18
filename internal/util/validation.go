package util

import (
	"fmt"
	"regexp"
	"strings"
)

// Bundle ID format: reverse domain notation (e.g., com.company.app)
// Must contain at least 2 segments separated by dots
// Each segment can contain alphanumeric, hyphen, and underscore
var bundleIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*\.[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// ValidateBundleID validates that a bundle ID follows Apple's conventions
func ValidateBundleID(bundleID string) error {
	if bundleID == "" {
		return fmt.Errorf("bundle ID cannot be empty")
	}

	if len(bundleID) > 255 {
		return fmt.Errorf("bundle ID too long (max 255 characters)")
	}

	// Check for consecutive dots
	if strings.Contains(bundleID, "..") {
		return fmt.Errorf("bundle ID cannot contain consecutive dots")
	}

	// Check format using regex
	if !bundleIDPattern.MatchString(bundleID) {
		return fmt.Errorf("invalid bundle ID format (expected reverse domain notation like com.company.app)")
	}

	return nil
}
