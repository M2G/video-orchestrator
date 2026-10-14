package interfaces

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"video-orchestrator/internal/domain"

	"github.com/sirupsen/logrus"
)

type HTTPServer struct {
	repo       domain.Repository
	uploadDir  string
	streamsDir string
	log        *logrus.Logger
	server     *http.Server
}

func NewHTTPServer(repo domain.Repository, uploadDir string, streamsDir string, log *logrus.Logger) *HTTPServer {
	h := &HTTPServer{
		repo:       repo,
		uploadDir:  uploadDir,
		streamsDir: streamsDir,
		log:        log,
	}

	http.HandleFunc("GET /test", h.test)
	http.HandleFunc("POST /jobs", h.handleUpload)
	http.HandleFunc("GET /jobs/{id}", h.handleGetJob)
	http.HandleFunc("GET /streams/video/{id}/{filename}", h.handleServeFile)

	h.server = &http.Server{}

	return h
}

func (h *HTTPServer) test(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
}

func (h *HTTPServer) Start(addr string) error {
	h.server.Addr = addr
	h.log.WithField("addr", addr).Info("http_server_started")
	return h.server.ListenAndServe()
}

func (h *HTTPServer) Stop(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func (h *HTTPServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid content-type")
		return
	}

	boundary := params["boundary"]
	if boundary == "" {
		h.sendError(w, http.StatusBadRequest, "missing boundary")
		return
	}

	mr := multipart.NewReader(r.Body, boundary)
	part, err := mr.NextPart()
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid multipart")
		return
	}
	defer part.Close()

	originalFilename := part.FileName()
	if !strings.HasSuffix(originalFilename, ".mp4") {
		h.sendError(w, http.StatusBadRequest, "only MP4 files are accepted")
		return
	}

	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to create upload dir")
		return
	}

	// écriture dans un fichier temporaire : le préfixe "." et l'extension
	// .tmp-... empêchent video_segmenter de le prendre pendant l'upload
	tempPath := filepath.Join(h.uploadDir, fmt.Sprintf(".tmp-%d.part", time.Now().UnixNano()))
	f, err := os.Create(tempPath)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to create file")
		return
	}

	const maxSize = 2 << 30 // 2GB
	if _, err := io.Copy(f, io.LimitReader(part, maxSize)); err != nil {
		f.Close()
		os.Remove(tempPath)
		h.sendError(w, http.StatusInternalServerError, "failed to write file")
		return
	}
	if err := f.Close(); err != nil {
		os.Remove(tempPath)
		h.sendError(w, http.StatusInternalServerError, "failed to write file")
		return
	}

	// le job donne l'ID dont on a besoin pour nommer le fichier final
	job, err := h.repo.CreateJob(r.Context(), originalFilename)
	if err != nil {
		os.Remove(tempPath)
		h.sendError(w, http.StatusInternalServerError, "failed to create job")
		return
	}

	// video_segmenter produira alors streams/{id}/{id}.m3u8
	finalPath := filepath.Join(h.uploadDir, fmt.Sprintf("%d.mp4", job.ID))
	if err := os.Rename(tempPath, finalPath); err != nil {
		os.Remove(tempPath)
		h.sendError(w, http.StatusInternalServerError, "failed to finalize upload")
		return
	}

	h.log.WithField("job_id", job.ID).Info("job_created")
	h.sendJSON(w, http.StatusCreated, map[string]any{
		"job_id":   job.ID,
		"filename": job.Filename,
	})
}

func (h *HTTPServer) handleGetJob(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	job, err := h.repo.GetJobByID(r.Context(), id)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "job not found")
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]any{
		"job_id":   job.ID,
		"filename": job.Filename,
		"status":   job.Status,
	})
}

func (h *HTTPServer) handleServeFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")

	ext := filepath.Ext(filename)
	switch ext {
	case ".m3u8":
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	case ".ts":
		w.Header().Set("Content-Type", "video/mp2t")
	default:
		h.sendError(w, http.StatusForbidden, "file type not allowed")
		return
	}

	filePath := filepath.Join(h.streamsDir, id, filename)
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	http.ServeFile(w, r, filePath)
}

func (h *HTTPServer) sendJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func (h *HTTPServer) sendError(w http.ResponseWriter, status int, msg string) {
	h.sendJSON(w, status, map[string]string{"error": msg})
}
