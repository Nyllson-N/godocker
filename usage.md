# godocker — Guia de uso

Cliente Go para a Docker Engine API sem dependências externas.  
Funciona em Linux nativo, WSL2 e Windows (Docker Desktop).

## Instalação

```bash
go get github.com/Nyllson-N/godocker
```

---

## Conexão

### Cliente padrão (automático)

A forma mais simples. O cliente detecta o ambiente automaticamente.

```go
import docker "github.com/Nyllson-N/godocker"

// docker.DefaultClient já está pronto para uso — sem precisar chamar nada.
info, err := docker.Info()
```

A ordem de detecção é:

| Prioridade | Condição                        | Transporte             |
|------------|---------------------------------|------------------------|
| 1          | `DOCKER_HOST` definida          | conforme a variável    |
| 2          | Windows nativo                  | TCP `localhost:2375`   |
| 3          | WSL2 com Docker Desktop ativo   | TCP `localhost:2375`   |
| 4          | Linux com socket disponível     | Unix `/var/run/docker.sock` |
| 5          | Fallback                        | TCP `localhost:2375`   |

### Cliente personalizado

```go
// TCP — Docker remoto ou Docker Desktop exposto
client := docker.NewTCP("192.168.1.100:2375")

// Unix socket customizado
client := docker.NewUnix("/run/user/1000/docker.sock")

// Novo cliente com detecção automática
client := docker.New()

// Sobrescrever a versão da API
client.APIVersion = "v1.45"
```

### Variáveis de ambiente

| Variável              | Efeito                                      |
|-----------------------|---------------------------------------------|
| `DOCKER_HOST`         | Endereço do daemon (`unix://…` ou `tcp://…`) |
| `DOCKER_API_VERSION`  | Versão da API (mínimo `v1.44` aplicado automaticamente) |

---

## Daemon — `/info`

```go
// Via DefaultClient
info, err := docker.Info()

// Via cliente próprio
info, err := client.Info()

fmt.Println(info.ServerVersion)    // "27.0.3"
fmt.Println(info.NCPU)             // 8
fmt.Println(info.MemTotal)         // bytes
fmt.Println(info.ContainersRunning)
fmt.Println(info.StorageDriver)    // "overlay2"
fmt.Println(info.OperatingSystem)
fmt.Println(info.KernelVersion)
fmt.Println(info.Swarm.LocalNodeState) // "inactive" | "active"
```

**Tipo retornado:** `*DockerInfo`

---

## Containers

### Listar

```go
// all=false → só containers em execução
// all=true  → todos (incluindo parados)
containers, err := docker.ListContainers(true)

for _, c := range containers {
    fmt.Println(c.ID[:12])   // ID curto
    fmt.Println(c.Names[0])  // ex: "/meu-container"
    fmt.Println(c.Image)
    fmt.Println(c.State)     // "running" | "exited" | "paused" …
    fmt.Println(c.Status)    // "Up 2 hours"
    fmt.Println(c.Command)
    fmt.Println(c.Created)   // unix timestamp
    
    for _, p := range c.Ports {
        fmt.Printf("%d/%s → %s:%d\n", p.PrivatePort, p.Type, p.IP, p.PublicPort)
    }
    
    for _, m := range c.Mounts {
        fmt.Printf("%s → %s (rw=%v)\n", m.Source, m.Destination, m.RW)
    }
}
```

**Tipo retornado:** `[]Container`

### Inspecionar (dados completos)

```go
inspect, err := docker.InspectContainer("meu-container") // ID ou nome

fmt.Println(inspect.Name)
fmt.Println(inspect.State.Status)    // "running"
fmt.Println(inspect.State.Pid)
fmt.Println(inspect.State.StartedAt)
fmt.Println(inspect.State.OOMKilled)

// Config
fmt.Println(inspect.Config.Image)
fmt.Println(inspect.Config.Env)
fmt.Println(inspect.Config.Cmd)
fmt.Println(inspect.Config.WorkingDir)

// HostConfig
fmt.Println(inspect.HostConfig.NetworkMode)
fmt.Println(inspect.HostConfig.Memory)       // bytes (0 = sem limite)
fmt.Println(inspect.HostConfig.CPUShares)
fmt.Println(inspect.HostConfig.RestartPolicy.Name) // "always" | "no" …
fmt.Println(inspect.HostConfig.Privileged)
fmt.Println(inspect.HostConfig.Binds)        // []string com os volumes montados

// Rede
for netName, ep := range inspect.NetworkSettings.Networks {
    fmt.Printf("%s → IP %s | GW %s\n", netName, ep.IPAddress, ep.Gateway)
}

// Portas expostas
for port, bindings := range inspect.NetworkSettings.Ports {
    for _, b := range bindings {
        fmt.Printf("%s → %s:%s\n", port, b.HostIP, b.HostPort)
    }
}

// Saúde (healthcheck)
if h := inspect.State.Health; h != nil {
    fmt.Println(h.Status) // "healthy" | "unhealthy" | "starting"
    for _, log := range h.Log {
        fmt.Printf("  [%d] %s\n", log.ExitCode, log.Output)
    }
}
```

