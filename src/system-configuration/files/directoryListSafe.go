package main

import (
	"log"
	"net/http"
	"os"
)

type justFilesFilesystem struct {
	fs http.FileSystem
}

func (fs justFilesFilesystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return neuteredReaddirFile{f}, nil
}

type neuteredReaddirFile struct {
	http.File
}

func (f neuteredReaddirFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func main() {
	fs := justFilesFilesystem{http.Dir("tmp/static/")}
	err := http.ListenAndServe(":8080", http.StripPrefix("/tmp/static", http.FileServer(fs)))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
