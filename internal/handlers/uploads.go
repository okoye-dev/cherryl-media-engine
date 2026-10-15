package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/okoye-dev/cherryl-media-engine/internal/storage"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

const maxFilenameLength = 180

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

type PresignUploadRequest struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

type PresignUploadResponse struct {
	UploadURL  string `json:"upload_url"`
	StorageKey string `json:"storage_key"`
	PublicURL  string `json:"public_url"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	ExpiresIn  int    `json:"expires_in"`
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

	var req PresignUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Size <= 0 {
		http.Error(w, "name and size (>0) are required", http.StatusBadRequest)
		return
	}

	fileName := sanitizeFilename(req.Name)
	fileID := uuid.New().String()
	ext := path.Ext(fileName)
	storageKey := fmtKey(fileID, ext)

	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if h.storage.IsConfigured() {
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
			ID:         fileID,
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
		ID:         fileID,
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

	key := r.URL.Query().Get("key")
	if key == "" {
		key = "upload-" + uuid.New().String()
	}

	uploadDir := "./uploads"
	_ = os.MkdirAll(uploadDir, 0755)
	destPath := filepath.Join(uploadDir, filepath.Base(key))

	out, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, r.Body)
	if err != nil {
		http.Error(w, "Failed to write file bytes", http.StatusInternalServerError)
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

func fmtKey(id, ext string) string {
	now := time.Now().UTC()
	return path.Join("media", now.Format("2006/01"), id+ext)
}
