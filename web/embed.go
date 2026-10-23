package web

import "embed"

// Files guarda las plantillas y los archivos estáticos dentro del programa,
// así funcionan igual en tu PC, en Docker y en Vercel.
//
//go:embed templates static
var Files embed.FS
