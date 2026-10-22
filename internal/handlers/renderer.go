package handlers

import (
    "fmt"
    "html/template"
    "path/filepath"

    "github.com/gin-gonic/gin/render"
)

// TemplateRenderer guarda un conjunto de plantillas independiente por página,
// para que el bloque "content" de una página no pise el de otra.
type TemplateRenderer struct {
    templates map[string]*template.Template
}

// NewTemplateRenderer combina base.html con cada página por separado.
func NewTemplateRenderer(templatesDir string) (*TemplateRenderer, error) {
    basePath := filepath.Join(templatesDir, "base.html")

    pages, err := filepath.Glob(filepath.Join(templatesDir, "*.html"))
    if err != nil {
        return nil, err
    }

    templates := make(map[string]*template.Template)
    for _, page := range pages {
        name := filepath.Base(page)
        if name == "base.html" || name == "layout.html" {
            continue
        }

        parsed, err := template.ParseFiles(basePath, page)
        if err != nil {
            return nil, fmt.Errorf("error cargando plantilla %s: %w", name, err)
        }
        templates[name] = parsed
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
