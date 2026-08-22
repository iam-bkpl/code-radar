package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const repoOwner = "iam-bkpl"
const repoName = "code-radar"
const githubAPI = "https://api.github.com/repos/" + repoOwner + "/" + repoName

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Name    string        `json:"name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func getExecutablePath() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return exe
}

func getCurrentVersion() string {
	return version
}

func getLatestRelease() (*githubRelease, error) {
	resp, err := http.Get(githubAPI + "/releases/latest")
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("no releases found")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error (status %d)", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}
	return &release, nil
}

func getAllReleases() ([]githubRelease, error) {
	resp, err := http.Get(githubAPI + "/releases")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}
	return releases, nil
}

func getAssetName(version string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	archStr := goarch
	if goarch == "amd64" {
		archStr = "x86_64"
	} else if goarch == "386" {
		archStr = "i386"
	}

	if goos == "windows" {
		return fmt.Sprintf("%s_%s_%s.zip", repoName, capitalize(goos), archStr)
	}
	return fmt.Sprintf("%s_%s_%s.tar.gz", repoName, capitalize(goos), archStr)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func findAsset(release *githubRelease, assetName string) *githubAsset {
	for i, a := range release.Assets {
		if a.Name == assetName {
			return &release.Assets[i]
		}
	}
	return nil
}

func downloadAndInstall(version string) error {
	release, err := getReleaseByTag(version)
	if err != nil {
		return err
	}

	assetName := getAssetName(version)
	asset := findAsset(release, assetName)
	if asset == nil {
		available := make([]string, 0)
		for _, a := range release.Assets {
			if strings.Contains(a.Name, "checksums") {
				continue
			}
			available = append(available, a.Name)
		}
		return fmt.Errorf("no binary found for %s/%s\n  Available: %s", runtime.GOOS, runtime.GOARCH, strings.Join(available, ", "))
	}

	fmt.Printf("  Downloading %s...\n", asset.Name)
	tmpFile, err := downloadFile(asset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpFile)

	fmt.Printf("  Extracting...\n")
	binaryPath, err := extractBinary(tmpFile, asset.Name)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}
	defer os.Remove(binaryPath)

	exePath := getExecutablePath()
	if exePath == "" {
		return fmt.Errorf("cannot determine executable path")
	}

	fmt.Printf("  Installing to %s...\n", exePath)

	if err := copyFile(binaryPath, exePath); err != nil {
		// Try with sudo
		fmt.Printf("  Retrying with sudo...\n")
		cmd := exec.Command("sudo", "cp", binaryPath, exePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("install failed: %w\n  Try: sudo cp %s %s", err, binaryPath, exePath)
		}
	}

	os.Chmod(exePath, 0755)

	// macOS Gatekeeper fix
	if runtime.GOOS == "darwin" {
		fmt.Printf("  Fixing macOS permissions...\n")
		exec.Command("xattr", "-cr", exePath).Run()
	}

	return nil
}

func getReleaseByTag(tag string) (*githubRelease, error) {
	resp, err := http.Get(githubAPI + "/releases/tags/" + tag)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release %s: %w", tag, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("release %s not found", tag)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}
	return &release, nil
}

func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", "code-radar-*")
	if err != nil {
		return "", err
	}

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func extractBinary(archivePath, assetName string) (string, error) {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath)
	}
	return extractTarGz(archivePath)
}

func extractTarGz(archivePath string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		name := filepath.Base(header.Name)
		if name == repoName || name == repoName+".exe" {
			tmpBin, err := os.CreateTemp("", "code-radar-bin-*")
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(tmpBin, tr); err != nil {
				tmpBin.Close()
				return "", err
			}
			tmpBin.Close()
			os.Chmod(tmpBin.Name(), 0755)
			return tmpBin.Name(), nil
		}
	}
	return "", fmt.Errorf("binary not found in archive")
}

func extractZip(archivePath string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		name := filepath.Base(f.Name)
		if name == repoName || name == repoName+".exe" {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			tmpBin, err := os.CreateTemp("", "code-radar-bin-*")
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(tmpBin, rc); err != nil {
				tmpBin.Close()
				return "", err
			}
			tmpBin.Close()
			os.Chmod(tmpBin.Name(), 0755)
			return tmpBin.Name(), nil
		}
	}
	return "", fmt.Errorf("binary not found in archive")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func runUpgrade() {
	currentVersion := getCurrentVersion()
	if currentVersion == "dev" {
		fmt.Println()
		fmt.Println("  You're running a dev build. Install a release to enable upgrades.")
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Println("  Checking for latest release...")
	latest, err := getLatestRelease()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		os.Exit(1)
	}

	if latest.TagName == currentVersion {
		fmt.Printf("  Already on latest version (%s)\n", currentVersion)
		fmt.Println()
		return
	}

	fmt.Printf("  Current:  %s\n", currentVersion)
	fmt.Printf("  Latest:   %s\n", latest.TagName)
	fmt.Println()
	fmt.Println("  Upgrading...")
	fmt.Println()

	if err := downloadAndInstall(latest.TagName); err != nil {
		fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("  Upgraded to %s successfully!\n", latest.TagName)
	fmt.Println()
}

func runCheck() {
	currentVersion := getCurrentVersion()
	if currentVersion == "dev" {
		fmt.Println()
		fmt.Println("  Running dev build")
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Println("  Checking for latest release...")
	latest, err := getLatestRelease()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Current:  %s\n", currentVersion)
	fmt.Printf("  Latest:   %s\n", latest.TagName)
	if latest.TagName == currentVersion {
		fmt.Println("  Status:   Up to date")
	} else {
		fmt.Println("  Status:   Update available")
		fmt.Printf("\n  Run 'code-radar --upgrade' to upgrade\n")
	}
	fmt.Println()
}

func runVersions() {
	fmt.Println()
	fmt.Println("  Fetching releases...")
	fmt.Println()

	releases, err := getAllReleases()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		os.Exit(1)
	}

	currentVersion := getCurrentVersion()

	for _, r := range releases {
		marker := "  "
		if r.TagName == currentVersion {
			marker = "→ "
		}
		fmt.Printf("  %s%s\n", marker, r.TagName)
	}

	fmt.Printf("\n  → = current version\n")
	fmt.Printf("\n  Install a specific version: code-radar --install <version>\n")
	fmt.Println()
}

func runInstall(version string) {
	if version == "" {
		fmt.Fprintln(os.Stderr, "  Error: version is required")
		fmt.Fprintln(os.Stderr, "  Usage: code-radar --install <version>")
		fmt.Fprintln(os.Stderr, "  Example: code-radar --install v1.0.0")
		os.Exit(1)
	}

	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	fmt.Println()
	fmt.Printf("  Installing %s...\n", version)
	fmt.Println()

	if err := downloadAndInstall(version); err != nil {
		fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Installed %s successfully!\n", version)
	fmt.Println()
}
