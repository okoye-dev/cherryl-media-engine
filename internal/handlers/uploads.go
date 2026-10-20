package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/okoye-dev/cherryl-media-engine/internal/storage"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

const (
	maxFilenameLength  = 180
	MaxUploadSizeBytes = 100 * 1024 * 1024 // 100 MB max file size for editorial media
	MaxJsonBytes       = 2 * 1024 * 1024   // 2 MB max for json metadata payloads
)

var allowedMediaMIMEs = map[string]bool{
	"image/jpeg":       true,
	"image/jpg":        true,
	"image/png":        true,
	"image/webp":       true,
	"image/gif":        true,
	"image/avif":       true,
	"video/mp4":        true,
	"video/webm":       true,
	"video/quicktime":  true,
	"audio/mpeg":       true,
	"audio/mp3":        true,
	"audio/mp4":        true,
	"audio/m4a":        true,
	"audio/wav":        true,
	"audio/x-wav":      true,
	"audio/ogg":        true,
	"application/pdf":  true,
}

var dangerousExtensions = map[string]bool{
	".exe":   true,
	".sh":    true,
	".bash":  true,
	".bat":   true,
	".cmd":   true,
	".php":   true,
	".phtml": true,
	".js":    true,
	".mjs":   true,
	".html":  true,
	".htm":   true,
	".svg":   true, // Disallow raw SVG to prevent Stored XSS via inline scripts
	".cgi":   true,
	".pl":    true,
	".py":    true,
	".vbs":   true,
}

func isAllowedMedia(mimeType, filename string) bool {
	ext := strings.ToLower(path.Ext(filename))
	if dangerousExtensions[ext] {
		return false
	}
	baseMime := strings.ToLower(strings.TrimSpace(strings.Split(mimeType, ";")[0]))
	if allowedMediaMIMEs[baseMime] {
		return true
	}
	// Also allow if safe extension matches known safe formats even if MIME was octet-stream
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".avif", ".mp4", ".webm", ".mov", ".mp3", ".m4a", ".wav", ".ogg", ".pdf":
		return true
	}
	return false
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = path.Base(name)

	cleaned := strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, name)

	cleaned = strings.TrimSpace(cleaned)
	if runes := []rune(cleaned); len(runes) > maxFilenameLength {
		ext := path.Ext(cleaned)
		if len([]rune(ext)) > 20 {
			ext = ""
		}
		keep := maxFilenameLength - len([]rune(ext))
		cleaned = string(runes[:keep]) + ext
	}

	if cleaned == "" || cleaned == "." || cleaned == ".." {
		return "upload"
	}
	return cleaned
}

// MediaNamespace is the fixed UUID v5 namespace for Cherryl Media assets.
// It ensures that deterministic UUIDs are consistently generated across all uploads and migrations.
var MediaNamespace = uuid.MustParse("c7e841aa-8521-4f10-91a5-8314e3650226")

// GenerateDeterministicKey generates an idempotent, deterministic UUID v5
// so file keys are not dependent on where they are stored, enabling reproducible migrations.
func GenerateDeterministicKey(fileName, contentType string, size int64, checksum, sourceID, customKey string) (uuid.UUID, string) {
	if customKey != "" {
		cleanKey := strings.TrimPrefix(customKey, "/")
		base := strings.TrimSuffix(path.Base(cleanKey), path.Ext(cleanKey))
		if u, err := uuid.Parse(base); err == nil {
			return u, cleanKey
		}
	}

	var input string
	if checksum != "" {
		input = fmt.Sprintf("checksum:%s", checksum)
	} else if sourceID != "" {
		input = fmt.Sprintf("source:%s", sourceID)
	} else {
		cleanName := strings.ToLower(sanitizeFilename(fileName))
		input = fmt.Sprintf("file:%s:%d:%s", cleanName, size, contentType)
	}

	id := uuid.NewSHA1(MediaNamespace, []byte(input))
	ext := strings.ToLower(path.Ext(fileName))
	if ext == "" {
		switch contentType {
		case "image/webp":
			ext = ".webp"
		case "image/jpeg", "image/jpg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "video/mp4":
			ext = ".mp4"
		default:
			ext = ".bin"
		}
	}
	storageKey := fmt.Sprintf("media/%s%s", id.String(), ext)
	return id, storageKey
}

type PresignUploadRequest struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	Checksum    string `json:"checksum,omitempty"`   // Optional SHA-256 or content hash
	SourceID    string `json:"source_id,omitempty"`  // Optional migration identifier (e.g. WP attachment ID or legacy path)
	CustomKey   string `json:"custom_key,omitempty"` // Optional explicit deterministic key
}

