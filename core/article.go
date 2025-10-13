package core

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
	"time"
	"os"
	"log"
	"path/filepath"
	"regexp"
)

// ArticleFrontMatter defines the structure of our YAML front matter.
type ArticleFrontMatter struct {
	Title string    `yaml:"title"`
	Date  time.Time `yaml:"date"`
	// We can add more fields here later, like categories, tags, etc.
}

// Article represents a fully parsed markdown file.
type Article struct {
	FrontMatter ArticleFrontMatter
	Body        string
	FilePath    string // The relative path of the file
}

func (e *Engine) ParseArticleFile(projectName, filePath string) (*Article, error) {
	content, err := e.ReadFileContent(projectName, filePath)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid front matter format in %s", filePath)
	}

	article := &Article{FilePath: filePath}
	if err := yaml.Unmarshal([]byte(parts[1]), &article.FrontMatter); err != nil {
		return nil, fmt.Errorf("could not parse front matter for %s: %w", filePath, err)
	}

	article.Body = strings.TrimSpace(parts[2])
	return article, nil
}

func (e *Engine) SaveArticle(projectName string, articleData *Article, originalFilePath string) (string, error) {
    newSlug := slugify(articleData.FrontMatter.Title)
    if newSlug == "" {
        return "", fmt.Errorf("article title cannot be empty or invalid")
    }

    var finalPath string
    if originalFilePath == "" {
        fileName := fmt.Sprintf("%s.md", newSlug)
        finalPath = filepath.Join("posts", fileName)
    } else {
        originalSlug := strings.TrimSuffix(filepath.Base(originalFilePath), ".md")
        if newSlug != originalSlug {
            newFileName := fmt.Sprintf("%s.md", newSlug)
            finalPath = filepath.Join(filepath.Dir(originalFilePath), newFileName)
        } else {
            finalPath = originalFilePath
        }
    }
    articleData.FilePath = finalPath

    if err := e.WriteArticleFile(projectName, articleData); err != nil {
        return "", err
    }

    if originalFilePath != "" && originalFilePath != finalPath {
        log.Printf("Renaming article, deleting old file: %s", originalFilePath)
        oldFullPath := filepath.Join(e.config.Projects[0].Path, "content", originalFilePath) // Simplified for example
        os.Remove(oldFullPath)
    }

    return finalPath, nil
}
func (e *Engine) WriteArticleFile(projectName string, article *Article) error {
	frontMatterBytes, err := yaml.Marshal(&article.FrontMatter)
	if err != nil {
		return fmt.Errorf("could not marshal front matter: %w", err)
	}

	var contentBuilder bytes.Buffer
	contentBuilder.WriteString("---\n")
	contentBuilder.Write(frontMatterBytes)
	contentBuilder.WriteString("---\n\n")
	contentBuilder.WriteString(article.Body)

	return e.WriteFileContent(projectName, article.FilePath, contentBuilder.String())
}

func slugify(title string) string {
    slug := strings.ToLower(title)
    reg := regexp.MustCompile("[^a-z0-9-]+")
    slug = reg.ReplaceAllString(slug, "-")
    slug = strings.Trim(slug, "-")
    return slug
}
