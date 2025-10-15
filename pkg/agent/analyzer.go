package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectInfo contains information about the current project
type ProjectInfo struct {
	RootPath       string            `json:"root_path"`
	Language       string            `json:"language"`
	Framework      string            `json:"framework,omitempty"`
	PackageManager string            `json:"package_manager,omitempty"`
	BuildTool      string            `json:"build_tool,omitempty"`
	TestFramework  string            `json:"test_framework,omitempty"`
	Dependencies   []string          `json:"dependencies,omitempty"`
	Scripts        map[string]string `json:"scripts,omitempty"`
	Files          []string          `json:"files,omitempty"`
}

// ProjectAnalyzer analyzes project structure and context
type ProjectAnalyzer struct {
	rootPath string
}

// NewProjectAnalyzer creates a new project analyzer
func NewProjectAnalyzer(rootPath string) *ProjectAnalyzer {
	return &ProjectAnalyzer{
		rootPath: rootPath,
	}
}

// AnalyzeProject analyzes the current project structure
func (pa *ProjectAnalyzer) AnalyzeProject() (*ProjectInfo, error) {
	info := &ProjectInfo{
		RootPath: pa.rootPath,
		Scripts:  make(map[string]string),
	}

	// Detect project language and tools
	if err := pa.detectLanguageAndTools(info); err != nil {
		return nil, fmt.Errorf("failed to detect project language: %w", err)
	}

	// Scan project files
	if err := pa.scanProjectFiles(info); err != nil {
		return nil, fmt.Errorf("failed to scan project files: %w", err)
	}

	return info, nil
}

// detectLanguageAndTools detects the primary language and tools used in the project
func (pa *ProjectAnalyzer) detectLanguageAndTools(info *ProjectInfo) error {
	// Check for Go project
	if pa.fileExists("go.mod") {
		info.Language = "Go"
		info.BuildTool = "go"
		info.PackageManager = "go"
		return pa.analyzeGoProject(info)
	}

	// Check for Node.js project
	if pa.fileExists("package.json") {
		info.Language = "JavaScript"
		info.PackageManager = "npm"
		if pa.fileExists("yarn.lock") {
			info.PackageManager = "yarn"
		} else if pa.fileExists("pnpm-lock.yaml") {
			info.PackageManager = "pnpm"
		}
		return pa.analyzeNodeProject(info)
	}

	// Check for Python project
	if pa.fileExists("requirements.txt") || pa.fileExists("pyproject.toml") || pa.fileExists("setup.py") {
		info.Language = "Python"
		info.PackageManager = "pip"
		if pa.fileExists("pyproject.toml") {
			info.BuildTool = "poetry"
		}
		return pa.analyzePythonProject(info)
	}

	// Check for Rust project
	if pa.fileExists("Cargo.toml") {
		info.Language = "Rust"
		info.BuildTool = "cargo"
		info.PackageManager = "cargo"
		return pa.analyzeRustProject(info)
	}

	// Default to unknown
	info.Language = "Unknown"
	return nil
}

// analyzeGoProject analyzes Go-specific project details
func (pa *ProjectAnalyzer) analyzeGoProject(info *ProjectInfo) error {
	// Read go.mod for dependencies
	goModPath := filepath.Join(pa.rootPath, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	inRequireBlock := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}
		
		if inRequireBlock && line == ")" {
			inRequireBlock = false
			continue
		}
		
		if inRequireBlock || strings.HasPrefix(line, "require ") {
			// Parse dependency
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				dep := parts[0]
				if strings.HasPrefix(dep, "require") && len(parts) >= 3 {
					dep = parts[1]
				}
				if !strings.Contains(dep, "//") {
					info.Dependencies = append(info.Dependencies, dep)
				}
			}
		}
	}

	// Check for common Go testing frameworks
	if pa.containsImport("github.com/stretchr/testify") {
		info.TestFramework = "testify"
	}

	return nil
}

// analyzeNodeProject analyzes Node.js-specific project details
func (pa *ProjectAnalyzer) analyzeNodeProject(info *ProjectInfo) error {
	packagePath := filepath.Join(pa.rootPath, "package.json")
	content, err := os.ReadFile(packagePath)
	if err != nil {
		return err
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Scripts         map[string]string `json:"scripts"`
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		return err
	}

	// Extract dependencies
	for dep := range pkg.Dependencies {
		info.Dependencies = append(info.Dependencies, dep)
	}
	for dep := range pkg.DevDependencies {
		info.Dependencies = append(info.Dependencies, dep)
	}

	// Extract scripts
	info.Scripts = pkg.Scripts

	// Detect framework
	if _, exists := pkg.Dependencies["react"]; exists {
		info.Framework = "React"
	} else if _, exists := pkg.Dependencies["vue"]; exists {
		info.Framework = "Vue"
	} else if _, exists := pkg.Dependencies["angular"]; exists {
		info.Framework = "Angular"
	} else if _, exists := pkg.Dependencies["express"]; exists {
		info.Framework = "Express"
	}

	// Detect test framework
	if _, exists := pkg.DevDependencies["jest"]; exists {
		info.TestFramework = "Jest"
	} else if _, exists := pkg.DevDependencies["mocha"]; exists {
		info.TestFramework = "Mocha"
	}

	return nil
}

