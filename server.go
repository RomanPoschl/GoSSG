package main

import (
	"io/fs"
	"log"
	"net/http"
)

func startServer(port string) error {
	subFs, err := fs.Sub(adminUI, "admin")

	if err != nil {
		return err
	}

	fs := http.FS(subFs)
	fileServer := http.StripPrefix("/", http.FileServer(fs))
	http.Handle("/", fileServer)

	log.Printf("Starting admin server on http://localhost:%s\n", port)

	return http.ListenAndServe(":"+port, nil)
}
