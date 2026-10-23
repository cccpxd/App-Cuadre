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

// NewTemplateRenderer combina base.html con cada página por separado.
// Lee las plantillas desde templatesDir dentro del sistema de archivos recibido.
func NewTemplateRenderer(files fs.FS, templatesDir string) (*TemplateRenderer, error) {
	basePath := path.Join(templatesDir, "base.html")

	pages, err := fs.Glob(files, path.Join(templatesDir, "*.html"))
	if err != nil {
		return nil, err
	}

	templates := make(map[string]*template.Template)
	for _, page := range pages {
		name := path.Base(page)
		if name == "base.html" || name == "layout.html" {
			continue
		}

		parsed, err := template.ParseFS(files, basePath, page)
		if err != nil {
			return nil, fmt.Errorf("error cargando plantilla %s: %w", name, err)
		}
		templates[name] = parsed
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
