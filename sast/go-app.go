package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

const maxBodyBytes = 1 << 20 // 1 MiB

var filesRoot = filepath.Clean("/var/www/files")

var greetTemplate = template.Must(template.New("greet").Parse(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Greet</title></head>
<body><h1>Hello, {{.Name}}</h1></body></html>`))

type greetData struct {
	Name string
}

type deserializeRequest struct {
	Message string `json:"message"`
	Count   int    `json:"count"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hash", hashHandler)
	mux.HandleFunc("/user", sqlInjectionHandler)
	mux.HandleFunc("/exec", safeFileReadHandler)
	mux.HandleFunc("/deserialize", deserializeHandler)
	mux.HandleFunc("/file", directoryTraversalHandler)
	mux.HandleFunc("/greet", xssHandler)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	// HTTPS only (cleartext ListenAndServe is intentionally not used). Example:
	//   openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 1 -nodes -subj "/CN=localhost"
	//   TLS_CERT=cert.pem TLS_KEY=key.pem DATABASE_DSN='user:pass@tcp(127.0.0.1:3306)/db' go run .
	certFile := os.Getenv("TLS_CERT")
	keyFile := os.Getenv("TLS_KEY")
	if certFile == "" || keyFile == "" {
		log.Fatal("TLS_CERT and TLS_KEY must be set to PEM file paths (e.g. from openssl). Cleartext HTTP is disabled.")
	}
	log.Printf("listening on https://%s", srv.Addr)
	if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

// hashHandler demonstrates password hashing with bcrypt (not MD5).
func hashHandler(w http.ResponseWriter, r *http.Request) {
	password := r.URL.Query().Get("password")
	if password == "" {
		http.Error(w, "missing password query parameter", http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "hash error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "bcrypt hash: %s", string(hash))
}

func sqlInjectionHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		http.Error(w, "DATABASE_DSN is not set", http.StatusServiceUnavailable)
		return
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	const q = `SELECT name FROM users WHERE id = ?`
	var name string
	if err := db.QueryRowContext(ctx, q, userID).Scan(&name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Hello, %s", html.EscapeString(name))
}

// safeFileReadHandler reads a file from filesRoot without invoking the shell.
func safeFileReadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	path, err := safePathUnderRoot(filesRoot, filename)
	if err != nil {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "File read error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(data)
}

func deserializeHandler(w http.ResponseWriter, r *http.Request) {
	body := http.MaxBytesReader(w, r.Body, maxBodyBytes)
	defer body.Close()

	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()

	var req deserializeRequest
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "JSON parse error", http.StatusBadRequest)
		return
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "trailing JSON not allowed", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": req.Message,
		"count":   req.Count,
	})
}

func directoryTraversalHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	path, err := safePathUnderRoot(filesRoot, filename)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "File read error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(data)
}

func xssHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := greetTemplate.Execute(w, greetData{Name: name}); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func safePathUnderRoot(root, userRel string) (string, error) {
	if userRel == "" || strings.Contains(userRel, "\x00") {
		return "", errors.New("invalid path")
	}
	cleanUser := filepath.Clean(userRel)
	if cleanUser == "." || filepath.IsAbs(cleanUser) {
		return "", errors.New("invalid path")
	}
	rootAbs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", err
	}
	fullAbs, err := filepath.Abs(filepath.Join(rootAbs, cleanUser))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(rootAbs, fullAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("path escapes root")
	}
	return fullAbs, nil
}
