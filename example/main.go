package main

import (
	"fmt"
	"runtime"
	"strings"

	docker "github.com/Nyllson-N/godocker"
)

func main() {
	fmt.Printf("OS: %s | modo: %s → %s\n\n",
		runtime.GOOS, docker.DefaultClient.Mode(), docker.DefaultClient.BaseURL())

	// ── Info do daemon ────────────────────────────────────────────────────────
	info, err := docker.Info()
	if err != nil {
		fmt.Println("Erro ao conectar:", err)
		return
	}
	fmt.Printf("Docker %s | %d rodando / %d total | %s\n\n",
		info.ServerVersion, info.ContainersRunning, info.Containers, info.OperatingSystem)

	// ── Redes ─────────────────────────────────────────────────────────────────
	fmt.Println("═══ Redes ═══")
	networks, _ := docker.ListNetworks()
	for _, n := range networks {
		subnet := ""
		if len(n.IPAM.Config) > 0 {
			subnet = n.IPAM.Config[0].Subnet
		}
		fmt.Printf("  %-20s driver=%-10s scope=%-8s %s\n",
			n.Name, n.Driver, n.Scope, subnet)
	}

	// ── Criar / remover rede de teste ─────────────────────────────────────────
	fmt.Println("\n═══ Teste CreateNetwork / RemoveNetwork ═══")
	id, err := docker.CreateNetwork(docker.NetworkCreateOptions{
		Name:   "teste-godocker",
		Driver: "bridge",
		Labels: map[string]string{"criado-por": "godocker"},
		IPAM: &docker.NetworkIPAM{
			Driver: "default",
			Config: []docker.IPAMConfig{
				{Subnet: "172.30.0.0/24", Gateway: "172.30.0.1"},
			},
		},
	})
	if err != nil {
		fmt.Println("Erro ao criar rede:", err)
	} else {
		fmt.Printf("Criada: %s\n", id)
		_ = docker.RemoveNetwork(id)
		fmt.Println("Removida.")
	}

	// ── Containers ────────────────────────────────────────────────────────────
	fmt.Println("\n═══ Containers ═══")
	containers, _ := docker.ListContainers(true)
	for _, c := range containers {
		name := strings.TrimPrefix(c.Names[0], "/")
		fmt.Printf("  [%s] %-30s %s\n", c.State, name, c.Status)

		if c.State == "running" {
			stats, err := docker.ContainerStats(c.ID)
			if err == nil {
				cpu := docker.CPUPercent(stats)
				mem := float64(stats.MemoryStats.Usage) / 1024 / 1024
				lim := float64(stats.MemoryStats.Limit) / 1024 / 1024
				fmt.Printf("         CPU: %.2f%% | Mem: %.1fMB / %.1fMB\n", cpu, mem, lim)
			}
		}
	}
}
