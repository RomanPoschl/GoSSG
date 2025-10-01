package core

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Project holds the metadata for a single website project.
type Project struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Config holds the list of all projects managed by our application.
type Config struct {
	Projects   []Project `json:"projects"`
	configFile string    // The path to the config file, not saved in the JSON itself.
}

// loadConfig reads the projects.json file from the user's config directory.
// If the file doesn't exist, it returns an empty Config struct.
func loadConfig(path string) (*Config, error) {
	config := &Config{
		configFile: path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// If the file doesn't exist, that's okay. We'll create it on save.
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// save writes the current list of projects back to the projects.json file.
func (e *Engine) saveConfig() error {
	data, err := json.MarshalIndent(e.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	// Ensure the directory exists before writing the file.
	dir := filepath.Dir(e.config.configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(e.config.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// addProject adds a new project to the config and creates its directory structure.
func (e *Engine) AddProject(name, customPath string) error {
	for _, p := range e.config.Projects {
		if p.Name == name {
			return fmt.Errorf("project with name '%s' already exists", name)
		}
	}

	var projectRoot string
	if customPath == "" {
		// If no path is provided, create the project in the current directory.
		projectRoot = "."
	} else {
		projectRoot = customPath
	}

	// The final path is the root combined with the project name.
	projectPath, err := filepath.Abs(filepath.Join(projectRoot, name))
	if err != nil {
		return fmt.Errorf("could not determine absolute path for project: %w", err)
	}

	log.Printf("Creating new project '%s' at: %s\n", name, projectPath)

	dirs := []string{
		"content", "public", "themes/default/templates/partials",
		"themes/default/static/css", "themes/default/static/js", "addons",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}
	}

	e.config.Projects = append(e.config.Projects, Project{Name: name, Path: projectPath})
	return e.saveConfig()
}

// GetProjects returns the current list of projects.
func (e *Engine) GetProjects() []Project {
	return e.config.Projects
}

// findProjectByName searches for a project in the config by its name.
func (e *Engine) FindProjectByName(name string) (*Project, error) {
	for _, p := range e.config.Projects {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("project '%s' not found", name)
}
