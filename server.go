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

	e.GET("/api/ui/editor/:name/new", showNewEditorHandler(engine))
	e.POST("/api/ui/save-article/:name", handleSaveArticleHandler(engine))
	e.GET("/api/ui/editor/:name/*", showEditorHandler(engine))

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
		return renderTemplate(c, filepath.Join("pages", "main-view.html"), nil)
	}
}

func listProjectsHandler(engine *core.Engine) echo.HandlerFunc {
	return func(c echo.Context) error {
		projects := engine.GetProjects()
		return renderTemplate(c, filepath.Join("partials", "projects-list.html"), projects)
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
		return renderTemplate(c, filepath.Join("pages", "project-dashboard.html"), data)
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
		return renderTemplate(c, filepath.Join("partials", "toast-success.html"), data)
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

		return renderTemplate(c, filepath.Join("partials", toastTemplate), data)
	}
}

func showNewEditorHandler(engine *core.Engine) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Render the editor with a blank, default Article object
		data := map[string]interface{}{
			"ProjectName": c.Param("name"),
			"Article":     &core.Article{ /* Default values here */ },
			"IsNew":       true,
		}
		return renderTemplate(c, filepath.Join("pages", "editor.html"), data)
	}
}

// showEditorHandler now uses our new parsing method
func showEditorHandler(engine *core.Engine) echo.HandlerFunc {
	return func(c echo.Context) error {
		projectName := c.Param("name")
		filePath := c.Param("*")
		article, err := engine.ParseArticleFile(projectName, filePath)

		if err != nil {
			return renderTemplate(c, filepath.Join("partials", "toast-error.html"), err.Error())
		}
		data := map[string]interface{}{
			"ProjectName": projectName,
			"Article":     article,
			"IsNew":       false,
		}
		return renderTemplate(c, filepath.Join("pages", "editor.html"), data)
	}
}

func handleSaveArticleHandler(engine *core.Engine) echo.HandlerFunc {
	return func(c echo.Context) error {
		projectName := c.Param("name")
		title := c.FormValue("title")
		body := c.FormValue("content")
		originalPath := c.FormValue("originalFilePath")

		articeData := &core.Article{
			FrontMatter: core.ArticleFrontMatter{
				Title: title,
				Date:  time.Now(),
			},
			Body: body,
		}

		finalPath, err := engine.SaveArticle(projectName, articeData, originalPath)

		if err != nil {
			// If the forge fails, send back an error toast.
			log.Printf("ERROR: Failed to save article for project '%s': %v", projectName, err)
			data := map[string]interface{}{
				"Timestamp": time.Now().UnixNano(),
				"Error":     err.Error(),
			}
			return renderTemplate(c, filepath.Join("partials", "toast-error.html"), data)
		}

		log.Printf("SUCCESS: Article saved to '%s'. Redirecting.", finalPath)
		editorURL := fmt.Sprintf("/api/ui/editor/%s/%s", projectName, finalPath)
		c.Response().Header().Set("HX-Redirect", editorURL)

		return c.NoContent(http.StatusOK)
	}
}
