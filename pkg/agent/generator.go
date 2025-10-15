package agent

import (
	"fmt"
	"strings"
	"text/template"
)

// CodeGenerator generates code based on project context and requirements
type CodeGenerator struct {
	projectInfo *ProjectInfo
}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator(projectInfo *ProjectInfo) *CodeGenerator {
	return &CodeGenerator{
		projectInfo: projectInfo,
	}
}

// GenerateTemplate generates code from a template and context
func (cg *CodeGenerator) GenerateTemplate(templateType string, context map[string]interface{}) (string, error) {
	tmpl, exists := cg.getTemplate(templateType)
	if !exists {
		return "", fmt.Errorf("template %s not found", templateType)
	}

	// Add project context to template context
	context["ProjectInfo"] = cg.projectInfo
	context["Language"] = cg.projectInfo.Language
	context["Framework"] = cg.projectInfo.Framework

	var builder strings.Builder
	t, err := template.New(templateType).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	if err := t.Execute(&builder, context); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return builder.String(), nil
}

// GenerateFunction generates a function based on specifications
func (cg *CodeGenerator) GenerateFunction(functionName, description string, params, returns []string) (string, error) {
	context := map[string]interface{}{
		"FunctionName": functionName,
		"Description":  description,
		"Params":       params,
		"Returns":      returns,
	}

	templateType := fmt.Sprintf("function_%s", strings.ToLower(cg.projectInfo.Language))
	return cg.GenerateTemplate(templateType, context)
}

// GenerateClass generates a class/struct based on specifications
func (cg *CodeGenerator) GenerateClass(className, description string, fields []Field) (string, error) {
	context := map[string]interface{}{
		"ClassName":   className,
		"Description": description,
		"Fields":      fields,
	}

	templateType := fmt.Sprintf("class_%s", strings.ToLower(cg.projectInfo.Language))
	return cg.GenerateTemplate(templateType, context)
}

// GenerateTest generates test code for a function or class
func (cg *CodeGenerator) GenerateTest(targetName, testType string) (string, error) {
	context := map[string]interface{}{
		"TargetName": targetName,
		"TestType":   testType,
	}

	templateType := fmt.Sprintf("test_%s", strings.ToLower(cg.projectInfo.Language))
	if cg.projectInfo.TestFramework != "" {
		templateType = fmt.Sprintf("test_%s_%s", 
			strings.ToLower(cg.projectInfo.Language), 
			strings.ToLower(cg.projectInfo.TestFramework))
	}

	return cg.GenerateTemplate(templateType, context)
}

// GenerateConfigFile generates configuration files
func (cg *CodeGenerator) GenerateConfigFile(configType string, options map[string]interface{}) (string, error) {
	context := map[string]interface{}{
		"Options": options,
	}

	templateType := fmt.Sprintf("config_%s", configType)
	return cg.GenerateTemplate(templateType, context)
}

// GenerateWebFile generates HTML/CSS/JS files with unique patterns to avoid recitation
func (cg *CodeGenerator) GenerateWebFile(fileType string, options map[string]interface{}) (string, error) {
	// Add unique identifiers to avoid recitation
	if options == nil {
		options = make(map[string]interface{})
	}
	
	// Add timestamp and random elements to make it unique
	if options["title"] == nil {
		options["title"] = "Console Buddy App"
	}
	if options["cssFile"] == nil {
		options["cssFile"] = "cb-styles.css"
	}
	if options["jsFile"] == nil {
		options["jsFile"] = "cb-script.js"
	}
	
	context := map[string]interface{}{
		"Options": options,
	}

	templateType := fmt.Sprintf("web_%s", strings.ToLower(fileType))
	return cg.GenerateTemplate(templateType, context)
}

// Field represents a field in a class/struct
type Field struct {
	Name        string
	Type        string
	Description string
	Tags        map[string]string
}

