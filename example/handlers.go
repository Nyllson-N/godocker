package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	docker "github.com/Nyllson-N/godocker"
)

// apiResp é o envelope padrão de todas as respostas da API.
// Em sucesso: {"data": {...}, "error": null}
// Em erro:    {"data": null, "error": "mensagem de erro"}
type apiResp struct {
	Data  any     `json:"data"`
	Error *string `json:"error"`
}

// statusResp é a resposta das operações de ciclo de vida (start, stop, restart, remove).
type statusResp struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// allData agrupa todos os recursos Docker em uma única estrutura para GET /all.
type allData struct {
	Daemon     *docker.DockerInfo `json:"daemon"`
	Containers []docker.Container `json:"containers"`
	Images     []docker.Image     `json:"images"`
	Networks   []docker.Network   `json:"networks"`
	Volumes    []docker.Volume    `json:"volumes"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func ok(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, apiResp{Data: data})
}

func fail(w http.ResponseWriter, status int, err error) {
	msg := err.Error()
	writeJSON(w, status, apiResp{Error: &msg})
}

// handleRoutes godoc
//
//	@Summary	Lista todas as rotas disponíveis
//	@Tags		Utilitários
//	@Produce	json
//	@Success	200	{array}	route
//	@Router		/ [get]
func handleRoutes(w http.ResponseWriter, r *http.Request) {
	ok(w, routes)
}

// handleInfo godoc
//
//	@Summary		Informações do daemon Docker
//	@Description	Versão, SO, número de containers, memória, plugins, etc.
//	@Tags			Daemon
//	@Produce		json
//	@Success		200	{object}	apiResp{data=docker.DockerInfo}
//	@Failure		500	{object}	apiResp	"Docker inacessível"
//	@Router			/info [get]
func handleInfo(w http.ResponseWriter, r *http.Request) {
	info, err := client.Info()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, info)
}

// handleContainers godoc
//
//	@Summary		Lista containers
//	@Description	Containers em execução. Use ?all=1 para incluir os parados.
//	@Tags			Containers
//	@Produce		json
//	@Param			all	query		int	false	"1 = inclui containers parados"	Enums(0,1)
//	@Success		200	{object}	apiResp{data=[]docker.Container}
//	@Failure		500	{object}	apiResp	"erro Docker"
//	@Router			/containers [get]
func handleContainers(w http.ResponseWriter, r *http.Request) {
	all := r.URL.Query().Get("all") == "1"
	list, err := client.ListContainers(all)
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, list)
}

// handleContainer godoc
//
//	@Summary		Inspect completo de um container
//	@Description	Retorna todos os dados do container: config, rede, estado, mounts, etc.
//	@Tags			Containers
//	@Produce		json
//	@Param			id	path		string	true	"ID completo, prefixo ou nome do container"
//	@Success		200	{object}	apiResp{data=docker.ContainerInspect}
//	@Failure		404	{object}	apiResp	"container não encontrado"
//	@Router			/containers/{id} [get]
func handleContainer(w http.ResponseWriter, r *http.Request) {
	inspect, err := client.InspectContainer(r.PathValue("id"))
	if err != nil {
		fail(w, http.StatusNotFound, err)
		return
	}
	ok(w, inspect)
}

// handleContainerStats godoc
//
//	@Summary		Métricas de CPU, memória, I/O e rede
//	@Description	Snapshot de uso de recursos. Só funciona para containers em execução.
//	@Tags			Containers
//	@Produce		json
//	@Param			id	path		string	true	"ID ou nome do container"
//	@Success		200	{object}	apiResp{data=docker.Stats}
//	@Failure		500	{object}	apiResp	"container parado ou erro"
//	@Router			/containers/{id}/stats [get]
func handleContainerStats(w http.ResponseWriter, r *http.Request) {
	stats, err := client.ContainerStats(r.PathValue("id"))
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, stats)
}

// handleContainerLogs godoc
//
//	@Summary		Logs do container
//	@Description	Últimas N linhas de log (stdout + stderr). Padrão: 100 linhas.
//	@Tags			Containers
//	@Produce		json
//	@Param			id		path		string	true	"ID ou nome do container"
//	@Param			tail	query		int		false	"Número de linhas (padrão 100)"
//	@Success		200		{object}	apiResp{data=string}
//	@Failure		500		{object}	apiResp	"erro ao buscar logs"
//	@Router			/containers/{id}/logs [get]
func handleContainerLogs(w http.ResponseWriter, r *http.Request) {
	tail, _ := strconv.Atoi(r.URL.Query().Get("tail"))
	if tail <= 0 {
		tail = 100
	}
	logs, err := client.ContainerLogs(r.PathValue("id"), tail)
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, logs)
}

// handleCreateContainer godoc
//
//	@Summary		Cria um container
//	@Description	Cria sem iniciar. Use POST /containers/{id}/start para iniciar.
//	@Tags			Containers
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string							false	"Nome do container"
//	@Param			body	body		docker.ContainerCreateOptions	true	"Opções de criação"
//	@Success		201		{object}	apiResp{data=docker.ContainerCreateResponse}
//	@Failure		400		{object}	apiResp	"body inválido"
//	@Failure		500		{object}	apiResp	"erro Docker"
//	@Router			/containers [post]
func handleCreateContainer(w http.ResponseWriter, r *http.Request) {
	var opts docker.ContainerCreateOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if name := r.URL.Query().Get("name"); name != "" {
		opts.Name = name
	}
	result, err := client.CreateContainer(opts)
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, apiResp{Data: result})
}

// handleRemoveContainer godoc
//
//	@Summary		Remove um container
//	@Description	Remove permanentemente. ?force=1 remove mesmo estando em execução.
//	@Tags			Containers
//	@Produce		json
//	@Param			id		path		string	true	"ID ou nome do container"
//	@Param			force	query		int		false	"1 = remove mesmo rodando"		Enums(0,1)
//	@Param			volumes	query		int		false	"1 = remove volumes anônimos"	Enums(0,1)
//	@Success		200		{object}	apiResp{data=statusResp}
//	@Failure		500		{object}	apiResp	"erro Docker"
//	@Router			/containers/{id} [delete]
func handleRemoveContainer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	force := r.URL.Query().Get("force") == "1"
	removeVolumes := r.URL.Query().Get("volumes") == "1"
	if err := client.RemoveContainer(id, force, removeVolumes); err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, statusResp{ID: id, Status: "removed"})
}

// handleStartContainer godoc
//
//	@Summary		Inicia um container
//	@Description	Sem erro se já estiver rodando (idempotente).
//	@Tags			Containers
//	@Produce		json
//	@Param			id	path		string	true	"ID ou nome do container"
//	@Success		200	{object}	apiResp{data=statusResp}
//	@Failure		500	{object}	apiResp	"erro Docker"
//	@Router			/containers/{id}/start [post]
func handleStartContainer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := client.StartContainer(id); err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, statusResp{ID: id, Status: "started"})
}

// handleStopContainer godoc
//
//	@Summary		Para um container
//	@Description	Envia SIGTERM e aguarda. Após timeout envia SIGKILL. Retorna sucesso se já parado.
//	@Tags			Containers
//	@Produce		json
//	@Param			id	path		string	true	"ID ou nome do container"
//	@Param			t	query		int		false	"Segundos antes do SIGKILL (padrão 10)"
//	@Success		200	{object}	apiResp{data=statusResp}
//	@Failure		500	{object}	apiResp	"erro Docker"
//	@Router			/containers/{id}/stop [post]
func handleStopContainer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	timeout, _ := strconv.Atoi(r.URL.Query().Get("t"))
	if err := client.StopContainer(id, timeout); err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, statusResp{ID: id, Status: "stopped"})
}

// handleRestartContainer godoc
//
//	@Summary		Reinicia um container
//	@Description	Equivale a stop + start. O ?t=N aplica-se à fase de stop.
//	@Tags			Containers
//	@Produce		json
//	@Param			id	path		string	true	"ID ou nome do container"
//	@Param			t	query		int		false	"Segundos antes do SIGKILL no stop"
//	@Success		200	{object}	apiResp{data=statusResp}
//	@Failure		500	{object}	apiResp	"erro Docker"
//	@Router			/containers/{id}/restart [post]
func handleRestartContainer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	timeout, _ := strconv.Atoi(r.URL.Query().Get("t"))
	if err := client.RestartContainer(id, timeout); err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, statusResp{ID: id, Status: "restarted"})
}

// handleImages godoc
//
//	@Summary	Lista imagens locais
//	@Tags		Imagens
//	@Produce	json
//	@Success	200	{object}	apiResp{data=[]docker.Image}
//	@Failure	500	{object}	apiResp	"erro Docker"
//	@Router		/images [get]
func handleImages(w http.ResponseWriter, r *http.Request) {
	images, err := client.ListImages()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, images)
}

// handleNetworks godoc
//
//	@Summary	Lista redes Docker
//	@Tags		Redes
//	@Produce	json
//	@Success	200	{object}	apiResp{data=[]docker.Network}
//	@Failure	500	{object}	apiResp	"erro Docker"
//	@Router		/networks [get]
func handleNetworks(w http.ResponseWriter, r *http.Request) {
	networks, err := client.ListNetworks()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, networks)
}

// handleNetwork godoc
//
//	@Summary		Inspect completo de uma rede
//	@Description	Driver, configuração IPAM e containers conectados.
//	@Tags			Redes
//	@Produce		json
//	@Param			id	path		string	true	"ID ou nome da rede"
//	@Success		200	{object}	apiResp{data=docker.Network}
//	@Failure		404	{object}	apiResp	"rede não encontrada"
//	@Router			/networks/{id} [get]
func handleNetwork(w http.ResponseWriter, r *http.Request) {
	network, err := client.InspectNetwork(r.PathValue("id"))
	if err != nil {
		fail(w, http.StatusNotFound, err)
		return
	}
	ok(w, network)
}

// handleVolumes godoc
//
//	@Summary	Lista volumes Docker
//	@Tags		Volumes
//	@Produce	json
//	@Success	200	{object}	apiResp{data=[]docker.Volume}
//	@Failure	500	{object}	apiResp	"erro Docker"
//	@Router		/volumes [get]
func handleVolumes(w http.ResponseWriter, r *http.Request) {
	volumes, err := client.ListVolumes()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, volumes)
}

// handleAll godoc
//
//	@Summary		Todos os recursos em uma requisição
//	@Description	daemon + containers + images + networks + volumes. Útil para dashboards.
//	@Tags			Utilitários
//	@Produce		json
//	@Param			all	query		int	false	"1 = inclui containers parados"	Enums(0,1)
//	@Success		200	{object}	apiResp{data=allData}
//	@Failure		500	{object}	apiResp	"Docker inacessível"
//	@Router			/all [get]
func handleAll(w http.ResponseWriter, r *http.Request) {
	all := r.URL.Query().Get("all") == "1"

	info, err := client.Info()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}

	containers, _ := client.ListContainers(all)
	images, _ := client.ListImages()
	networks, _ := client.ListNetworks()
	volumes, _ := client.ListVolumes()

	ok(w, allData{
		Daemon:     info,
		Containers: containers,
		Images:     images,
		Networks:   networks,
		Volumes:    volumes,
	})
}
