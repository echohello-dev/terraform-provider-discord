package provider

import (
	"fmt"
	"strconv"
)

func parsePermissions(s string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("permissions string is empty")
	}
	p, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid permissions value %q: %w", s, err)
	}
	return p, nil
}

func formatPermissions(p int64) string {
	return fmt.Sprintf("%d", p)
}

func int64Ptr(i int64) *int64 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}