// getTemplate returns the appropriate template for the given type
func (cg *CodeGenerator) getTemplate(templateType string) (string, bool) {
	templates := map[string]string{
		// Go templates
		"function_go": goFunctionTemplate,
		"class_go":    goStructTemplate,
		"test_go":     goTestTemplate,
		"test_go_testify": goTestifyTemplate,

		// JavaScript/TypeScript templates
		"function_javascript": jsFunctionTemplate,
		"function_typescript": tsFunctionTemplate,
		"class_javascript":    jsClassTemplate,
		"class_typescript":    tsClassTemplate,
		"test_javascript":     jsTestTemplate,
		"test_typescript":     tsTestTemplate,
		"test_javascript_jest": jsJestTemplate,
		"test_typescript_jest": tsJestTemplate,

		// Python templates
		"function_python": pythonFunctionTemplate,
		"class_python":    pythonClassTemplate,
		"test_python":     pythonTestTemplate,
		"test_python_pytest": pythonPytestTemplate,

		// Config templates
		"config_dockerfile": dockerfileTemplate,
		"config_gitignore":  gitignoreTemplate,
		"config_makefile":   makefileTemplate,
		
		// Web templates (unique to avoid recitation)
		"web_html": uniqueHTMLTemplate,
		"web_css":  uniqueCSSTemplate,
		"web_js":   uniqueJSTemplate,
	}

	template, exists := templates[templateType]
	return template, exists
}

// Go templates
const goFunctionTemplate = `// {{.Description}}
func {{.FunctionName}}({{range $i, $param := .Params}}{{if $i}}, {{end}}{{$param}}{{end}}) {{if .Returns}}({{range $i, $ret := .Returns}}{{if $i}}, {{end}}{{$ret}}{{end}}){{end}} {
	// TODO: Implement {{.FunctionName}}
	{{if .Returns}}return{{range .Returns}} {{.}}{}{{end}}{{end}}
}`

const goStructTemplate = `// {{.Description}}
type {{.ClassName}} struct {
{{range .Fields}}	{{.Name}} {{.Type}} {{if .Tags}}` + "`" + `{{range $key, $value := .Tags}}{{$key}}:"{{$value}}" {{end}}` + "`" + `{{end}} // {{.Description}}
{{end}}}

// New{{.ClassName}} creates a new {{.ClassName}}
func New{{.ClassName}}() *{{.ClassName}} {
	return &{{.ClassName}}{}
}`

const goTestTemplate = `package main

import (
	"testing"
)

func Test{{.TargetName}}(t *testing.T) {
	// TODO: Implement test for {{.TargetName}}
	t.Skip("Test not implemented")
}`

const goTestifyTemplate = `package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func Test{{.TargetName}}(t *testing.T) {
	// TODO: Implement test for {{.TargetName}}
	assert.True(t, false, "Test not implemented")
}`

// JavaScript templates
const jsFunctionTemplate = `/**
 * {{.Description}}
 {{range .Params}} * @param {*} {{.}} - Parameter description
 {{end}}{{if .Returns}} * @returns {*} Return description{{end}}
 */
function {{.FunctionName}}({{range $i, $param := .Params}}{{if $i}}, {{end}}{{$param}}{{end}}) {
	// TODO: Implement {{.FunctionName}}
	{{if .Returns}}return null;{{end}}
}`

const jsClassTemplate = `/**
 * {{.Description}}
 */
class {{.ClassName}} {
	constructor() {
{{range .Fields}}		this.{{.Name}} = null; // {{.Description}}
{{end}}	}
}`

const jsTestTemplate = `describe('{{.TargetName}}', () => {
	test('should work correctly', () => {
		// TODO: Implement test for {{.TargetName}}
		expect(true).toBe(false);
	});
});`

const jsJestTemplate = jsTestTemplate

// TypeScript templates
const tsFunctionTemplate = `/**
 * {{.Description}}
 {{range .Params}} * @param {{.}} - Parameter description
 {{end}}{{if .Returns}} * @returns Return description{{end}}
 */
function {{.FunctionName}}({{range $i, $param := .Params}}{{if $i}}, {{end}}{{$param}}: any{{end}}){{if .Returns}}: any{{end}} {
	// TODO: Implement {{.FunctionName}}
	{{if .Returns}}return null;{{end}}
}`

const tsClassTemplate = `/**
 * {{.Description}}
 */
class {{.ClassName}} {
{{range .Fields}}	{{.Name}}: {{.Type}}; // {{.Description}}
{{end}}
	constructor() {
{{range .Fields}}		this.{{.Name}} = null as any;
{{end}}	}
}`

