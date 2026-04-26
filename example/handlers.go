// handlers.go contém todos os handlers HTTP do servidor.
// Cada handler corresponde a uma rota registrada em main.go.
// Todos seguem o mesmo padrão:
//  1. Lê parâmetros da URL (path params ou query string)
//  2. Chama a função da biblioteca godocker
//  3. Retorna ok(w, dados) em caso de sucesso ou fail(w, status, err) em caso de erro
package main

import (
	"encoding/json" // json.NewEncoder — serializa structs Go em JSON para a resposta HTTP
	"net/http"      // http.ResponseWriter, http.Request, http.Status* — tipos e constantes HTTP
	"strconv"       // strconv.Atoi — converte string para int (parâmetro ?tail=N)

	docker "github.com/Nyllson-N/godocker" // biblioteca godocker — acesso ao Docker
)

// ══════════════════════════════════════════════════════════════════════════════
// ENVELOPE DE RESPOSTA
// Todas as respostas usam o mesmo formato JSON para facilitar o consumo da API.
// ══════════════════════════════════════════════════════════════════════════════

// apiResp é o envelope padrão de todas as respostas da API.
// Em sucesso: {"data": {...}, "error": null}
// Em erro:    {"data": null, "error": "mensagem de erro"}
type apiResp struct {
	Data  any     `json:"data"`  // dados retornados (qualquer tipo Go → JSON)
	Error *string `json:"error"` // ponteiro: null em sucesso, string em erro
}

// writeJSON escreve uma resposta HTTP com Content-Type JSON e indentação.
// É a função base chamada por ok() e fail().
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status) // define o código HTTP (200, 404, 500, etc.)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ") // indenta o JSON com 2 espaços para facilitar leitura no browser
	enc.Encode(v)           // serializa v como JSON e escreve direto no ResponseWriter
}

// ok retorna HTTP 200 com os dados no campo "data" e "error": null.
func ok(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, apiResp{Data: data})
}

