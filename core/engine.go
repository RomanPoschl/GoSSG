package core

import (
	"log"
	"os"
	"path/filepath"
)

// Engine is the central struct that manages all core functionality.
// It holds the application's configuration and provides methods to interact with it.
// The UI layer (Wails) will hold an instance of this Engine.
type Engine struct {
	config *Config
}

// NewEngine creates and initializes a new Engine instance.
// It's responsible for finding the user's config directory and loading the
// projects.json file. This is the main entry point to our core logic.
func NewEngine() (*Engine, error) {
	// Find the platform-specific user config directory.
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Could not find user config directory: %v", err)
		return nil, err
	}

	// Define the path for our application's configuration file.
	configPath := filepath.Join(userConfigDir, "GoStaticCMS", "projects.json")

	// Load the configuration. The loadConfig function (which we will move
	// to project.go) will handle creating the file if it doesn't exist.
	config, err := loadConfig(configPath)
	if err != nil {
		log.Printf("Error loading configuration: %v", err)
		return nil, err
	}

	log.Println("Core engine initialized successfully.")

	// Return a new Engine instance containing the loaded config.
	return &Engine{
		config: config,
	}, nil
}

// NOTE: We will need to add methods to this Engine struct. For example:
// func (e *Engine) AddProject(name, path string) error { ... }
// func (e *Engine) GetProjects() []Project { ... }
// func (e *Engine) BuildProject(name string) error { ... }
