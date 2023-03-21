package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	addr := ":6666"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		r.ParseMultipartForm(19999)
		form := r.MultipartForm

		fmt.Println(form.Value)
		fmt.Println(form.File)

		file, ok := form.File["obraz"]

		if !ok {
			http.Error(w, "No file", http.StatusBadRequest)
			return
		}

		f, err := file[len(file)-1].Open()

		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}

		fmt.Println(f)

		newFile, err := os.Create("plik")

		io.Copy(newFile, f)

		// reader, err := r.MultipartReader()
		// if err != nil {
		// 	log.Printf("Error: %s", err)
		// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
		// 	return
		// }

		// for {
		// 	part, err := reader.NextPart()
		// 	if err == io.EOF {
		// 		break
		// 	}
		// }

		// NewReader

		log.Println("req")
	})

	log.Printf("Server on up %s", addr)
	http.ListenAndServe(":6666", nil)
}
