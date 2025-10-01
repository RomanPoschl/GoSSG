package core

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v3"
)

// Page holds the data for a single rendered page.
type Page struct {
	FrontMatter map[string]interface{}
	Content     template.HTML
}

// BuildProject is the main method for generating the static site for a given project.
func (e *Engine) BuildProject(projectName string) error {
	project, err := e.FindProjectByName(projectName)
	if err != nil {
		return err // Project not found
	}

	log.Printf("Starting build for project: %s", project.Name)

	// Define key paths
	contentDir := filepath.Join(project.Path, "content")
	publicDir := filepath.Join(project.Path, "public")
	themeDir := filepath.Join(project.Path, "themes", "default") // Assuming 'default' theme for now
	templatePath := filepath.Join(themeDir, "templates", "page.html")

	// 1. Clean the public directory
	log.Println("Cleaning public directory...")
	if err := os.RemoveAll(publicDir); err != nil {
		return fmt.Errorf("failed to clean public directory: %w", err)
	}
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		return fmt.Errorf("failed to recreate public directory: %w", err)
	}

	// 2. Parse the main page template once
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("could not parse page template '%s': %w", templatePath, err)
	}

	// 3. Process content files
	log.Println("Processing content files...")
	err = filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate the relative path to preserve directory structure
		relPath, err := filepath.Rel(contentDir, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(publicDir, relPath)

		// Ensure the destination directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Process Markdown files
		if strings.HasSuffix(info.Name(), ".md") {
			destPath = strings.TrimSuffix(destPath, ".md") + ".html"
			return processMarkdownFile(path, destPath, tmpl)
		} else {
			// Copy other files directly
			return copyFile(path, destPath)
		}
	})
	if err != nil {
		return fmt.Errorf("error walking content directory: %w", err)
	}

	// 4. Copy theme static assets
	log.Println("Copying static assets...")
	staticDir := filepath.Join(themeDir, "static")
	return copyStaticAssets(staticDir, publicDir)
}

// processMarkdownFile reads, parses, and renders a single markdown file using the provided template.
func processMarkdownFile(sourcePath, destPath string, tmpl *template.Template) error {
	log.Printf("Processing markdown file: %s", sourcePath)
	fileData, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", sourcePath, err)
	}

	parts := strings.SplitN(string(fileData), "---", 3)
	if len(parts) < 3 {
		return fmt.Errorf("invalid front matter in file %s", sourcePath)
	}

	page := Page{
		FrontMatter: make(map[string]interface{}),
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &page.FrontMatter); err != nil {
		return fmt.Errorf("failed to parse front matter in %s: %w", sourcePath, err)
	}

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	htmlContent := markdown.ToHTML([]byte(parts[2]), p, nil)
	page.Content = template.HTML(htmlContent)

	outputFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", destPath, err)
	}
	defer outputFile.Close()

	return tmpl.Execute(outputFile, page)
}

// copyStaticAssets recursively copies files from a source to a destination directory.
func copyStaticAssets(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}
		return copyFile(path, destPath)
	})
}

// copyFile is a simple utility to copy a single file.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func (e *Engine) ListContentFiles(projectName string) ([]string, error) {
	project, err := e.FindProjectByName(projectName)
	if err != nil {
		// If the project doesn't exist, we can't list its files.
		return nil, err
	}

	contentDir := filepath.Join(project.Path, "content")
	var files []string

	err = filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// We only want to list files, not the directories themselves.
		if !info.IsDir() {
			relPath, err := filepath.Rel(contentDir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
