package provider

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
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

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// isNotFound returns true if err is a discordgo REST 404 response.
// discordgo returns a *discordgo.RESTError for non-2xx responses; we
// inspect the HTTP status code rather than matching on the error string
// (which is fragile and breaks if discordgo changes its format).
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	var restErr *discordgo.RESTError
	if errors.As(err, &restErr) && restErr.Response != nil {
		return restErr.Response.StatusCode == http.StatusNotFound
	}
	return false
}