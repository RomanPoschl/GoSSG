package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReadFileContent finds a project and a specific file within its content directory,
// and returns the content of that file as a string.
func (e *Engine) ReadFileContent(projectName, filePath string) (string, error) {
	// 1. Find the project to ensure we're working in the right context.
	project, err := e.FindProjectByName(projectName)
	if err != nil {
		return "", err
	}

	// 2. Construct the full, absolute path to the target file.
	// filepath.Join is used for security and cross-platform compatibility.
	// It prevents path traversal attacks (e.g., ../../some_other_file).
	fullPath := filepath.Join(project.Path, "content", filePath)

	// 3. Read the file from the disk.
	content, err := os.ReadFile(fullPath)
	if err != nil {
		// This will handle cases where the file doesn't exist or we don't have permission.
		return "", fmt.Errorf("could not read file '%s': %w", filePath, err)
	}

	// 4. Return the content as a string.
	return string(content), nil
}

// WriteFileContent finds a project and writes new content to a specific file
// within its content directory.
func (e *Engine) WriteFileContent(projectName, filePath, newContent string) error {
	// 1. Find the project.
	project, err := e.FindProjectByName(projectName)
	if err != nil {
		return err
	}

	// 2. Construct the full, absolute path.
	fullPath := filepath.Join(project.Path, "content", filePath)

	// 3. Ensure the directory for the file exists before writing.
	// For example, if the path is "posts/new-post.md", this creates the "posts" directory.
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create directory for file '%s': %w", filePath, err)
	}

	// 4. Write the new content to the file, overwriting it if it exists.
	// 0644 is a standard file permission.
	if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("could not write to file '%s': %w", filePath, err)
	}

	return nil
}
