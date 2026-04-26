// Servidor HTTP que expõe dados do Docker como API REST.
// Cada recurso Docker (containers, imagens, redes, volumes) tem sua própria rota.
// Para rodar: cd example && go run .
// O servidor sobe em http://localhost:8080
package main

import (
	"fmt"      // fmt.Printf — imprime mensagens no terminal
	"log"      // log.Fatal — encerra o programa em caso de erro fatal
	"net/http" // http.NewServeMux, http.HandleFunc, http.ListenAndServe — servidor HTTP
)

// route descreve uma rota da API para exibição no endpoint raiz (GET /).
// Quando alguém acessa GET /, recebe um JSON com esta lista.
type route struct {
	Method      string `json:"method"`          // verbo HTTP: "GET", "POST", etc.
	Path        string `json:"path"`            // caminho da rota (ex: "/containers/{id}")
	Description string `json:"description"`     // o que esta rota faz
	Params      string `json:"params,omitempty"` // parâmetros opcionais de query
}

// routes é a tabela com todas as rotas disponíveis neste servidor.
// É usada tanto para registrar os handlers quanto para exibir em GET /.
var routes = []route{
	{"GET", "/", "lista todas as rotas disponíveis", ""},
	{"GET", "/info", "informações completas do daemon Docker", ""},
	{"GET", "/containers", "lista containers", "?all=1 inclui containers parados"},
	{"GET", "/containers/{id}", "inspect completo de um container", "id = ID ou nome"},
	{"GET", "/containers/{id}/stats", "uso de CPU, memória, rede e I/O de um container", "id = ID ou nome"},
	{"GET", "/containers/{id}/logs", "logs de um container", "id = ID ou nome | ?tail=N (padrão 100)"},
	{"GET", "/images", "lista imagens locais", ""},
	{"GET", "/networks", "lista redes", ""},
	{"GET", "/networks/{id}", "inspect completo de uma rede", "id = ID ou nome"},
	{"GET", "/volumes", "lista volumes", ""},
	{"GET", "/all", "todos os dados de uma vez (daemon + containers + images + networks + volumes)", "?all=1 inclui containers parados"},

	// ── Operações de ciclo de vida ────────────────────────────────────────────
	{"POST", "/containers", "cria um container (sem iniciá-lo)", "body JSON: {Image, Name, Cmd, Env, HostConfig, ...}"},
	{"DELETE", "/containers/{id}", "remove um container", "?force=1 força remoção | ?volumes=1 remove volumes anônimos"},
	{"POST", "/containers/{id}/start", "inicia um container parado", "id = ID ou nome"},
	{"POST", "/containers/{id}/stop", "para um container em execução", "id = ID ou nome | ?t=N segundos de espera"},
	{"POST", "/containers/{id}/restart", "reinicia um container", "id = ID ou nome | ?t=N segundos de espera"},
}

func main() {
	// ServeMux é o roteador HTTP padrão do Go (Go 1.22+).
	// Suporta padrões com método: "GET /path" e parâmetros: {id}
	mux := http.NewServeMux()

	// ── Raiz ──────────────────────────────────────────────────────────────────
	// "GET /{$}" significa exatamente GET / — o /{$} evita que esta rota
	// capture todos os outros paths não registrados
	mux.HandleFunc("GET /{$}", handleRoutes)

	// ── Daemon ────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /info", handleInfo)

	// ── Containers ────────────────────────────────────────────────────────────
	// IMPORTANTE: as rotas mais específicas (/stats, /logs) devem ser registradas
	// ANTES da rota genérica /{id}, senão o roteador não as alcança
	mux.HandleFunc("GET /containers", handleContainers)
	mux.HandleFunc("GET /containers/{id}/stats", handleContainerStats)
	mux.HandleFunc("GET /containers/{id}/logs", handleContainerLogs)
	mux.HandleFunc("GET /containers/{id}", handleContainer)

	// ── Images ────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /images", handleImages)

	// ── Networks ──────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /networks", handleNetworks)
	mux.HandleFunc("GET /networks/{id}", handleNetwork)

	// ── Volumes ───────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /volumes", handleVolumes)

	// ── Tudo de uma vez ───────────────────────────────────────────────────────
	mux.HandleFunc("GET /all", handleAll)

	// ── Ciclo de vida de containers ───────────────────────────────────────────
	// POST /containers usa o mesmo prefixo de "GET /containers", mas o método
	// diferente garante que o Go 1.22 ServeMux roteia para o handler correto.
	mux.HandleFunc("POST /containers", handleCreateContainer)
	mux.HandleFunc("DELETE /containers/{id}", handleRemoveContainer)
	mux.HandleFunc("POST /containers/{id}/start", handleStartContainer)
	mux.HandleFunc("POST /containers/{id}/stop", handleStopContainer)
	mux.HandleFunc("POST /containers/{id}/restart", handleRestartContainer)

	const addr = ":8080"
	fmt.Printf("servidor rodando em http://localhost%s\n", addr)
	fmt.Printf("rotas disponíveis : GET http://localhost%s/\n\n", addr)

	// ListenAndServe bloqueia aqui e só retorna em caso de erro
	// log.Fatal encerra o programa imprimindo o erro
	log.Fatal(http.ListenAndServe(addr, mux))
}