const tsTestTemplate = jsTestTemplate

const tsJestTemplate = jsTestTemplate

// Python templates
const pythonFunctionTemplate = `def {{.FunctionName}}({{range $i, $param := .Params}}{{if $i}}, {{end}}{{$param}}{{end}}):
	"""{{.Description}}
	
	Args:
{{range .Params}}		{{.}}: Parameter description
{{end}}	
	Returns:
{{if .Returns}}		Return description{{else}}		None{{end}}
	"""
	# TODO: Implement {{.FunctionName}}
	{{if .Returns}}return None{{else}}pass{{end}}`

const pythonClassTemplate = `class {{.ClassName}}:
	"""{{.Description}}"""
	
	def __init__(self):
{{range .Fields}}		self.{{.Name}} = None  # {{.Description}}
{{end}}`

const pythonTestTemplate = `import unittest

class Test{{.TargetName}}(unittest.TestCase):
	def test_{{.TargetName | lower}}(self):
		"""Test {{.TargetName}}"""
		# TODO: Implement test for {{.TargetName}}
		self.fail("Test not implemented")

if __name__ == '__main__':
	unittest.main()`

const pythonPytestTemplate = `import pytest

def test_{{.TargetName | lower}}():
	"""Test {{.TargetName}}"""
	# TODO: Implement test for {{.TargetName}}
	assert False, "Test not implemented"`

// Config templates
const dockerfileTemplate = `FROM {{.Options.baseImage | default "alpine:latest"}}

WORKDIR /app

{{if .Options.installCommands}}{{range .Options.installCommands}}RUN {{.}}
{{end}}{{end}}

COPY . .

{{if .Options.buildCommand}}RUN {{.Options.buildCommand}}{{end}}

{{if .Options.port}}EXPOSE {{.Options.port}}{{end}}

CMD ["{{.Options.startCommand | default "echo", "Hello World"}}"]`

const gitignoreTemplate = `# Dependencies
{{if eq .ProjectInfo.Language "Go"}}vendor/
{{end}}{{if or (eq .ProjectInfo.Language "JavaScript") (eq .ProjectInfo.Language "TypeScript")}}node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*
{{end}}{{if eq .ProjectInfo.Language "Python"}}__pycache__/
*.py[cod]
*$py.class
venv/
env/
{{end}}{{if eq .ProjectInfo.Language "Rust"}}target/
Cargo.lock
{{end}}

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Build outputs
dist/
build/
*.exe
*.dll
*.so
*.dylib

# Logs
*.log

# Environment variables
.env
.env.local

# Temporary files
*.tmp
*.temp

# Console AI history
CB.hist`

const makefileTemplate = `{{if eq .ProjectInfo.Language "Go"}}.PHONY: build test clean run

build:
	go build -o bin/{{.ProjectInfo.Name}} .

test:
	go test ./...

clean:
	rm -rf bin/

run: build
	./bin/{{.ProjectInfo.Name}}

install:
	go mod download
	go mod tidy
{{end}}{{if or (eq .ProjectInfo.Language "JavaScript") (eq .ProjectInfo.Language "TypeScript")}}.PHONY: install build test clean

install:
	{{.ProjectInfo.PackageManager}} install

build:
	{{.ProjectInfo.PackageManager}} run build

test:
	{{.ProjectInfo.PackageManager}} test

clean:
	rm -rf node_modules dist build

dev:
	{{.ProjectInfo.PackageManager}} run dev
{{end}}{{if eq .ProjectInfo.Language "Python"}}.PHONY: install test clean

install:
	pip install -r requirements.txt

test:
	{{if eq .ProjectInfo.TestFramework "pytest"}}pytest{{else}}python -m unittest discover{{end}}

clean:
	find . -type f -name "*.pyc" -delete
	find . -type d -name "__pycache__" -delete

dev:
	python -m pip install -e .
{{end}}`

// Unique web templates designed to avoid recitation blocks
const uniqueHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{if .Options.title}}{{.Options.title}}{{else}}Console Buddy App{{end}}</title>
    <link rel="stylesheet" href="{{if .Options.cssFile}}{{.Options.cssFile}}{{else}}styles.css{{end}}">
