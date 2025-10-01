package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"html/template"
	"log"
	"my-ssg/core"
	"net/http"
	"path/filepath"
	"time"
)

func createWebServerMux(engine *core.Engine) http.Handler {
	e := echo.New()

	e.Static("/", "frontend/dist")

	e.GET("/api/ui/projects-view", projectsHandler(engine))
	e.GET("/api/ui/projects", listProjectsHandler(engine))
	e.GET("/api/ui/project/:name", projectDashboardHandler(engine))
	e.POST("/api/ui/project/:name/build", handleBuildProject(engine))

	e.GET("/api/ui/editor/:name/*", showEditorHandler(engine))
	e.POST("/api/ui/editor/:name/*", saveFileHandler(engine))

	log.Println("Internal web server mux created.")
	return e
}

func renderTemplate(c echo.Context, name string, data interface{}) error {
	tmpl, err := template.ParseFiles(filepath.Join("templates", name))
	if err != nil {
		return c.String(http.StatusInternalServerError, "Template not found")
	}
	return tmpl.Execute(c.Response().Writer, data)
}

func projectsHandler(engine *core.Engine) echo.HandlerFunc {
	return func(c echo.Context) error {
		return renderTemplate(c, "main-view.html", nil)
	}
}

func listProjectsHandler(engine *core.Engine) echo.HandlerFunc {
	return func(c echo.Context) error {
		projects := engine.GetProjects()
		return renderTemplate(c, "projects.html", projects)
	}
}

func projectDashboardHandler(engine *core.Engine) echo.HandlerFunc {
	return func(c echo.Context) error {
		projectName := c.Param("name")

		project, err := engine.FindProjectByName(projectName)

		if err != nil {
			return c.String(http.StatusNotFound, err.Error())
		}

		files, err := engine.ListContentFiles(projectName)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		data := map[string]interface{}{"Project": project, "Files": files}
		return renderTemplate(c, "project-dashboard.html", data)
	}
}

func handleBuildProject(engine *core.Engine) echo.HandlerFunc {
	return func(c echo.Context) error {
		projectName := c.Param("name")

		err := engine.BuildProject(projectName)

		// Prepare data for the feedback template
		data := map[string]interface{}{
			"ProjectName": projectName,
			"Timestamp":   time.Now().UnixNano(), // Unique ID for the toast element
		}

		if err != nil {
			log.Printf("ERROR: Build failed for project '%s': %v", projectName, err)
			data["Error"] = err.Error()
			return renderTemplate(c, "toast-error.html", data)
		}

		log.Printf("SUCCESS: Project '%s' built successfully.", projectName)
		data["Message"] = fmt.Sprintf("Project '%s' built successfully.", projectName)
		return renderTemplate(c, "toast-success.html", data)
	}
}

func showEditorHandler(engine *core.Engine) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 1. Echo cleanly separates the named parameter and the wildcard.
		projectName := c.Param("name")
		// The wildcard parameter is accessed with "*".
		filePath := c.Param("*")

		// 2. Ask the core engine to read the file content.
		content, err := engine.ReadFileContent(projectName, filePath)
		if err != nil {
			// In a real app, you might render a nice error page here.
			return c.String(http.StatusNotFound, err.Error())
		}

		// 3. Prepare the data for the editor.html template.
		data := map[string]interface{}{
			"ProjectName": projectName,
			"FilePath":    filePath,
			"FileContent": content,
		}

		// 4. Render the editor template.
		return renderTemplate(c, "editor.html", data)
	}
}

func saveFileHandler(engine *core.Engine) echo.HandlerFunc {
	return func(c echo.Context) error {
		projectName := c.Param("name")
		filePath := c.Param("*")

		// The new content is in the form data, with the key "content"
		// which matches the name="" attribute of our <textarea>.
		newContent := c.FormValue("content")

		// Tell the engine to write the changes to the disk.
		err := engine.WriteFileContent(projectName, filePath, newContent)

		// Now, we'll render a toast notification to give the user feedback.
		var toastTemplate string
		data := map[string]interface{}{
			"Timestamp":   time.Now().UnixNano(),
			"ProjectName": projectName, // For a more descriptive message
		}

		if err != nil {
			toastTemplate = "toast-error.html"
			data["Error"] = err.Error()
			log.Printf("ERROR: Failed to save file '%s': %v", filePath, err)
		} else {
			toastTemplate = "toast-success.html"
			// You could add a more specific message if you want.
			data["Message"] = fmt.Sprintf("File '%s' saved successfully!", filePath)
			log.Printf("SUCCESS: File '%s' saved in project '%s'.", filePath, projectName)
		}

		return renderTemplate(c, toastTemplate, data)
	}
}
