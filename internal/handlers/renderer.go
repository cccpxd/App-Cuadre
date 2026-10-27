package handlers

import (
	"fmt"
	"html/template"
	"io/fs"
	"path"

	"github.com/gin-gonic/gin/render"
)

// TemplateRenderer guarda un conjunto de plantillas independiente por página,
// para que el bloque "content" de una página no pise el de otra.
type TemplateRenderer struct {
	templates map[string]*template.Template
}

// NewTemplateRenderer combina la base con cada página por separado.
// Lee las plantillas desde templatesDir dentro del sistema de archivos recibido.
func NewTemplateRenderer(files fs.FS, templatesDir string) (*TemplateRenderer, error) {
	baseCandidates := []string{
		path.Join(templatesDir, "base.tmpl"),
		path.Join(templatesDir, "base.html"),
	}

	basePath := ""
	for _, candidate := range baseCandidates {
		if _, err := fs.Stat(files, candidate); err == nil {
			basePath = candidate
			break
		}
	}
	if basePath == "" {
		return nil, fmt.Errorf("no se encontró la plantilla base en %s", templatesDir)
	}

	pagePatterns := []string{
		path.Join(templatesDir, "*.tmpl"),
		path.Join(templatesDir, "*.html"),
	}

	templates := make(map[string]*template.Template)
	for _, pattern := range pagePatterns {
		pages, err := fs.Glob(files, pattern)
		if err != nil {
			return nil, err
		}

		for _, page := range pages {
			name := path.Base(page)
			if name == path.Base(basePath) || name == "layout.tmpl" || name == "layout.html" {
				continue
			}

			parsed, err := template.ParseFS(files, basePath, page)
			if err != nil {
				return nil, fmt.Errorf("error cargando plantilla %s: %w", name, err)
			}
			templates[name] = parsed
		}
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf("no se encontraron plantillas en %s", templatesDir)
	}

	return &TemplateRenderer{templates: templates}, nil
}

// Instance implementa gin's render.HTMLRender.
func (r *TemplateRenderer) Instance(name string, data any) render.Render {
	tpl, ok := r.templates[name]
	if !ok {
		return render.String{Format: "plantilla no encontrada: %s", Data: []any{name}}
	}
	return render.HTML{Template: tpl, Name: name, Data: data}
}