// fail retorna um código HTTP de erro com a mensagem no campo "error" e "data": null.
// O erro do Docker (ex: "docker API 404: no such container") vai direto para o cliente.
func fail(w http.ResponseWriter, status int, err error) {
	msg := err.Error()    // extrai a string do erro
	writeJSON(w, status, apiResp{Error: &msg}) // ponteiro para string → não é null no JSON
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: GET /
// Retorna a lista de todas as rotas disponíveis neste servidor.
// ══════════════════════════════════════════════════════════════════════════════

func handleRoutes(w http.ResponseWriter, r *http.Request) {
	// `routes` é a variável global definida em main.go com todas as rotas registradas
	ok(w, routes)
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: GET /info
// Retorna informações completas do daemon Docker.
// ══════════════════════════════════════════════════════════════════════════════

func handleInfo(w http.ResponseWriter, r *http.Request) {
	info, err := docker.Info() // chama GET /v1.47/info no Docker
	if err != nil {
		// erro mais comum: Docker não está rodando ou não acessível
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, info)
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: GET /containers?all=1
// Lista containers. Sem ?all=1, retorna apenas os em execução.
// ══════════════════════════════════════════════════════════════════════════════

func handleContainers(w http.ResponseWriter, r *http.Request) {
	// r.URL.Query().Get("all") lê o parâmetro ?all=1 da URL
	// se for "1", inclui containers parados na listagem
	all := r.URL.Query().Get("all") == "1"

	list, err := docker.ListContainers(all)
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, list)
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: GET /containers/{id}
// Retorna o inspect completo de um container pelo ID ou nome.
// ══════════════════════════════════════════════════════════════════════════════

func handleContainer(w http.ResponseWriter, r *http.Request) {
	// r.PathValue("id") extrai o {id} do padrão da URL (Go 1.22+)
	// ex: GET /containers/meu-nginx → id = "meu-nginx"
	id := r.PathValue("id")

	inspect, err := docker.InspectContainer(id)
	if err != nil {
		// 404 quando o container não existe
		fail(w, http.StatusNotFound, err)
		return
	}
	ok(w, inspect)
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: GET /containers/{id}/stats
// Retorna um snapshot de uso de CPU, memória, I/O e rede do container.
// Só funciona para containers em execução.
// ══════════════════════════════════════════════════════════════════════════════

func handleContainerStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	stats, err := docker.ContainerStats(id)
	if err != nil {
		// erro comum: container parado não tem stats disponíveis
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, stats)
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: GET /containers/{id}/logs?tail=N
// Retorna as últimas N linhas de log do container (padrão: 100).
// ══════════════════════════════════════════════════════════════════════════════

func handleContainerLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// lê o parâmetro ?tail=50 da URL e converte para int
	// strconv.Atoi retorna 0 e erro se o valor não for número — ignoramos o erro
	tail, _ := strconv.Atoi(r.URL.Query().Get("tail"))
	if tail <= 0 {
		tail = 100 // valor padrão quando ?tail não foi passado ou é inválido
	}

	logs, err := docker.ContainerLogs(id, tail)
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, logs) // retorna as linhas de log como string dentro do campo "data"
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: GET /images
// Lista todas as imagens Docker armazenadas localmente.
// ══════════════════════════════════════════════════════════════════════════════

func handleImages(w http.ResponseWriter, r *http.Request) {
	images, err := docker.ListImages()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, images)
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: GET /networks
// Lista todas as redes Docker disponíveis.
// ══════════════════════════════════════════════════════════════════════════════

func handleNetworks(w http.ResponseWriter, r *http.Request) {
	networks, err := docker.ListNetworks()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, networks)
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: GET /networks/{id}
// Retorna o inspect completo de uma rede pelo ID ou nome.
// ══════════════════════════════════════════════════════════════════════════════

func handleNetwork(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	network, err := docker.InspectNetwork(id)
	if err != nil {
		// 404 quando a rede não existe
		fail(w, http.StatusNotFound, err)
		return
	}
	ok(w, network)
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: GET /volumes
// Lista todos os volumes Docker criados na máquina.
// ══════════════════════════════════════════════════════════════════════════════

func handleVolumes(w http.ResponseWriter, r *http.Request) {
	volumes, err := docker.ListVolumes()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	
	ok(w, volumes)
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: GET /all?all=1
// Retorna todos os recursos em uma única requisição.
// Útil para dashboards que precisam carregar tudo de uma vez.
// ══════════════════════════════════════════════════════════════════════════════

// allData agrupa todos os recursos Docker em uma única estrutura
// para ser retornada no endpoint GET /all.
type allData struct {
	Daemon     *docker.DockerInfo `json:"daemon"`     // informações do daemon
	Containers []docker.Container `json:"containers"` // lista de containers
	Images     []docker.Image     `json:"images"`     // lista de imagens
	Networks   []docker.Network   `json:"networks"`   // lista de redes
	Volumes    []docker.Volume    `json:"volumes"`    // lista de volumes
}

func handleAll(w http.ResponseWriter, r *http.Request) {
	allContainers := r.URL.Query().Get("all") == "1"

	// Info é verificado primeiro: se o Docker não estiver acessível,
	// não faz sentido tentar buscar o resto
	info, err := docker.Info()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}

	// Os demais recursos ignoram erros individualmente com `_`.
	// Se uma chamada falhar (ex: permissão em volumes), os outros ainda são retornados.
	// O campo ficará nil/vazio no JSON mas o request não falha por completo.
	containers, _ := docker.ListContainers(allContainers)
	images, _ := docker.ListImages()
	networks, _ := docker.ListNetworks()
	volumes, _ := docker.ListVolumes()

	ok(w, allData{
		Daemon:     info,
		Containers: containers,
		Images:     images,
		Networks:   networks,
		Volumes:    volumes,
	})
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: POST /containers
// Cria um container sem iniciá-lo.
// O corpo da requisição deve ser um JSON com as opções de criação.
// O nome do container pode ser passado como ?name=meu-container na URL.
// Retorna HTTP 201 com o ID gerado e possíveis avisos do Docker.
// ══════════════════════════════════════════════════════════════════════════════

func handleCreateContainer(w http.ResponseWriter, r *http.Request) {
	// decodifica o JSON do corpo da requisição em ContainerCreateOptions
	var opts docker.ContainerCreateOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}

	// o nome pode vir como ?name=meu-nginx na URL (o campo json:"-" não vem do body)
	if name := r.URL.Query().Get("name"); name != "" {
		opts.Name = name
	}

	result, err := docker.CreateContainer(opts)
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	// 201 Created em vez de 200 — semântica correta para criação de recurso
	writeJSON(w, http.StatusCreated, apiResp{Data: result})
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: DELETE /containers/{id}
// Remove um container pelo ID ou nome.
// ?force=1   → remove mesmo que o container esteja rodando
// ?volumes=1 → remove também os volumes anônimos associados ao container
// ══════════════════════════════════════════════════════════════════════════════

func handleRemoveContainer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	force := r.URL.Query().Get("force") == "1"
	removeVolumes := r.URL.Query().Get("volumes") == "1"

	if err := docker.RemoveContainer(id, force, removeVolumes); err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, map[string]string{"id": id, "status": "removed"})
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: POST /containers/{id}/start
// Inicia um container parado.
// Retorna sucesso mesmo que o container já esteja rodando (Docker responde 304).
// ══════════════════════════════════════════════════════════════════════════════

func handleStartContainer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := docker.StartContainer(id); err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, map[string]string{"id": id, "status": "started"})
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: POST /containers/{id}/stop
// Para um container em execução.
// ?t=N → aguarda N segundos pelo encerramento antes de enviar SIGKILL.
// Retorna sucesso mesmo que o container já esteja parado (Docker responde 304).
// ══════════════════════════════════════════════════════════════════════════════

func handleStopContainer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// ?t=N configura o timeout; se não fornecido ou inválido, passa 0 (usa padrão do Docker)
	timeout, _ := strconv.Atoi(r.URL.Query().Get("t"))

	if err := docker.StopContainer(id, timeout); err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, map[string]string{"id": id, "status": "stopped"})
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLER: POST /containers/{id}/restart
// Reinicia um container (stop + start internamente no Docker).
// ?t=N → aguarda N segundos pelo encerramento antes de SIGKILL na fase de stop.
// ══════════════════════════════════════════════════════════════════════════════

func handleRestartContainer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	timeout, _ := strconv.Atoi(r.URL.Query().Get("t"))

	if err := docker.RestartContainer(id, timeout); err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, map[string]string{"id": id, "status": "restarted"})
}
