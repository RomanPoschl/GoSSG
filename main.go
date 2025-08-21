package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// --- Data Structures ---

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

// --- Core Functions ---
// findProjectByName searches for a project in the config by its name.
func (c *Config) findProjectByName(name string) (*Project, error) {
	for _, p := range c.Projects {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("project '%s' not found", name)
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
func (c *Config) save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	// Ensure the directory exists before writing the file.
	dir := filepath.Dir(c.configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(c.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// addProject adds a new project to the config and creates its directory structure.
func (c *Config) addProject(name, customPath string) error {
	for _, p := range c.Projects {
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

	c.Projects = append(c.Projects, Project{Name: name, Path: projectPath})
	return c.save()
}

// --- Main Application Logic ---

func main() {
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "GoSSG",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}

	// --- 1. Setup: Locate and load the configuration file ---
	// We'll store projects.json in a standard user config location.
	// userConfigDir, err := os.UserConfigDir()
	// if err != nil {
	// 	log.Fatalf("Could not find user config directory: %v", err)
	// }
	// configPath := filepath.Join(userConfigDir, "GoStaticCMS", "projects.json")
	// configPath := "projects.json"

	// config, err := loadConfig(configPath)
	// if err != nil {
	// 	log.Fatalf("Error loading configuration: %v", err)
	// }

	// // --- 2. Handle Command-Line Arguments ---
	// // We use flags to tell our app what to do.
	// addProjectName := flag.String("add", "", "The name of a new project to create.")
	// listProjects := flag.Bool("list", false, "List all managed projects.")
	// buildProjectName := flag.String("build", "", "The name of a project to build")
	// serve := flag.Bool("serve", false, "Start the admin web server.")
	// flag.Parse()

	// // --- 3. Execute Actions ---
	// if *addProjectName != "" {
	// 	if err := config.addProject(*addProjectName, ""); err != nil {
	// 		log.Fatalf("Error adding project: %v", err)
	// 	}
	// 	log.Printf("Project '%s' created successfully!", *addProjectName)
	// 	return
	// }

	// if *listProjects {
	// 	if len(config.Projects) == 0 {
	// 		fmt.Println("No projects found. Use the --add flag to create one.")
	// 		return
	// 	}
	// 	fmt.Println("Managed Projects:")
	// 	for _, p := range config.Projects {
	// 		fmt.Printf("  - %s (%s)\n", p.Name, p.Path)
	// 	}
	// 	return
	// }

	// if *buildProjectName != "" {
	// 	project, err := config.findProjectByName(*buildProjectName)
	// 	if err != nil {
	// 		log.Fatalf("Error: %v", err)
	// 	}
	// 	if err := buildProject(project); err != nil {
	// 		log.Fatalf("Build failed: %v", err)
	// 	}
	// 	return
	// }

	// if *serve {
	// 	// Call the function from our new server.go file
	// 	if err := startServer("8080", *config); err != nil {
	// 		log.Fatalf("Failed to start server: %v", err)
	// 	}
	// 	return
	// }

	// fmt.Println("No action specified. Use --add 'Project Name' or --list.")
}
