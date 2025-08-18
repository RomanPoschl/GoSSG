package main

import (
	"bytes"
	"fmt"
	"github.com/gomarkdown/markdown"
	"gopkg.in/yaml.v3"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Page represents a single content file with its metadata and HTML content.
type Page struct {
	FrontMatter map[string]interface{}
	Content     template.HTML
}

// buildProject processes the content directory and generates the static site.
func buildProject(project *Project) error {
	log.Printf("Building project '%s'...", project.Name)
	contentDir := filepath.Join(project.Path, "content")
	publicDir := filepath.Join(project.Path, "public")
	themeDir := filepath.Join(project.Path, "themes", "default")
	staticDir := filepath.Join(themeDir, "static")

	// --- 1. Load Templates ---
	// We'll start with a single template, but this can be expanded.
	templatesDir := filepath.Join(themeDir, "templates")
	pageTemplatePath := filepath.Join(templatesDir, "page.html")

	if _, err := os.Stat(pageTemplatePath); os.IsNotExist(err) {
		return fmt.Errorf("page template not found at %s", pageTemplatePath)
	}
	log.Printf("Loading template: %s", pageTemplatePath)
	tmpl, err := template.ParseFiles(pageTemplatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// --- 2. Clean the public directory ---
	log.Println("Cleaning public directory...")
	if err := os.RemoveAll(publicDir); err != nil {
		return fmt.Errorf("failed to remove public directory: %w", err)
	}
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		return fmt.Errorf("failed to create public directory: %w", err)
	}

	// --- 3. Walk the content directory and process files ---
	log.Println("Processing content...")
	err = filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(contentDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}
		destPath := filepath.Join(publicDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		if strings.HasSuffix(strings.ToLower(path), ".md") {
			log.Printf("  -> Processing %s\n", relPath)
			page, err := processMarkdownFile(path)
			if err != nil {
				return fmt.Errorf("failed to process markdown file %s: %w", path, err)
			}

			// Create the destination HTML file.
			newDestPath := strings.TrimSuffix(destPath, ".md") + ".html"
			file, err := os.Create(newDestPath)
			if err != nil {
				return fmt.Errorf("failed to create destination file %s: %w", newDestPath, err)
			}
			defer file.Close()

			// Execute the template with the page data.
			err = tmpl.Execute(file, page)
			if err != nil {
				return fmt.Errorf("failed to execute template for %s: %w", path, err)
			}
		} else {
			log.Printf("  -> Copying %s\n", relPath)
			return copyFile(path, destPath)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error during build process: %w", err)
	}

	// --- 4. Copy static assets from the theme directory ---
	log.Println("Copying static assets from theme...")
	// Check if the theme's static directory exists before trying to copy from it.
	if _, err := os.Stat(staticDir); !os.IsNotExist(err) {
		err = filepath.WalkDir(staticDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if path == staticDir {
				return nil
			}

			relPath, err := filepath.Rel(staticDir, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path for %s: %w", path, err)
			}
			destPath := filepath.Join(publicDir, relPath)

			if d.IsDir() {
				return os.MkdirAll(destPath, 0755)
			}

			log.Printf("  -> Copying static asset %s\n", relPath)
			return copyFile(path, destPath)
		})

		if err != nil {
			return fmt.Errorf("error during static asset copying: %w", err)
		}
	} else {
		log.Println("No static directory found in theme, skipping.")
	}
	log.Println("Build complete!")
	return nil
}

// processMarkdownFile reads a markdown file, parses its front matter,
// and converts the content to HTML.
func processMarkdownFile(path string) (*Page, error) {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	page := &Page{
		FrontMatter: make(map[string]interface{}),
	}

	// Split the file into front matter and markdown content.
	// The "---" separator is the key.
	parts := bytes.SplitN(fileContent, []byte("---"), 3)
	var content []byte
	if len(parts) >= 3 {
		// We have front matter.
		if err := yaml.Unmarshal(parts[1], &page.FrontMatter); err != nil {
			return nil, fmt.Errorf("failed to parse front matter: %w", err)
		}
		// The rest is the content.
		content = parts[2]
	} else {
		// No front matter found.
		content = fileContent
	}

	// Convert markdown to HTML.
	htmlContent := markdown.ToHTML(content, nil, nil)

	page.Content = template.HTML(htmlContent)
	return page, nil
}

// copyFile is a helper to copy a single file from src to dst.
func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", src, err)
	}
	return os.WriteFile(dst, content, 0644)
}
