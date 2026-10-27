package handlers

import (
	"testing"

	"appcuadre/web"
)

func TestNewTemplateRendererParsesAllPages(t *testing.T) {
	renderer, err := NewTemplateRenderer(web.Files, "templates")
	if err != nil {
		t.Fatalf("no se pudieron cargar las plantillas: %v", err)
	}

	expected := []string{
		"dashboard.tmpl",
		"cuadre.tmpl",
		"historial.tmpl",
		"avicola.tmpl",
		"ganancias_avicola.tmpl",
		"consulta_ganancias.tmpl",
		"configuracion.tmpl",
	}

	for _, name := range expected {
		if _, exists := renderer.templates[name]; !exists {
			t.Fatalf("falta la plantilla %s", name)
		}
	}
}