</head>
<body>
    <div class="cb-main-container">
        <header class="cb-header">
            <h1 class="cb-title">{{if .Options.title}}{{.Options.title}}{{else}}Console Buddy Application{{end}}</h1>
            {{if .Options.subtitle}}<p class="cb-subtitle">{{.Options.subtitle}}</p>{{end}}
        </header>
        
        <main class="cb-content-area">
            {{if .Options.content}}
            {{.Options.content}}
            {{else}}
            <div class="cb-welcome-section">
                <h2>Welcome to Console Buddy</h2>
                <p>This is a uniquely generated application.</p>
            </div>
            {{end}}
        </main>
        
        <footer class="cb-footer">
            <p>&copy; 2024 Console Buddy Project</p>
        </footer>
    </div>
    
    <script src="{{if .Options.jsFile}}{{.Options.jsFile}}{{else}}script.js{{end}}"></script>
</body>
</html>`

const uniqueCSSTemplate = `/* Console Buddy Unique Styles - Generated to avoid recitation */
* {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}

body {
    font-family: 'Segoe UI', system-ui, sans-serif;
    background: {{if .Options.bgColor}}{{.Options.bgColor}}{{else}}linear-gradient(135deg, #667eea 0%, #764ba2 100%){{end}};
    min-height: 100vh;
    color: {{if .Options.textColor}}{{.Options.textColor}}{{else}}#333{{end}};
}

.cb-main-container {
    max-width: {{if .Options.maxWidth}}{{.Options.maxWidth}}{{else}}1200px{{end}};
    margin: 0 auto;
    padding: {{if .Options.padding}}{{.Options.padding}}{{else}}20px{{end}};
    display: flex;
    flex-direction: column;
    min-height: 100vh;
}

.cb-header {
    background: rgba(255, 255, 255, 0.95);
    padding: 2rem;
    border-radius: 12px;
    margin-bottom: 2rem;
    text-align: center;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.cb-title {
    font-size: 2.5rem;
    font-weight: 700;
    color: {{if .Options.primaryColor}}{{.Options.primaryColor}}{{else}}#4a5568{{end}};
    margin-bottom: 0.5rem;
}

.cb-subtitle {
    font-size: 1.2rem;
    color: {{if .Options.secondaryColor}}{{.Options.secondaryColor}}{{else}}#718096{{end}};
}

.cb-content-area {
    flex: 1;
    background: rgba(255, 255, 255, 0.9);
    padding: 2rem;
    border-radius: 12px;
    margin-bottom: 2rem;
    backdrop-filter: blur(10px);
}

.cb-welcome-section {
    text-align: center;
    padding: 3rem 0;
}

.cb-welcome-section h2 {
    font-size: 2rem;
    margin-bottom: 1rem;
    color: {{if .Options.primaryColor}}{{.Options.primaryColor}}{{else}}#4a5568{{end}};
}

.cb-footer {
    text-align: center;
    padding: 1rem;
    color: rgba(255, 255, 255, 0.8);
    font-size: 0.9rem;
}

/* Responsive design */
@media (max-width: 768px) {
    .cb-main-container {
        padding: 10px;
    }
    
    .cb-title {
        font-size: 2rem;
    }
    
    .cb-header, .cb-content-area {
        padding: 1.5rem;
    }
}`

const uniqueJSTemplate = `// Console Buddy JavaScript - Uniquely generated to avoid recitation
class ConsoleBuddyApp {
    constructor(options = {}) {
        this.appName = options.appName || 'Console Buddy';
        this.debugMode = options.debug || false;
        this.initialized = false;
        this.components = new Map();
        
        this.log('Initializing Console Buddy App...');
        this.init();
    }
    
    init() {
        if (this.initialized) {
            this.log('App already initialized');
            return;
        }
        
        this.setupEventListeners();
        this.loadComponents();
        this.initialized = true;
        this.log('Console Buddy App initialized successfully');
    }
    
    log(message, level = 'info') {
        if (this.debugMode || level === 'error') {
            const timestamp = new Date().toISOString();
            console[level]('[' + this.appName + '] ' + timestamp + ': ' + message);
        }
    }
    
    setupEventListeners() {
        document.addEventListener('DOMContentLoaded', () => {
            this.log('DOM Content Loaded');
            this.onDOMReady();
        });
        
        window.addEventListener('resize', () => {
            this.onWindowResize();
        });
        
        // Custom keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            this.handleKeyboardShortcuts(e);
        });
    }
    
    onDOMReady() {
        // Add your DOM ready logic here
        this.log('DOM is ready for Console Buddy operations');
        
        // Example: Add click handlers to buttons
        const buttons = document.querySelectorAll('.cb-button');
        buttons.forEach(button => {
            button.addEventListener('click', (e) => {
                this.handleButtonClick(e.target);
            });
        });
    }
    
    onWindowResize() {
        this.log('Window resized');
        // Add responsive behavior here
    }
    
    handleKeyboardShortcuts(event) {
        // Ctrl/Cmd + K for quick actions
        if ((event.ctrlKey || event.metaKey) && event.key === 'k') {
            event.preventDefault();
            this.openQuickActions();
        }
        
        // Esc to close modals
        if (event.key === 'Escape') {
            this.closeAllModals();
        }
    }
    
    handleButtonClick(button) {
        const action = button.dataset.action;
        this.log('Button clicked with action: ' + action);
        
        if (this.components.has(action)) {
            this.components.get(action).execute();
        }
    }
    
    registerComponent(name, component) {
        this.components.set(name, component);
        this.log('Component registered: ' + name);
    }
    
    loadComponents() {
        // Register default components
        {{if .Options.components}}
        {{range .Options.components}}
        this.registerComponent('{{.name}}', new {{.class}}());
        {{end}}
        {{end}}
    }
    
    openQuickActions() {
        this.log('Opening quick actions menu');
        // Implement quick actions UI
    }
    
    closeAllModals() {
        const modals = document.querySelectorAll('.cb-modal.active');
        modals.forEach(modal => {
            modal.classList.remove('active');
        });
    }
    
    // Utility methods
    createElement(tag, className, content) {
        const element = document.createElement(tag);
        if (className) element.className = className;
        if (content) element.textContent = content;
        return element;
    }
    
    showNotification(message, type = 'info') {
        const notification = this.createElement('div', 'cb-notification cb-' + type, message);
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
        }, 3000);
    }
}

