package version

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName    string               `json:"tag_name"`
	Name       string               `json:"name"`
	HTMLURL    string               `json:"html_url"`
	Prerelease bool                 `json:"prerelease"`
	Draft      bool                 `json:"draft"`
	Assets     []GitHubReleaseAsset `json:"assets"`
}

// GitHubReleaseAsset is a downloadable file attached to a release
type GitHubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	Available   bool
	CurrentVer  string
	LatestVer   string
	ReleaseURL  string
	ReleaseName string
}

// parseVersion extracts version numbers from a version string (e.g., "v1.2.3" -> [1, 2, 3])
func parseVersion(version string) []int {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	parts := strings.Split(version, ".")
	nums := make([]int, len(parts))

	for i, part := range parts {
		// Remove any non-numeric suffixes (e.g., "1-beta", "2-rc1")
		fields := strings.FieldsFunc(part, func(r rune) bool {
			return r == '-' || r == '+' || r == '_'
		})
		// A component made only of separators (e.g. "1..2") has no fields,
		// so leave it at zero instead of indexing an empty slice.
		if len(fields) == 0 {
			continue
		}

		if num, err := strconv.Atoi(fields[0]); err == nil {
			nums[i] = num
		}
	}

	return nums
}

// compareVersions compares two version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	nums1 := parseVersion(v1)
	nums2 := parseVersion(v2)

	// Pad with zeros to make lengths equal
	maxLen := len(nums1)
	if len(nums2) > maxLen {
		maxLen = len(nums2)
	}

	for len(nums1) < maxLen {
		nums1 = append(nums1, 0)
	}
	for len(nums2) < maxLen {
		nums2 = append(nums2, 0)
	}

	// Compare each part
	for i := 0; i < maxLen; i++ {
		if nums1[i] < nums2[i] {
			return -1
		}
		if nums1[i] > nums2[i] {
			return 1
		}
	}

	return 0
}

// Compare reports how two version strings relate: -1 if a < b, 0 if equal,
// 1 if a > b. Suffixes such as "-beta" or "-rc1" are ignored.
func Compare(a, b string) int {
	return compareVersions(a, b)
}

// FetchRelease returns a GitHub release. When tag is empty the latest stable
// release is fetched, otherwise the release with the exact tag is looked up.
func FetchRelease(ctx context.Context, tag, userAgent string) (*GitHubRelease, error) {
	url := "https://api.github.com/repos/Nicozilabe/baqueano/releases/latest"
	if tag != "" {
		url = "https://api.github.com/repos/Nicozilabe/baqueano/releases/tags/" + tag
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if userAgent == "" {
		userAgent = "baq"
	}
	req.Header.Set("User-Agent", userAgent)
	if token := githubToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		if tag == "" {
			return nil, errors.New("no releases found")
		}
		return nil, fmt.Errorf("release %q not found", tag)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &release, nil
}

// FetchReleases returns the most recent releases, newest first.
func FetchReleases(ctx context.Context, userAgent string) ([]GitHubRelease, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.github.com/repos/Nicozilabe/baqueano/releases?per_page=15", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if userAgent == "" {
		userAgent = "baq"
	}
	req.Header.Set("User-Agent", userAgent)
	if token := githubToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return releases, nil
}

// githubToken returns an optional GitHub token from the environment, used to
// lift the unauthenticated API rate limit when available.
func githubToken() string {
	for _, key := range []string{"BAQUEANO_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

// CheckForUpdates checks GitHub for the latest release of baq
func CheckForUpdates(ctx context.Context, currentVersion string) (*UpdateInfo, error) {
	// Skip version check if current version is "dev"
	if currentVersion == "dev" {
		return &UpdateInfo{
			Available:  false,
			CurrentVer: currentVersion,
		}, nil
	}

	release, err := FetchRelease(ctx, "", "baq/"+currentVersion)
	if err != nil {
		return nil, err
	}

	// Skip pre-releases and drafts
	if release.Prerelease || release.Draft {
		return &UpdateInfo{
			Available:  false,
			CurrentVer: currentVersion,
		}, nil
	}

	// Compare versions
	updateAvailable := compareVersions(currentVersion, release.TagName) < 0

	return &UpdateInfo{
		Available:   updateAvailable,
		CurrentVer:  currentVersion,
		LatestVer:   release.TagName,
		ReleaseURL:  release.HTMLURL,
		ReleaseName: release.Name,
	}, nil
}