**Tipo retornado:** `*ContainerInspect`

### Stats de CPU e memória

```go
stats, err := docker.ContainerStats("meu-container")

// Helper pronto para calcular % de CPU
cpu := docker.CPUPercent(stats)
memMB := float64(stats.MemoryStats.Usage) / 1024 / 1024
limMB := float64(stats.MemoryStats.Limit) / 1024 / 1024

fmt.Printf("CPU: %.2f%% | Mem: %.1f MB / %.1f MB\n", cpu, memMB, limMB)

// PIDs
fmt.Println(stats.PidsStats.Current)

// I/O de bloco
for _, e := range stats.BlkioStats.IoServiceBytesRecursive {
    fmt.Printf("blkio %s: %d bytes\n", e.Op, e.Value)
}

// Rede por interface
for iface, n := range stats.Networks {
    fmt.Printf("%s rx=%d tx=%d\n", iface, n.RxBytes, n.TxBytes)
}
```

**Tipo retornado:** `*Stats`

### Logs

```go
// Últimas 50 linhas de stdout + stderr
logs, err := docker.ContainerLogs("meu-container", 50)
fmt.Print(logs)
```

---

## Images

### Listar

```go
images, err := docker.ListImages()

for _, img := range images {
    fmt.Println(img.ID)
    fmt.Println(img.RepoTags)    // ["nginx:latest", "nginx:1.25"]
    fmt.Println(img.RepoDigests)
    fmt.Printf("%.1f MB\n", float64(img.Size)/1024/1024)
    fmt.Println(img.Created)     // unix timestamp
    fmt.Println(img.Containers)  // quantos containers usam esta imagem
}
```

**Tipo retornado:** `[]Image`

---

## Networks

### Listar

```go
networks, err := docker.ListNetworks()

for _, n := range networks {
    fmt.Println(n.ID[:12])
    fmt.Println(n.Name)
    fmt.Println(n.Driver)  // "bridge" | "overlay" | "host" …
    fmt.Println(n.Scope)   // "local" | "swarm"
    if len(n.IPAM.Config) > 0 {
        fmt.Println(n.IPAM.Config[0].Subnet)
        fmt.Println(n.IPAM.Config[0].Gateway)
    }
    for id, c := range n.Containers {
        fmt.Printf("  container %s → %s\n", id[:12], c.IPv4Address)
    }
}
```

### Inspecionar

```go
net, err := docker.InspectNetwork("bridge") // ID ou nome
```

### Criar

```go
// Mínimo
id, err := docker.CreateNetwork(docker.NetworkCreateOptions{
    Name: "minha-rede",
})

// Com subnet customizada e labels
id, err := docker.CreateNetwork(docker.NetworkCreateOptions{
    Name:       "minha-rede",
    Driver:     "bridge",
    Internal:   false,
    Attachable: true,
    Labels:     map[string]string{"env": "prod"},
    IPAM: &docker.NetworkIPAM{
        Driver: "default",
        Config: []docker.IPAMConfig{
            {Subnet: "172.28.0.0/24", Gateway: "172.28.0.1"},
        },
    },
})
```

### Remover

```go
err := docker.RemoveNetwork("minha-rede") // ID ou nome
```

### Conectar container

```go
// IP automático
err := docker.ConnectNetwork("minha-rede", "meu-container", nil)

// IP fixo com alias
err := docker.ConnectNetwork("minha-rede", "meu-container", &docker.NetworkConnectOptions{
    EndpointConfig: &docker.EndpointConfig{
        IPAMConfig: &docker.EndpointIPAMConfig{IPv4Address: "172.28.0.10"},
        Aliases:    []string{"backend"},
    },
})
```

