// Servidor HTTP que expõe dados do Docker como API REST.
// Para rodar: cd example && go run .  (ou execute inicia.bat no Windows)
// Configuracao via .env — copie .env.example para .env e ajuste os valores.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// route descreve uma rota da API para exibicao no endpoint raiz (GET /).
type route struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Params      string `json:"params,omitempty"`
}

// routes é a tabela com todas as rotas disponíveis neste servidor.
var routes = []route{
	// Utilitários
	{"GET", "/", "lista todas as rotas disponíveis", ""},
	{"GET", "/swagger", "interface Swagger UI para explorar a API", ""},
	{"GET", "/docs/swagger.json", "especificação OpenAPI em JSON", ""},
	{"GET", "/docs/swagger.yml", "especificação OpenAPI em YAML", ""},

	// Daemon
	{"GET", "/info", "informações completas do daemon Docker", ""},

	// Containers — leitura
	{"GET", "/containers", "lista containers", "?all=1 inclui parados"},
	{"GET", "/containers/{id}", "inspect completo de um container", ""},
	{"GET", "/containers/{id}/stats", "métricas de CPU, memória, I/O e rede", ""},
	{"GET", "/containers/{id}/logs", "logs do container", "?tail=N"},

	// Containers — ciclo de vida
	{"POST", "/containers", "cria um container (sem iniciá-lo)", "?name= | body JSON"},
	{"DELETE", "/containers/{id}", "remove um container", "?force=1 | ?volumes=1"},
	{"POST", "/containers/{id}/start", "inicia um container parado", ""},
	{"POST", "/containers/{id}/stop", "para um container", "?t=N segundos"},
	{"POST", "/containers/{id}/restart", "reinicia um container", "?t=N segundos"},

	// Imagens
	{"GET", "/images", "lista imagens locais", ""},

	// Redes
	{"GET", "/networks", "lista redes Docker", ""},
	{"GET", "/networks/{id}", "inspect completo de uma rede", ""},

	// Volumes
	{"GET", "/volumes", "lista volumes Docker", ""},

	// Agregado
	{"GET", "/all", "todos os recursos em uma requisição", "?all=1"},
}

//	@title		godocker API
//	@version	1.0
//	@description	API REST para gerenciar recursos Docker.
//	@BasePath	/
func main() {
	if err := loadDotEnv(".env"); err != nil {
		log.Printf("aviso: erro ao ler .env: %v", err)
	}

	if err := initClient(); err != nil {
		log.Fatalf("erro ao configurar cliente Docker: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Utilitários
	mux.HandleFunc("GET /{$}", handleRoutes)
	mux.HandleFunc("GET /swagger", handleSwaggerUI)
	mux.HandleFunc("GET /docs/swagger.json", handleDocsJSON)
	mux.HandleFunc("GET /docs/swagger.yml", handleDocsYAML)

	// Daemon
	mux.HandleFunc("GET /info", handleInfo)

	// Containers — leitura (rotas específicas antes da genérica /{id})
	mux.HandleFunc("GET /containers", handleContainers)
	mux.HandleFunc("GET /containers/{id}/stats", handleContainerStats)
	mux.HandleFunc("GET /containers/{id}/logs", handleContainerLogs)
	mux.HandleFunc("GET /containers/{id}", handleContainer)

	// Containers — ciclo de vida
	mux.HandleFunc("POST /containers", handleCreateContainer)
	mux.HandleFunc("DELETE /containers/{id}", handleRemoveContainer)
	mux.HandleFunc("POST /containers/{id}/start", handleStartContainer)
	mux.HandleFunc("POST /containers/{id}/stop", handleStopContainer)
	mux.HandleFunc("POST /containers/{id}/restart", handleRestartContainer)

	// Imagens
	mux.HandleFunc("GET /images", handleImages)

	// Redes
	mux.HandleFunc("GET /networks", handleNetworks)
	mux.HandleFunc("GET /networks/{id}", handleNetwork)

	// Volumes
	mux.HandleFunc("GET /volumes", handleVolumes)

	// Agregado
	mux.HandleFunc("GET /all", handleAll)

	addr := ":" + port
	fmt.Printf("servidor rodando em http://localhost%s\n", addr)
	fmt.Printf("swagger UI : http://localhost%s/swagger\n\n", addr)

	log.Fatal(http.ListenAndServe(addr, mux))
}
