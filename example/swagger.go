package main

import (
	_ "embed" // necessário para usar as diretivas //go:embed abaixo
	"net/http"
)

// ══════════════════════════════════════════════════════════════════════════════
// SWAGGER UI E ESPECIFICAÇÃO OPENAPI
// Os arquivos docs/swagger.json e docs/swagger.yml são embutidos no binário
// em tempo de compilação via //go:embed — sem precisar servir arquivos externos.
// ══════════════════════════════════════════════════════════════════════════════

// swaggerJSON é a especificação OpenAPI 3.0 em formato JSON,
// lida do arquivo docs/swagger.json em tempo de compilação.
//
//go:embed docs/swagger.json
var swaggerJSON []byte

// swaggerYAML é a especificação OpenAPI 3.0 em formato YAML,
// lida do arquivo docs/swagger.yml em tempo de compilação.
//
//go:embed docs/swagger.yml
var swaggerYAML []byte

// handleSwaggerUI serve a página HTML do Swagger UI carregada via CDN.
// O Swagger UI aponta para /docs/swagger.json como fonte da especificação.
// Funciona com qualquer URL de acesso: localhost, IP, Cloudflare Tunnel, Tailscale, etc.
//
// Rota: GET /swagger
func handleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(swaggerHTML))
}

// handleDocsJSON serve a especificação OpenAPI 3.0 em formato JSON.
// O arquivo é embutido no binário — não depende de arquivos em disco.
//
// Rota: GET /docs/swagger.json
func handleDocsJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(swaggerJSON)
}

// handleDocsYAML serve a especificação OpenAPI 3.0 em formato YAML.
// Pode ser importada diretamente no Postman, Insomnia, etc.
//
// Rota: GET /docs/swagger.yml
func handleDocsYAML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.Write(swaggerYAML)
}

// swaggerHTML é o HTML mínimo que carrega o Swagger UI via CDN (unpkg.com).
// Usa url: "/docs/swagger.json" — funciona com qualquer domínio ou túnel de acesso.
const swaggerHTML = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>godocker API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>body { margin: 0; }</style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "/docs/swagger.json",
      dom_id: "#swagger-ui",
      deepLinking: true,
      requestInterceptor: function(req) {
        // substitui o host gerado pelo swag pelo host atual do browser
        // — funciona com localhost, IP, Cloudflare Tunnel, Tailscale, etc.
        req.url = req.url.replace(/^https?:\/\/[^/]+/, window.location.origin);
        return req;
      },
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
      layout: "BaseLayout"
    });
  </script>
</body>
</html>`
