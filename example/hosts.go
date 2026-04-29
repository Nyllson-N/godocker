package main

import docker "github.com/Nyllson-N/godocker"

// client é o cliente Docker usado por todos os handlers.
// Inicializado em initClient() a partir de DOCKER_HOST ou auto-detect.
var client *docker.Client

// initClient cria o cliente Docker.
// Lê DOCKER_HOST do ambiente; sem ele, detecta automaticamente:
//   - Windows / WSL2: tcp://127.0.0.1:2375
//   - Linux:          unix:///var/run/docker.sock
func initClient() error {
	client = docker.New()
	return nil
}