### Desconectar container

```go
err := docker.DisconnectNetwork("minha-rede", "meu-container", false)
// force=true desconecta mesmo com o container rodando
```

### Verificar existência

```go
exists, err := docker.NetworkExists("minha-rede")
```

### Renomear

```go
// Recria a rede com novo nome preservando driver, IPAM e labels.
// Containers conectados precisam ser reconectados manualmente.
newID, err := docker.RenameNetwork("nome-antigo", "nome-novo")
```

### Limpar redes sem uso

```go
removidas, err := docker.PruneNetworks()
fmt.Println("removidas:", removidas)
```

---

## Volumes

### Listar

```go
volumes, err := docker.ListVolumes()

for _, v := range volumes {
    fmt.Println(v.Name)
    fmt.Println(v.Driver)
    fmt.Println(v.Mountpoint)
    fmt.Println(v.Scope)      // "local"
    fmt.Println(v.CreatedAt)
    if v.UsageData != nil {
        fmt.Printf("%.1f MB usado por %d containers\n",
            float64(v.UsageData.Size)/1024/1024, v.UsageData.RefCount)
    }
}
```

**Tipo retornado:** `[]Volume`

---

## API bruta (JSON sem filtros)

Use `RawGet` quando precisar de campos não mapeados nos tipos ou explorar endpoints diretamente.

```go
// Via DefaultClient
raw, err := docker.RawGet("/info")
raw, err := docker.RawGet("/containers/json?all=true")
raw, err := docker.RawGet("/containers/abc123/json")
raw, err := docker.RawGet("/containers/abc123/stats?stream=false")
raw, err := docker.RawGet("/images/json")
raw, err := docker.RawGet("/networks")
raw, err := docker.RawGet("/volumes")

// Via cliente próprio
raw, err := client.RawGet("/version")

// Deserializar manualmente
var result map[string]any
json.Unmarshal(raw, &result)
```

---

## Usando múltiplos clientes

```go
local  := docker.New()
remoto := docker.NewTCP("10.0.0.5:2375")
sock   := docker.NewUnix("/run/user/1000/docker.sock")

// Cada cliente tem seus próprios métodos
localContainers,  _ := local.ListContainers(true)
remoteContainers, _ := remoto.ListContainers(true)
```

---

## Tratamento de erros

Erros da API Docker incluem o código HTTP e a mensagem original:

```go
containers, err := docker.ListContainers(true)
if err != nil {
    // ex: "docker API 404: no such container"
    // ex: "docker [tcp] GET /v1.47/containers/json: connection refused"
    log.Fatal(err)
}
```

---

## Referência rápida

| Função                                              | Descrição                              |
|-----------------------------------------------------|----------------------------------------|
| `docker.Info()`                                     | Info do daemon                         |
| `docker.ListContainers(all)`                        | Lista containers                       |
| `docker.InspectContainer(id)`                       | Inspect completo                       |
| `docker.ContainerStats(id)`                         | CPU, memória, rede, I/O               |
| `docker.ContainerLogs(id, tail)`                    | Últimas N linhas de log               |
| `docker.CPUPercent(stats)`                          | Calcula % de CPU a partir de Stats    |
| `docker.ListImages()`                               | Lista imagens locais                   |
| `docker.ListNetworks()`                             | Lista redes                            |
| `docker.InspectNetwork(id)`                         | Inspect de rede                        |
| `docker.CreateNetwork(opts)`                        | Cria rede, retorna ID                 |
| `docker.RemoveNetwork(id)`                          | Remove rede                            |
| `docker.ConnectNetwork(net, container, opts)`       | Conecta container à rede              |
| `docker.DisconnectNetwork(net, container, force)`   | Desconecta container da rede          |
| `docker.NetworkExists(id)`                          | Verifica se rede existe               |
| `docker.RenameNetwork(old, new)`                    | Renomeia rede                          |
| `docker.PruneNetworks()`                            | Remove redes sem uso                  |
| `docker.ListVolumes()`                              | Lista volumes                          |
| `docker.RawGet(path)`                               | JSON bruto de qualquer endpoint       |
| `docker.New()`                                      | Novo cliente (detecção automática)    |
| `docker.NewTCP(addr)`                               | Novo cliente TCP                       |
| `docker.NewUnix(path)`                              | Novo cliente Unix socket              |
