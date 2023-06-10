package file

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/xsni1/quick-bin/hasher"
)

type FileRepository interface {
	Insert(file File) error
	Get(id string) (*File, error)
	GetIfDownloadsLeft(id string) (*File, error)
}

type Handler struct {
	repository FileRepository
	logger     zerolog.Logger
}

func NewHandler(repository FileRepository, logger zerolog.Logger) *Handler {
	return &Handler{
		repository: repository,
		logger:     logger,
	}
}

type File struct {
	Name          string
	DownloadsLeft int
	Id            string
}

type uploadFileResponse struct {
	Id string
}

var filesPath = "./data/"
var maxFileSizeMB = 512

func getDownloads(downloads string) (int, error) {
	if downloads == "" {
		return -1, nil
	}

	return strconv.Atoi(downloads)
}

func (h *Handler) uploadFile(w http.ResponseWriter, r *http.Request) {
	h.logger.Trace().Msg("uploadFile method called")

	downloads, err := getDownloads(r.URL.Query().Get("downloads"))
	if err != nil {
		http.Error(w, "Incorrect parameter", http.StatusBadRequest)
		h.logger.Debug().AnErr("Invalid downloads parameter", err)
		return
	}

	// n Bytes * 2^20 = n Megabytes
	size := int64(maxFileSizeMB) << 20
	r.Body = http.MaxBytesReader(w, r.Body, size<<20)
	r.ParseMultipartForm(size << 20)
	file, header, err := r.FormFile("file")

	if errors.Is(err, http.ErrMissingFile) {
		http.Error(w, "No file provided", http.StatusBadRequest)
		h.logger.Debug().AnErr("No file provided", err)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Debug().AnErr("Failure during decoding form-file", err)
		return
	}

	defer file.Close()

	h.logger.Debug().
		Str("file name", header.Filename).
		Int64("size", header.Size).
		Msg("File received")

	id := hasher.Hasher(5)
	h.logger.Debug().
		Str("hash", id)

	path := fmt.Sprintf("%s%s", filesPath, id)
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Debug().AnErr("Failure during file read from request", err)
		return
	}

	err = os.WriteFile(path, fileData, 0644)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Debug().AnErr("Failure during file write", err)
		return
	}

	err = h.repository.Insert(File{
		Name:          header.Filename,
		DownloadsLeft: downloads,
		Id:            id,
	})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Debug().AnErr("Failure during db insertion operation", err)
		return
	}

	h.logger.Debug().Str("path", path).Msg("File written to disk")

	response, err := json.Marshal(uploadFileResponse{Id: id})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Debug().AnErr("JSON marshalling error", err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
	h.logger.Trace().Msg("uploadFile execution finished")
}

func (h *Handler) getFile(w http.ResponseWriter, r *http.Request) {
	fileId := chi.URLParam(r, "fileId")

	if fileId == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	file, err := h.repository.GetIfDownloadsLeft(fileId)

	if errors.Is(err, NoDownloadsLeftErr) {
		http.Error(w, "No downloads left", http.StatusForbidden)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Failure getting", err)
		return
	}

	path := fmt.Sprintf("%s%s", filesPath, file.Id)
	opened, err := os.Open(path)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Failure reading file from disk", err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	dat := bufio.NewReader(opened)

	dat.WriteTo(w)
}

func (h *Handler) log(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dumpedRequest, _ := httputil.DumpRequest(r, false)

		h.logger.Trace().
			Str("host", r.Host).
			Str("path", r.URL.Path).
			Bytes("request", dumpedRequest).
			Msg("Request start")
		next.ServeHTTP(w, r)
		h.logger.Trace().Msg("Request end")
	})
}

func (h *Handler) SetupRoutes(mux *chi.Mux) {
	mux.Post("/", h.log(h.uploadFile))
	mux.Get("/{fileId}", h.log(h.getFile))
}
