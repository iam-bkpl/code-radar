package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed web.html
var webFS embed.FS

func startWebServer(addr string) {
	cfg := loadConfig()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, _ := webFS.ReadFile("web.html")
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})

	http.HandleFunc("/api/scan", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Commit string `json:"commit"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Commit == "" {
			jsonError(w, "commit hash is required", http.StatusBadRequest)
			return
		}

		if err := verifyCommit(req.Commit); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}

		commitInfo, err := getCommitInfo(req.Commit)
		if err != nil {
			jsonError(w, "failed to get commit info", http.StatusInternalServerError)
			return
		}

		branches, err := findBranchesWithCommit(req.Commit)
		if err != nil {
			jsonError(w, "failed to find branches", http.StatusInternalServerError)
			return
		}

		parts := splitCommitInfo(commitInfo)
		result := scanResult{
			Hash:    parts[0],
			Message: safeIndex(parts, 1),
			Author:  safeIndex(parts, 2),
			Date:    safeIndex(parts, 3),
		}

		for _, b := range branches {
			env := matchEnvironment(b, cfg)
			result.Branches = append(result.Branches, branchInfo{
				Name: b,
				Env:  env,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(cfg)

		case http.MethodPost:
			var newCfg Config
			if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
				jsonError(w, "invalid config", http.StatusBadRequest)
				return
			}
			if len(newCfg.Environments) == 0 {
				jsonError(w, "environments cannot be empty", http.StatusBadRequest)
				return
			}

			home, err := os.UserHomeDir()
			if err != nil {
				jsonError(w, "cannot determine home directory", http.StatusInternalServerError)
				return
			}

			dir := filepath.Join(home, ".config", "code-radar")
			if err := os.MkdirAll(dir, 0755); err != nil {
				jsonError(w, "failed to create config directory", http.StatusInternalServerError)
				return
			}

			data, _ := json.MarshalIndent(newCfg, "", "  ")
			if err := os.WriteFile(filepath.Join(dir, "config.yaml"), data, 0644); err != nil {
				jsonError(w, "failed to save config", http.StatusInternalServerError)
				return
			}

			cfg = newCfg
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println()
	fmt.Println("  📡 Code Radar Web UI")
	fmt.Printf("  Listening on http://localhost%s\n\n", addr)
	fmt.Printf("  Open http://localhost%s in your browser\n\n", addr)
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println()

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "  Server error: %v\n", err)
		os.Exit(1)
	}
}

func splitCommitInfo(info string) []string {
	return strings.Split(info, "|")
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
