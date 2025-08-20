package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
)

func startServer(port string, config Config) error {
	subFs, err := fs.Sub(adminUI, "admin")

	if err != nil {
		return err
	}

	fs := http.FS(subFs)
	fileServer := http.StripPrefix("/", http.FileServer(fs))

	http.Handle("/", fileServer)
	http.HandleFunc("/api/ui/projects", projectsHandler(config))

	log.Printf("Starting admin server on http://localhost:%s\n", port)

	return http.ListenAndServe(":"+port, nil)
}

func projectsHandler(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			renderProjectList(w, config)
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Failed to parse form", http.StatusBadRequest)
				return
			}
			projectName := r.FormValue("projectName")
			// Get the new project path from the form.
			projectPath := r.FormValue("projectPath")

			if projectName == "" {
				http.Error(w, "Project name cannot be empty", http.StatusBadRequest)
				return
			}

			// Pass both name and path to the addProject function.
			if err := config.addProject(projectName, projectPath); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Printf("Error adding project: %v", err)
				return
			}
			log.Printf("Created new project from UI: %s", projectName)

			renderProjectList(w, config)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// renderProjectList is a helper function to render the projects template.
func renderProjectList(w http.ResponseWriter, config Config) {
	templatePath := filepath.Join("templates", "projects.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Could not parse template", http.StatusInternalServerError)
		log.Printf("Error parsing projects template: %v", err)
		return
	}

	err = tmpl.Execute(w, config.Projects)
	if err != nil {
		http.Error(w, "Could not execute template", http.StatusInternalServerError)
		log.Printf("Error executing projects template: %v", err)
		return
	}
}
