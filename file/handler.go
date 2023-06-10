package file

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/xsni1/quick-bin/hasher"
)

type FileRepository interface {
	Insert(file File) error
	Get(id string) (*File, error)
	GetIfDownloadsLeft(id string) (*File, error)
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
	log.Println("File upload started")

	downloads, err := getDownloads(r.URL.Query().Get("downloads"))

	if err != nil {
		http.Error(w, "Incorrect parameter", http.StatusBadRequest)
		log.Println("Failure during parameter conversion: ", err)
		return
	}

	// n Bytes * 2^20 = n Megabytes
	size := int64(maxFileSizeMB) << 20
	r.Body = http.MaxBytesReader(w, r.Body, size<<20)
	r.ParseMultipartForm(size << 20)
	file, header, err := r.FormFile("file")

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

	defer file.Close()

	log.Printf("File received { name: %s, size: %d, header: %s }", header.Filename, header.Size, header.Header)

	id := hasher.Hasher(5)
	log.Println("Generated hash: ", id)

	path := fmt.Sprintf("%s%s", filesPath, id)
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

	err = h.repository.Insert(File{
		Name:          header.Filename,
		DownloadsLeft: downloads,
		Id:            id,
	})

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Failure during db insertion: ", err)
		return
	}

	log.Printf("File written to disk { path: %s }", path)

	response, err := json.Marshal(uploadFileResponse{Id: id})

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("JSON Marshalling error: ", err)
		return
	}

	log.Println("Upload file response", response)
	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
}

func (h *Handler) getFile(w http.ResponseWriter, r *http.Request) {
	fileId := chi.URLParam(r, "fileId")

	if fileId == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	file, err := h.repository.GetIfDownloadsLeft(fileId)

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

func (h *Handler) SetupRoutes(mux *chi.Mux) {
	mux.Post("/", http.HandlerFunc(h.uploadFile))
	mux.Get("/{fileId}", http.HandlerFunc(h.getFile))
}