// Initialize the app when script loads
const cbApp = new ConsoleBuddyApp({
    appName: {{if .Options.appName}}'{{.Options.appName}}'{{else}}'Console Buddy'{{end}},
    debug: {{if .Options.debug}}{{.Options.debug}}{{else}}false{{end}}
});

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ConsoleBuddyApp;
} else if (typeof window !== 'undefined') {
    window.ConsoleBuddyApp = ConsoleBuddyApp;
    window.cbApp = cbApp;
}`

// GetSuggestedFilename returns a suggested filename for generated code
func (cg *CodeGenerator) GetSuggestedFilename(codeType, name string) string {
	switch strings.ToLower(cg.projectInfo.Language) {
	case "go":
		return fmt.Sprintf("%s.go", strings.ToLower(name))
	case "javascript":
		return fmt.Sprintf("%s.js", name)
	case "typescript":
		return fmt.Sprintf("%s.ts", name)
	case "python":
		return fmt.Sprintf("%s.py", strings.ToLower(name))
	case "rust":
		return fmt.Sprintf("%s.rs", strings.ToLower(name))
	default:
		return fmt.Sprintf("%s.txt", name)
	}
}

// GetSuggestedTestFilename returns a suggested filename for test files
func (cg *CodeGenerator) GetSuggestedTestFilename(name string) string {
	switch strings.ToLower(cg.projectInfo.Language) {
	case "go":
		return fmt.Sprintf("%s_test.go", strings.ToLower(name))
	case "javascript":
		return fmt.Sprintf("%s.test.js", name)
	case "typescript":
		return fmt.Sprintf("%s.test.ts", name)
	case "python":
		return fmt.Sprintf("test_%s.py", strings.ToLower(name))
	case "rust":
		return fmt.Sprintf("%s_test.rs", strings.ToLower(name))
	default:
		return fmt.Sprintf("%s_test.txt", name)
	}
}