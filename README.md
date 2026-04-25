# godocker

Cliente para a Docker Engine API escrito em Go puro — sem dependências externas.

Detecta automaticamente o ambiente de execução:

| Ambiente | Transporte |
|---|---|
| `DOCKER_HOST` definido | TCP ou Unix conforme o valor |
| Windows nativo | TCP `localhost:2375` |
| WSL2 | TCP `localhost:2375` |
| Linux nativo | Unix socket `/var/run/docker.sock` |
| Fallback | TCP `localhost:2375` |

## Instalação

```bash
go get github.com/seu-usuario/godocker
```

## Uso básico

```go
import "github.com/seu-usuario/godocker"

// Usa o DefaultClient (detecção automática)
info, err := godocker.Info()
containers, err := godocker.ListContainers(true)
networks, err := godocker.ListNetworks()
```

## Usar cliente customizado

```go
// TCP explícito
client := godocker.NewTCP("localhost:2375")

// Unix socket explícito
client := godocker.NewUnix("/var/run/docker.sock")

// Todos os métodos disponíveis no client
networks, err := client.ListNetworks()
id, err := client.CreateNetwork(godocker.NetworkCreateOptions{Name: "minha-rede"})
```

## API

### Daemon
```go
godocker.Info() (*DockerInfo, error)
```

### Containers
```go
godocker.ListContainers(all bool) ([]Container, error)
godocker.InspectContainer(id string) (*ContainerInspect, error)
godocker.ContainerStats(id string) (*Stats, error)
godocker.ContainerLogs(id string, tail int) (string, error)
godocker.CPUPercent(s *Stats) float64
```

### Networks
```go
godocker.ListNetworks() ([]Network, error)
godocker.InspectNetwork(id string) (*Network, error)
godocker.CreateNetwork(opts NetworkCreateOptions) (string, error)
godocker.RemoveNetwork(id string) error
godocker.ConnectNetwork(networkID, containerID string, opts *NetworkConnectOptions) error
godocker.DisconnectNetwork(networkID, containerID string, force bool) error
godocker.PruneNetworks() ([]string, error)
godocker.NetworkExists(nameOrID string) (bool, error)
godocker.RenameNetwork(oldName, newName string) (string, error)
```

### Images
```go
godocker.ListImages() ([]Image, error)
```

### Volumes
```go
godocker.ListVolumes() ([]Volume, error)
```

## Exemplos

### Criar rede com subnet customizada
```go
id, err := godocker.CreateNetwork(godocker.NetworkCreateOptions{
    Name:   "minha-rede",
    Driver: "bridge",
    IPAM: &godocker.NetworkIPAM{
        Driver: "default",
        Config: []godocker.IPAMConfig{
            {Subnet: "172.28.0.0/16", Gateway: "172.28.0.1"},
        },
    },
})
```

### Conectar container com IP fixo e alias
```go
err := godocker.ConnectNetwork("minha-rede", "meu-container", &godocker.NetworkConnectOptions{
    EndpointConfig: &godocker.EndpointConfig{
        IPAMConfig: &godocker.EndpointIPAMConfig{IPv4Address: "172.28.0.10"},
        Aliases:    []string{"backend"},
    },
})
```

### Stats de CPU e memória
```go
stats, err := godocker.ContainerStats("meu-container")
cpu := godocker.CPUPercent(stats)
memMB := float64(stats.MemoryStats.Usage) / 1024 / 1024
```
