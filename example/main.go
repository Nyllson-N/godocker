package main

import (
	"fmt"
	"log"
	"net/http"
)

// route descreve uma rota da API para exibição no endpoint raiz.
type route struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Params      string `json:"params,omitempty"`
}

// routes é a tabela de todas as rotas disponíveis — exibida em GET /.
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
}

func main() {
	mux := http.NewServeMux()

	// ── Raiz ──────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /{$}", handleRoutes)

	// ── Daemon ────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /info", handleInfo)

	// ── Containers ────────────────────────────────────────────────────────────
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

	const addr = ":8080"
	fmt.Printf("servidor rodando em http://localhost%s\n", addr)
	fmt.Printf("rotas disponíveis : GET http://localhost%s/\n\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
