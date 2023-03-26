package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var filesPath = "./files/"
var addr = ":6666"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 512<<20)
		r.ParseMultipartForm(512 << 20)
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
	})

	log.Printf("Server on up %s", addr)
	http.ListenAndServe(":6666", nil)
}