// analyzePythonProject analyzes Python-specific project details
func (pa *ProjectAnalyzer) analyzePythonProject(info *ProjectInfo) error {
	// Check for requirements.txt
	if pa.fileExists("requirements.txt") {
		content, err := os.ReadFile(filepath.Join(pa.rootPath, "requirements.txt"))
		if err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					// Extract package name (before version specifier)
					parts := strings.FieldsFunc(line, func(r rune) bool {
						return r == '=' || r == '>' || r == '<' || r == '!'
					})
					if len(parts) > 0 {
						info.Dependencies = append(info.Dependencies, parts[0])
					}
				}
			}
		}
	}

	// Check for common test frameworks
	if pa.containsDependency(info.Dependencies, "pytest") {
		info.TestFramework = "pytest"
	} else if pa.containsDependency(info.Dependencies, "unittest") {
		info.TestFramework = "unittest"
	}

	return nil
}

// analyzeRustProject analyzes Rust-specific project details
func (pa *ProjectAnalyzer) analyzeRustProject(info *ProjectInfo) error {
	cargoPath := filepath.Join(pa.rootPath, "Cargo.toml")
	content, err := os.ReadFile(cargoPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	inDepsSection := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "[dependencies]") {
			inDepsSection = true
			continue
		}
		
		if strings.HasPrefix(line, "[") && inDepsSection {
			inDepsSection = false
			continue
		}
		
		if inDepsSection && strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				dep := strings.TrimSpace(parts[0])
				info.Dependencies = append(info.Dependencies, dep)
			}
		}
	}

	return nil
}

// scanProjectFiles scans and lists important project files
func (pa *ProjectAnalyzer) scanProjectFiles(projectInfo *ProjectInfo) error {
	return filepath.Walk(pa.rootPath, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if there's an error
		}

		// Skip hidden directories and common ignore patterns
		if fileInfo.IsDir() {
			name := fileInfo.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "target" {
				return filepath.SkipDir
			}
			return nil
		}

		// Add relevant files
		relPath, err := filepath.Rel(pa.rootPath, path)
		if err != nil {
			return nil
		}

		// Include source files, config files, and documentation
		ext := strings.ToLower(filepath.Ext(relPath))
		name := strings.ToLower(fileInfo.Name())
		
		if isRelevantFile(ext, name) {
			projectInfo.Files = append(projectInfo.Files, relPath)
		}

		return nil
	})
}

// isRelevantFile determines if a file is relevant for project analysis
func isRelevantFile(ext, name string) bool {
	relevantExts := []string{
		".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".rs", ".java", ".c", ".cpp", ".h", ".hpp",
		".json", ".yaml", ".yml", ".toml", ".xml", ".md", ".txt", ".cfg", ".conf", ".ini",
	}

	relevantNames := []string{
		"readme", "license", "dockerfile", "makefile", "gitignore", "gitattributes",
	}

	for _, relevantExt := range relevantExts {
		if ext == relevantExt {
			return true
		}
	}

	for _, relevantName := range relevantNames {
		if strings.Contains(name, relevantName) {
			return true
		}
	}

	return false
}

// Helper functions
func (pa *ProjectAnalyzer) fileExists(filename string) bool {
	_, err := os.Stat(filepath.Join(pa.rootPath, filename))
	return !os.IsNotExist(err)
}

func (pa *ProjectAnalyzer) containsImport(importPath string) bool {
	// This is a simplified check - in practice, you'd parse Go files
	return filepath.Walk(pa.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") {
			return nil
		}
		
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		
		if strings.Contains(string(content), importPath) {
			return filepath.SkipAll // Found it, stop walking
		}
		
		return nil
	}) == filepath.SkipAll
}

func (pa *ProjectAnalyzer) containsDependency(deps []string, dep string) bool {
	for _, d := range deps {
		if strings.Contains(strings.ToLower(d), strings.ToLower(dep)) {
			return true
		}
	}
	return false
}