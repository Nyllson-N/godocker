package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	docker "github.com/Nyllson-N/godocker"
)

// ── Envelope de resposta ──────────────────────────────────────────────────────

type apiResp struct {
	Data  any     `json:"data"`
	Error *string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
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

// ── GET / ─────────────────────────────────────────────────────────────────────

func handleRoutes(w http.ResponseWriter, r *http.Request) {
	ok(w, routes)
}

// ── GET /info ─────────────────────────────────────────────────────────────────

func handleInfo(w http.ResponseWriter, r *http.Request) {
	info, err := docker.Info()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, info)
}

// ── GET /containers  ?all=1 ───────────────────────────────────────────────────

func handleContainers(w http.ResponseWriter, r *http.Request) {
	all := r.URL.Query().Get("all") == "1"
	list, err := docker.ListContainers(all)
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, list)
}

// ── GET /containers/{id} ──────────────────────────────────────────────────────

func handleContainer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inspect, err := docker.InspectContainer(id)
	if err != nil {
		fail(w, http.StatusNotFound, err)
		return
	}
	ok(w, inspect)
}

// ── GET /containers/{id}/stats ────────────────────────────────────────────────

func handleContainerStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	stats, err := docker.ContainerStats(id)
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, stats)
}

// ── GET /containers/{id}/logs  ?tail=N ───────────────────────────────────────

func handleContainerLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tail, _ := strconv.Atoi(r.URL.Query().Get("tail"))
	if tail <= 0 {
		tail = 100
	}
	logs, err := docker.ContainerLogs(id, tail)
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, logs)
}

// ── GET /images ───────────────────────────────────────────────────────────────

func handleImages(w http.ResponseWriter, r *http.Request) {
	images, err := docker.ListImages()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, images)
}

// ── GET /networks ─────────────────────────────────────────────────────────────

func handleNetworks(w http.ResponseWriter, r *http.Request) {
	networks, err := docker.ListNetworks()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, networks)
}

// ── GET /networks/{id} ────────────────────────────────────────────────────────

func handleNetwork(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	network, err := docker.InspectNetwork(id)
	if err != nil {
		fail(w, http.StatusNotFound, err)
		return
	}
	ok(w, network)
}

// ── GET /volumes ──────────────────────────────────────────────────────────────

func handleVolumes(w http.ResponseWriter, r *http.Request) {
	volumes, err := docker.ListVolumes()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	ok(w, volumes)
}

// ── GET /all  ?all=1 ─────────────────────────────────────────────────────────

type allData struct {
	Daemon     *docker.DockerInfo `json:"daemon"`
	Containers []docker.Container `json:"containers"`
	Images     []docker.Image     `json:"images"`
	Networks   []docker.Network   `json:"networks"`
	Volumes    []docker.Volume    `json:"volumes"`
}

func handleAll(w http.ResponseWriter, r *http.Request) {
	allContainers := r.URL.Query().Get("all") == "1"

	info, err := docker.Info()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}

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
