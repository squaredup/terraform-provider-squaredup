package provider

import "testing"

func TestDetermineBaseURL(t *testing.T) {
	tests := []struct {
		region string
		want   string
	}{
		{region: "us", want: "https://api.squaredup.com"},
		{region: "eu", want: "https://eu.api.squaredup.com"},
		{region: "au", want: "https://au.api.squaredup.com"},
		{region: "https://example.squaredup.com/", want: "https://example.squaredup.com"},
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			got, err := determineBaseURL(tt.region)
			if err != nil {
				t.Fatalf("determineBaseURL(%q) returned an error: %v", tt.region, err)
			}
			if got != tt.want {
				t.Errorf("determineBaseURL(%q) = %q, want %q", tt.region, got, tt.want)
			}
		})
	}
}

func TestDetermineBaseURLRejectsUnknownRegion(t *testing.T) {
	for _, region := range []string{"uk", "AU", "http://au.api.squaredup.com"} {
		if _, err := determineBaseURL(region); err == nil {
			t.Errorf("determineBaseURL(%q) returned no error", region)
		}
	}
}
