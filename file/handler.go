package file

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

type FileRepository interface {
}

type Handler struct {
	// Repository FileRepository
}

func NewHandler() *Handler {
	return &Handler{}
}

var filesPath = "./files/"
var maxFileSizeMB = 512

func (h *Handler) uploadFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxFileSizeMB)<<20)
	r.ParseMultipartForm(int64(maxFileSizeMB) << 20)
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

	log.Printf("File written to disk { path: %s }", path)
}

func (h *Handler) getFile(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) SetupRoutes(mux *chi.Mux) {
	mux.Post("/", http.HandlerFunc(h.uploadFile))
	mux.Get("/", http.HandlerFunc(h.getFile))
}
