package main

import (
	"github.com/labstack/echo/v4"
	"html/template"
	"log"
	"my-ssg/core"
	"net/http"
	"path/filepath"
	"time"
)

func createWebServerMux(engine *core.Engine) http.Handler {
	mux := echo.New()

	mux.Static("/", "frontend/dist")

	mux.GET("/api/ui/projects-view", projectsHandler(engine))
	mux.GET("/api/ui/projects", listProjectsHandler(engine))
	mux.GET("/api/ui/project/:name", projectDashboardHandler(engine))
	mux.POST("/api/ui/project/:name/build", handleBuildProject(engine))

	log.Println("Internal web server mux created.")
	return mux
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
		return renderTemplate(c, "toast-success.html", data)
	}
}
