package prompt

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

// Content holds the embedded prompt templates
//
//go:embed templates/*.txt
var Content embed.FS

// Manager handles prompt loading and execution
type Manager struct {
	templates map[string]*template.Template
}

// NewManager creates a new prompt manager
func NewManager() (*Manager, error) {
	m := &Manager{
		templates: make(map[string]*template.Template),
	}

	// Load default template
	if err := m.loadTemplate("default_analyze", "templates/default_analyze.txt"); err != nil {
		return nil, err
	}

	// Load teacher lesson template
	if err := m.loadTemplate("teacher_lesson", "templates/teacher_lesson.txt"); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) loadTemplate(name, path string) error {
	tmplContent, err := Content.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", path, err)
	}

	tmpl, err := template.New(name).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	m.templates[name] = tmpl
	return nil
}

// Execute executes a template with the given data
func (m *Manager) Execute(name string, data interface{}) (string, error) {
	tmpl, ok := m.templates[name]
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}
