package file

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type FileRepository interface {
	Insert(file File) error
	Get(id string) (*File, error)
}

type Handler struct {
	repository FileRepository
}

func NewHandler(repository FileRepository) *Handler {
	return &Handler{
		repository: repository,
	}
}

type File struct {
	Name string
	Id   string
}

var filesPath = "./data/"
var maxFileSizeMB = 512

func (h *Handler) uploadFile(w http.ResponseWriter, r *http.Request) {
	// n Bytes * 2^20 = n Megabytes
	size := int64(maxFileSizeMB) << 20
	r.Body = http.MaxBytesReader(w, r.Body, size<<20)
	r.ParseMultipartForm(size << 20)
	file, header, err := r.FormFile("file")

	defer file.Close()

	if errors.Is(err, http.ErrMissingFile) {
		http.Error(w, "No file provided", http.StatusBadRequest)
		log.Println("No file provided: ", err)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Failure during file read: ", err)
		return
	}

	log.Printf("File received { name: %s, size: %d, header: %s }", header.Filename, header.Size, header.Header)

	path := fmt.Sprintf("%s%s", filesPath, header.Filename)
	fileData, err := io.ReadAll(file)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Failured during file read: ", err)
		return
	}

	err = os.WriteFile(path, fileData, 0644)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Failure during file write: ", err)
		return
	}

	// TODO: Add hash
	id := uuid.New()

	err = h.repository.Insert(File{
		Name: header.Filename,
		Id:   id.String(),
	})

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Failure during db insertion: ", err)
		return
	}

	log.Printf("File written to disk { path: %s }", path)
}

func (h *Handler) getFile(w http.ResponseWriter, r *http.Request) {
	fileId := chi.URLParam(r, "fileId")

	if fileId == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	file, err := h.repository.Get(fileId)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Failure getting", err)
		return
	}

	path := fmt.Sprintf("%s%s", filesPath, file.Name)
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

func (h *Handler) SetupRoutes(mux *chi.Mux) {
	mux.Post("/", http.HandlerFunc(h.uploadFile))
	mux.Get("/{fileId}", http.HandlerFunc(h.getFile))
}
