package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

		parts := splitPipe(commitInfo)
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
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cfg)
			return
		}

		if r.Method == http.MethodPost {
			var newCfg Config
			if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
				jsonError(w, "invalid config", http.StatusBadRequest)
				return
			}
			if len(newCfg.Environments) == 0 {
				jsonError(w, "environments cannot be empty", http.StatusBadRequest)
				return
			}

			home, _ := os.UserHomeDir()
			dir := home + "/.config/code-radar"
			os.MkdirAll(dir, 0755)
			data, _ := json.MarshalIndent(newCfg, "", "  ")
			os.WriteFile(dir+"/config.yaml", data, 0644)

			cfg = newCfg
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	fmt.Println()
	fmt.Println("  📡 Code Radar Web UI")
	fmt.Printf("  Listening on http://localhost%s\n\n", addr)
	fmt.Printf("  Open http://localhost%s in your browser\n\n", addr)
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println()

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("  Server error: %v\n", err)
	}
}

func splitPipe(s string) []string {
	result := []string{}
	current := ""
	for _, c := range s {
		if c == '|' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	result = append(result, current)
	return result
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