type PresignUploadResponse struct {
	UploadURL     string `json:"upload_url"`
	StorageKey    string `json:"storage_key"`
	PublicURL     string `json:"public_url"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	ExpiresIn     int    `json:"expires_in"`
	AlreadyExists bool   `json:"already_exists,omitempty"`
}

type UploadHandler struct {
	storage storage.Storage
	store   store.UploadStore
}

func NewUploadHandler(s storage.Storage, st store.UploadStore) *UploadHandler {
	return &UploadHandler{
		storage: s,
		store:   st,
	}
}

func (h *UploadHandler) PresignUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxJsonBytes)
	var req PresignUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload or body too large", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Size <= 0 {
		http.Error(w, "name and size (>0) are required", http.StatusBadRequest)
		return
	}

	if req.Size > MaxUploadSizeBytes {
		http.Error(w, fmt.Sprintf("File exceeds maximum allowed size of %d MB", MaxUploadSizeBytes/(1024*1024)), http.StatusBadRequest)
		return
	}

	if !isAllowedMedia(req.ContentType, req.Name) {
		http.Error(w, "Unsupported or disallowed media format. Allowed formats: JPEG, PNG, WebP, GIF, AVIF, MP4, WebM, MP3, PDF.", http.StatusBadRequest)
		return
	}

	fileName := sanitizeFilename(req.Name)
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileID, storageKey := GenerateDeterministicKey(fileName, contentType, req.Size, req.Checksum, req.SourceID, req.CustomKey)

	if h.storage.IsConfigured() {
		// Idempotency check: see if this exact file already landed in S3
		if existingSize, err := h.storage.GetFileSize(r.Context(), storageKey); err == nil && existingSize > 0 {
			publicURL := h.storage.GetPublicURL(storageKey)
			if h.store != nil {
				_ = h.store.ConfirmUpload(r.Context(), storageKey, existingSize, publicURL)
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(PresignUploadResponse{
				UploadURL:     "",
				StorageKey:    storageKey,
				PublicURL:     publicURL,
				ID:            fileID.String(),
				Name:          fileName,
				ExpiresIn:     0,
				AlreadyExists: true,
			})
			return
		}

		uploadURL, err := h.storage.GetPresignedUploadURL(r.Context(), storageKey, req.Size, contentType)
		if err != nil {
			log.Printf("presign error: %v", err)
			http.Error(w, "Failed to create presigned upload URL", http.StatusInternalServerError)
			return
		}

		publicURL := h.storage.GetPublicURL(storageKey)
		if h.store != nil {
			_ = h.store.RecordUpload(r.Context(), storageKey, fileName, req.Size, contentType, publicURL)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PresignUploadResponse{
			UploadURL:  uploadURL,
			StorageKey: storageKey,
			PublicURL:  publicURL,
			ID:         fileID.String(),
			Name:       fileName,
			ExpiresIn:  int(storage.UploadPresignTTL.Seconds()),
		})
		return
	}

	// Local dev fallback when S3 is not yet configured: provide backend direct upload route
	localUploadURL := "/api/uploads/direct?key=" + storageKey
	publicURL := "/uploads/" + storageKey

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(PresignUploadResponse{
		UploadURL:  localUploadURL,
		StorageKey: storageKey,
		PublicURL:  publicURL,
		ID:         fileID.String(),
		Name:       fileName,
		ExpiresIn:  int(storage.UploadPresignTTL.Seconds()),
	})
}

// DirectUpload is a local dev fallback handler for uploading when S3 is not yet connected
func (h *UploadHandler) DirectUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSizeBytes)

	key := r.URL.Query().Get("key")
	if key == "" {
		key = "upload-" + uuid.New().String()
	}

	baseKey := filepath.Base(key)
	ext := strings.ToLower(path.Ext(baseKey))
	if dangerousExtensions[ext] {
		http.Error(w, "Disallowed file type", http.StatusBadRequest)
		return
	}

	uploadDir := "./uploads"
	_ = os.MkdirAll(uploadDir, 0755)
	destPath := filepath.Join(uploadDir, baseKey)

	out, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, r.Body)
	if err != nil {
		http.Error(w, "Failed to write file bytes or size limit exceeded", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

type ConfirmUploadRequest struct {
	StorageKey string `json:"storage_key"`
	FileName   string `json:"file_name"`
}

type ConfirmUploadResponse struct {
	Confirmed  bool   `json:"confirmed"`
	StorageKey string `json:"storage_key"`
	PublicURL  string `json:"public_url"`
	SizeBytes  int64  `json:"size_bytes"`
}

func (h *UploadHandler) ConfirmUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxJsonBytes)
	var req ConfirmUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StorageKey == "" {
		http.Error(w, `{"error":"storage_key is required"}`, http.StatusBadRequest)
		return
	}

	var size int64
	var err error

	if h.storage.IsConfigured() {
		size, err = h.storage.GetFileSize(r.Context(), req.StorageKey)
		if err != nil {
			log.Printf("confirm upload check %s: %v", req.StorageKey, err)
			http.Error(w, `{"error":"File not found in S3 bucket. Ensure upload finished before confirming."}`, http.StatusNotFound)
			return
		}
	} else {
		// local dev check
		destPath := filepath.Join("./uploads", filepath.Base(req.StorageKey))
		info, statErr := os.Stat(destPath)
		if statErr != nil {
			http.Error(w, `{"error":"File not found in local upload directory"}`, http.StatusNotFound)
			return
		}
		size = info.Size()
	}

	publicURL := h.storage.GetPublicURL(req.StorageKey)
	if !h.storage.IsConfigured() {
		publicURL = "/uploads/" + filepath.Base(req.StorageKey)
	}

	if h.store != nil {
		_ = h.store.ConfirmUpload(r.Context(), req.StorageKey, size, publicURL)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ConfirmUploadResponse{
		Confirmed:  true,
		StorageKey: req.StorageKey,
		PublicURL:  publicURL,
		SizeBytes:  size,
	})
}

func CleanKey(key string) string {
	return strings.TrimPrefix(key, "/")
}

