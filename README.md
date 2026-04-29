# godocker

Cliente Go para a Docker Engine API — sem dependências externas.

Detecta automaticamente o ambiente:

| Ambiente | Transporte |
|---|---|
| `DOCKER_HOST` definido | TCP ou Unix conforme o valor |
| Windows / WSL2 | TCP `localhost:2375` |
| Linux nativo | Unix socket `/var/run/docker.sock` |
| Fallback | TCP `localhost:2375` |

## Instalação

```bash
go get github.com/Nyllson-N/godocker
```

## Uso rápido

```go
import docker "github.com/Nyllson-N/godocker"

// DefaultClient detecta o ambiente automaticamente
info, err := docker.Info()
containers, err := docker.ListContainers(true) // true = inclui parados
networks, err := docker.ListNetworks()
images, err := docker.ListImages()
volumes, err := docker.ListVolumes()
```

## Carregando configuração via .env

```go
func main() {
    docker.LoadEnv(".env") // lê DOCKER_HOST, DOCKER_HOSTS, PORT, etc.

    info, err := docker.Info()
}
```

O `.env` é opcional — variáveis já definidas no shell têm prioridade.

## Cliente personalizado

```go
// TCP explícito
client := docker.NewTCP("192.168.1.10:2375")

// Unix socket explícito
client := docker.NewUnix("/var/run/docker.sock")

// Detecção automática
client := docker.New()

// Todos os métodos estão disponíveis no client
info, err := client.Info()
list, err := client.ListContainers(false)
```

## Referência da API

### Daemon

```go
docker.Info() (*DockerInfo, error)
```

### Containers — leitura

```go
docker.ListContainers(all bool) ([]Container, error)
docker.InspectContainer(id string) (*ContainerInspect, error)
docker.ContainerStats(id string) (*Stats, error)
docker.ContainerLogs(id string, tail int) (string, error)
docker.CPUPercent(s *Stats) float64
```

### Containers — ciclo de vida

```go
docker.CreateContainer(opts ContainerCreateOptions) (*ContainerCreateResponse, error)
docker.StartContainer(id string) error
docker.StopContainer(id string, timeout int) error
docker.RestartContainer(id string, timeout int) error
docker.RemoveContainer(id string, force, removeVolumes bool) error
```

### Imagens

```go
docker.ListImages() ([]Image, error)
```

### Redes

```go
docker.ListNetworks() ([]Network, error)
docker.InspectNetwork(id string) (*Network, error)
docker.CreateNetwork(opts NetworkCreateOptions) (string, error)
docker.RemoveNetwork(id string) error
docker.ConnectNetwork(networkID, containerID string, opts *NetworkConnectOptions) error
docker.DisconnectNetwork(networkID, containerID string, force bool) error
docker.PruneNetworks() ([]string, error)
docker.NetworkExists(nameOrID string) (bool, error)
docker.RenameNetwork(oldName, newName string) (string, error)
```

### Volumes

```go
docker.ListVolumes() ([]Volume, error)
```

### Acesso bruto

```go
docker.RawGet(path string) (json.RawMessage, error)
```

## Exemplos

### Métricas de CPU e memória

```go
stats, err := docker.ContainerStats("meu-container")
if err != nil {
    log.Fatal(err)
}
cpu := docker.CPUPercent(stats)
memMB := float64(stats.MemoryStats.Usage) / 1024 / 1024
fmt.Printf("CPU: %.2f%%  Mem: %.0f MB\n", cpu, memMB)
```

### Criar container e iniciá-lo

```go
resp, err := docker.CreateContainer(docker.ContainerCreateOptions{
    Name:  "meu-nginx",
    Image: "nginx:latest",
    HostConfig: &docker.HostConfig{
        PortBindings: map[string][]docker.PortBinding{
            "80/tcp": {{HostIP: "0.0.0.0", HostPort: "8090"}},
        },
        RestartPolicy: docker.RestartPolicy{Name: "unless-stopped"},
    },
})
if err != nil {
    log.Fatal(err)
}
docker.StartContainer(resp.ID)
```

### Criar rede com subnet customizada

```go
id, err := docker.CreateNetwork(docker.NetworkCreateOptions{
    Name:   "minha-rede",
    Driver: "bridge",
    IPAM: &docker.NetworkIPAM{
        Driver: "default",
        Config: []docker.IPAMConfig{
            {Subnet: "172.28.0.0/16", Gateway: "172.28.0.1"},
        },
    },
})
```

### Conectar container com IP fixo

```go
err := docker.ConnectNetwork("minha-rede", "meu-container", &docker.NetworkConnectOptions{
    EndpointConfig: &docker.EndpointConfig{
        IPAMConfig: &docker.EndpointIPAMConfig{IPv4Address: "172.28.0.10"},
        Aliases:    []string{"backend"},
    },
})
```
