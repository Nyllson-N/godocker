// types.go define todos os tipos de dados que espelham as respostas da Docker Engine API.
// Cada struct corresponde a um objeto JSON retornado pelo Docker.
// As tags `json:"..."` ensinam ao Go como mapear os campos do JSON para os campos da struct.
// A tag `omitempty` faz com que o campo seja omitido no JSON de saída quando estiver vazio.
package godocker

// ══════════════════════════════════════════════════════════════════════════════
// TIPOS COMPARTILHADOS / PRIMITIVOS
// Usados por múltiplos recursos (containers, redes, etc.)
// ══════════════════════════════════════════════════════════════════════════════

// Port representa um mapeamento de porta de um container.
// Exemplo: container porta 80/tcp → host IP 0.0.0.0 porta 8080
type Port struct {
	IP          string `json:"IP,omitempty"`          // IP do host (ex: "0.0.0.0")
	PrivatePort uint16 `json:"PrivatePort"`           // porta dentro do container
	PublicPort  uint16 `json:"PublicPort,omitempty"`  // porta exposta no host (0 se não exposta)
	Type        string `json:"Type"`                  // protocolo: "tcp" ou "udp"
}

// PortBinding mapeia uma porta do container para uma porta específica do host.
// Usado em HostConfig.PortBindings para configurar exposição de portas.
type PortBinding struct {
	HostIP   string `json:"HostIp"`   // IP do host onde a porta será vinculada (ex: "0.0.0.0")
	HostPort string `json:"HostPort"` // porta do host como string (ex: "8080")
}

// MountPoint representa um volume ou bind mount dentro de um container.
// Corresponde a um item de `-v` ou `--mount` no `docker run`.
type MountPoint struct {
	Type        string `json:"Type"`             // "bind", "volume" ou "tmpfs"
	Name        string `json:"Name,omitempty"`   // nome do volume (se Type="volume")
	Source      string `json:"Source"`           // caminho no host ou nome do volume
	Destination string `json:"Destination"`      // caminho dentro do container
	Driver      string `json:"Driver,omitempty"` // driver do volume (ex: "local")
	Mode        string `json:"Mode"`             // modo de montagem (ex: "ro", "rw", "z")
	RW          bool   `json:"RW"`               // true = leitura+escrita, false = somente leitura
	Propagation string `json:"Propagation"`      // propagação de bind: "rprivate", "shared", etc.
}

// RestartPolicy define como o Docker deve reiniciar o container em caso de falha.
type RestartPolicy struct {
	Name              string `json:"Name"`              // "no", "always", "on-failure", "unless-stopped"
	MaximumRetryCount int    `json:"MaximumRetryCount"` // apenas para "on-failure": máximo de tentativas
}

// LogConfig define o driver de log e suas opções para o container.
type LogConfig struct {
	Type   string            `json:"Type"`   // driver: "json-file", "syslog", "none", etc.
	Config map[string]string `json:"Config"` // opções específicas do driver (ex: max-size, max-file)
}

// Ulimit define um limite de recurso do sistema para o container (via ulimit do Linux).
type Ulimit struct {
	Name string `json:"Name"` // nome do limite (ex: "nofile" para arquivos abertos)
	Soft int64  `json:"Soft"` // limite flexível — processo pode ultrapassar até o Hard
	Hard int64  `json:"Hard"` // limite rígido — não pode ser ultrapassado
}

// WeightDevice define o peso de acesso a um dispositivo de bloco específico.
// Usado em HostConfig.BlkioWeightDevice para controle de I/O por dispositivo.
type WeightDevice struct {
	Path   string `json:"Path"`   // caminho do dispositivo (ex: "/dev/sda")
	Weight uint16 `json:"Weight"` // peso relativo de 10 a 1000
}

// ThrottleDevice define limite de taxa de I/O para um dispositivo de bloco.
type ThrottleDevice struct {
	Path string `json:"Path"` // caminho do dispositivo (ex: "/dev/sda")
	Rate uint64 `json:"Rate"` // taxa em bytes/s ou operações/s dependendo do campo pai
}

// DeviceMapping mapeia um dispositivo do host para dentro do container.
type DeviceMapping struct {
	PathOnHost        string `json:"PathOnHost"`        // caminho do dispositivo no host
	PathInContainer   string `json:"PathInContainer"`   // onde aparece dentro do container
	CgroupPermissions string `json:"CgroupPermissions"` // permissões: "r", "w", "m" ou combinações
}

// DeviceRequest é usado para solicitar acesso a dispositivos via CDI (ex: GPU NVIDIA).
type DeviceRequest struct {
	Driver       string            `json:"Driver"`       // driver do dispositivo (ex: "nvidia")
	Count        int               `json:"Count"`        // -1 = todos os dispositivos
	DeviceIDs    []string          `json:"DeviceIDs"`    // IDs específicos dos dispositivos
	Capabilities [][]string        `json:"Capabilities"` // capacidades requeridas (ex: [["gpu"]])
	Options      map[string]string `json:"Options"`      // opções adicionais do driver
}

// Address representa um endereço IP com máscara de sub-rede.
type Address struct {
	Addr      string `json:"Addr"`      // endereço IP (ex: "172.17.0.2")
	PrefixLen int    `json:"PrefixLen"` // tamanho do prefixo CIDR (ex: 16 para /16)
}

// ══════════════════════════════════════════════════════════════════════════════
// CONTAINER LIST  — endpoint: GET /containers/json
// Retornado por ListContainers(). Versão resumida, sem todos os detalhes.
// ══════════════════════════════════════════════════════════════════════════════

// Container é o resumo de um container retornado pela listagem.
// Para dados completos, use InspectContainer() que retorna ContainerInspect.
type Container struct {
	ID              string                     `json:"Id"`              // ID completo do container (64 hex)
	Names           []string                   `json:"Names"`           // nomes com "/" na frente (ex: ["/meu-app"])
	Image           string                     `json:"Image"`           // imagem usada (ex: "nginx:latest")
	ImageID         string                     `json:"ImageID"`         // ID da imagem (sha256:...)
	Command         string                     `json:"Command"`         // comando em execução
	Created         int64                      `json:"Created"`         // timestamp Unix de criação
	Ports           []Port                     `json:"Ports"`           // mapeamentos de porta
	SizeRw          int64                      `json:"SizeRw,omitempty"`     // tamanho da camada gravável (bytes)
	SizeRootFs      int64                      `json:"SizeRootFs,omitempty"` // tamanho total do filesystem (bytes)
	Labels          map[string]string          `json:"Labels"`          // labels personalizados
	State           string                     `json:"State"`           // "running", "exited", "paused", etc.
	Status          string                     `json:"Status"`          // texto legível (ex: "Up 2 hours")
	HostConfig      ContainerHostConfigSummary `json:"HostConfig"`      // resumo da configuração do host
	NetworkSettings ContainerNetworkSummary    `json:"NetworkSettings"` // resumo de redes conectadas
	Mounts          []MountPoint               `json:"Mounts"`          // volumes e bind mounts
}

// ContainerHostConfigSummary é o resumo de configuração de host na listagem.
// Para configuração completa, veja HostConfig em ContainerInspect.
type ContainerHostConfigSummary struct {
	NetworkMode string `json:"NetworkMode"` // modo de rede: "bridge", "host", "none", etc.
}

// ContainerNetworkSummary é o resumo de redes na listagem de containers.
// Para detalhes completos, veja NetworkSettings em ContainerInspect.
type ContainerNetworkSummary struct {
	Networks map[string]*EndpointSettings `json:"Networks"` // mapa: nome_da_rede → configuração
}

// ══════════════════════════════════════════════════════════════════════════════
// CONTAINER INSPECT  — endpoint: GET /containers/{id}/json
// Retornado por InspectContainer(). Versão completa com todos os detalhes.
// ══════════════════════════════════════════════════════════════════════════════

// ContainerInspect contém todos os dados de um container — configuração, estado,
// rede, volumes, healthcheck, filesystem e muito mais.
type ContainerInspect struct {
	ID              string          `json:"Id"`              // ID completo (64 hex)
	Created         string          `json:"Created"`         // data de criação (RFC3339)
	Path            string          `json:"Path"`            // executável do entrypoint
	Args            []string        `json:"Args"`            // argumentos passados ao executável
	State           ContainerState  `json:"State"`           // estado atual (running, exited, etc.)
	Image           string          `json:"Image"`           // ID da imagem (sha256:...)
	ResolvConfPath  string          `json:"ResolvConfPath"`  // caminho do resolv.conf no host
	HostnamePath    string          `json:"HostnamePath"`    // caminho do arquivo hostname no host
	HostsPath       string          `json:"HostsPath"`       // caminho do /etc/hosts no host
	LogPath         string          `json:"LogPath"`         // caminho do arquivo de log no host
	Name            string          `json:"Name"`            // nome com "/" (ex: "/meu-container")
	RestartCount    int             `json:"RestartCount"`    // quantas vezes foi reiniciado
	Driver          string          `json:"Driver"`          // driver de storage (ex: "overlay2")
	Platform        string          `json:"Platform"`        // plataforma (ex: "linux")
	MountLabel      string          `json:"MountLabel"`      // label SELinux para mounts
	ProcessLabel    string          `json:"ProcessLabel"`    // label SELinux para processos
	AppArmorProfile string          `json:"AppArmorProfile"` // perfil AppArmor aplicado
	ExecIDs         []string        `json:"ExecIDs"`         // IDs de exec em andamento
	HostConfig      HostConfig      `json:"HostConfig"`      // configuração completa de runtime
	GraphDriver     GraphDriverData `json:"GraphDriver"`     // dados do driver de armazenamento
	SizeRw          *int64          `json:"SizeRw,omitempty"`     // tamanho da camada gravável
	SizeRootFs      *int64          `json:"SizeRootFs,omitempty"` // tamanho total do filesystem
	Mounts          []MountPoint    `json:"Mounts"`          // volumes e bind mounts ativos
	Config          ContainerConfig `json:"Config"`          // configuração da imagem/container
	NetworkSettings NetworkSettings `json:"NetworkSettings"` // configuração completa de rede
}

// ContainerState descreve o estado atual de execução do container.
type ContainerState struct {
	Status     string  `json:"Status"`     // "created", "running", "paused", "restarting", "exited", "dead"
	Running    bool    `json:"Running"`    // true se está em execução agora
	Paused     bool    `json:"Paused"`     // true se está pausado (docker pause)
	Restarting bool    `json:"Restarting"` // true se está no processo de restart
	OOMKilled  bool    `json:"OOMKilled"`  // true se foi morto por falta de memória
	Dead       bool    `json:"Dead"`       // true se está em estado morto (falha crítica)
	Pid        int     `json:"Pid"`        // PID do processo principal no host (0 se parado)
	ExitCode   int     `json:"ExitCode"`   // código de saída (0=sucesso, outro=erro)
	Error      string  `json:"Error"`      // mensagem de erro se falhou
	StartedAt  string  `json:"StartedAt"`  // data/hora de início (RFC3339)
	FinishedAt string  `json:"FinishedAt"` // data/hora de fim (RFC3339, zero se ainda rodando)
	Health     *Health `json:"Health,omitempty"` // resultado do healthcheck (nil se não configurado)
}

// Health contém o resultado do healthcheck configurado no container.
type Health struct {
	Status        string      `json:"Status"`        // "healthy", "unhealthy", "starting", "none"
	FailingStreak int         `json:"FailingStreak"` // quantas verificações consecutivas falharam
	Log           []HealthLog `json:"Log"`           // histórico das últimas verificações
}

// HealthLog é o resultado de uma execução individual do healthcheck.
type HealthLog struct {
	Start    string `json:"Start"`    // início da verificação (RFC3339)
	End      string `json:"End"`      // fim da verificação (RFC3339)
	ExitCode int    `json:"ExitCode"` // 0=healthy, 1=unhealthy, 2=reserved
	Output   string `json:"Output"`   // saída do comando de verificação
}

// ContainerConfig é a configuração original da imagem e do container.
// Contém o que foi especificado no Dockerfile e no `docker run`.
type ContainerConfig struct {
	Hostname     string              `json:"Hostname"`     // hostname do container
	Domainname   string              `json:"Domainname"`   // domínio do container
	User         string              `json:"User"`         // usuário Unix (ex: "1000", "www-data")
	AttachStdin  bool                `json:"AttachStdin"`  // stdin conectado na criação
	AttachStdout bool                `json:"AttachStdout"` // stdout conectado na criação
	AttachStderr bool                `json:"AttachStderr"` // stderr conectado na criação
	ExposedPorts map[string]struct{} `json:"ExposedPorts"` // portas declaradas (ex: {"80/tcp":{}})
	Tty          bool                `json:"Tty"`          // pseudo-TTY alocado
	OpenStdin    bool                `json:"OpenStdin"`    // stdin mantido aberto
	StdinOnce    bool                `json:"StdinOnce"`    // fecha stdin após o primeiro cliente desconectar
	Env          []string            `json:"Env"`          // variáveis de ambiente (ex: ["PATH=/usr/bin"])
	Cmd          []string            `json:"Cmd"`          // comando padrão do container
	Healthcheck  *HealthcheckConfig  `json:"Healthcheck,omitempty"` // config do healthcheck
	ArgsEscaped  bool                `json:"ArgsEscaped,omitempty"` // args escapados (Windows)
	Image        string              `json:"Image"`        // nome da imagem base
	Volumes      map[string]struct{} `json:"Volumes"`      // volumes declarados (ex: {"/data":{}})
	WorkingDir   string              `json:"WorkingDir"`   // diretório de trabalho (WORKDIR)
	Entrypoint   []string            `json:"Entrypoint"`   // entrypoint do container
	OnBuild      []string            `json:"OnBuild"`      // instruções ONBUILD da imagem
	Labels       map[string]string   `json:"Labels"`       // labels do container
	StopSignal   string              `json:"StopSignal,omitempty"`  // sinal de parada (padrão: SIGTERM)
	StopTimeout  *int                `json:"StopTimeout,omitempty"` // segundos antes do SIGKILL
	Shell        []string            `json:"Shell,omitempty"`       // shell padrão para RUN, CMD, etc.
}

// HealthcheckConfig define como e quando executar o healthcheck do container.
type HealthcheckConfig struct {
	Test        []string `json:"Test"`        // comando: ["CMD", "curl", "-f", "http://localhost"]
	Interval    int64    `json:"Interval"`    // intervalo entre verificações (nanosegundos)
	Timeout     int64    `json:"Timeout"`     // timeout de cada verificação (nanosegundos)
	Retries     int      `json:"Retries"`     // falhas consecutivas antes de marcar unhealthy
	StartPeriod int64    `json:"StartPeriod"` // período inicial de graça (nanosegundos)
}

// GraphDriverData contém informações do driver de armazenamento do container.
type GraphDriverData struct {
	Name string            `json:"Name"` // nome do driver (ex: "overlay2")
	Data map[string]string `json:"Data"` // dados internos do driver (camadas, etc.)
}

// HostConfig contém todas as configurações de runtime do container.
// É tudo que pode ser passado em `docker run` — limites de recursos,
// segurança, rede, volumes, etc.
type HostConfig struct {
	// ── Montagens e arquivos ──────────────────────────────────────────────────
	Binds           []string `json:"Binds"`           // bind mounts: ["host:container:mode"]
	ContainerIDFile string   `json:"ContainerIDFile"` // arquivo para salvar o ID do container
	VolumeDriver    string   `json:"VolumeDriver"`    // driver de volume padrão
	VolumesFrom     []string `json:"VolumesFrom"`     // herda volumes de outros containers

	// ── Log ───────────────────────────────────────────────────────────────────
	LogConfig LogConfig `json:"LogConfig"` // driver e opções de log

	// ── Rede ──────────────────────────────────────────────────────────────────
	NetworkMode     string                   `json:"NetworkMode"`     // "bridge", "host", "none", etc.
	PortBindings    map[string][]PortBinding `json:"PortBindings"`    // mapeamento porta container → host
	PublishAllPorts bool                     `json:"PublishAllPorts"` // expõe todas as portas automaticamente
	DNS             []string                 `json:"Dns"`             // servidores DNS customizados
	DNSOptions      []string                 `json:"DnsOptions"`      // opções do resolv.conf
	DNSSearch       []string                 `json:"DnsSearch"`       // domínios de busca DNS
	ExtraHosts      []string                 `json:"ExtraHosts"`      // entradas /etc/hosts extras

	// ── Restart ───────────────────────────────────────────────────────────────
	RestartPolicy RestartPolicy `json:"RestartPolicy"` // política de restart
	AutoRemove    bool          `json:"AutoRemove"`    // remove container automaticamente ao parar

	// ── Segurança ─────────────────────────────────────────────────────────────
	CapAdd          []string `json:"CapAdd"`          // capabilities Linux adicionadas (ex: ["NET_ADMIN"])
	CapDrop         []string `json:"CapDrop"`         // capabilities Linux removidas
	Privileged      bool     `json:"Privileged"`      // acesso total ao host (perigoso)
	ReadonlyRootfs  bool     `json:"ReadonlyRootfs"`  // filesystem raiz somente leitura
	SecurityOpt     []string `json:"SecurityOpt"`     // opções de segurança (SELinux, AppArmor, seccomp)
	MaskedPaths     []string `json:"MaskedPaths"`     // caminhos ocultados dentro do container
	ReadonlyPaths   []string `json:"ReadonlyPaths"`   // caminhos somente leitura dentro do container

	// ── Namespaces ────────────────────────────────────────────────────────────
	PidMode      string `json:"PidMode"`      // namespace PID: "" (privado) ou "host"
	IpcMode      string `json:"IpcMode"`      // namespace IPC: "private", "host", "shareable"
	UTSMode      string `json:"UTSMode"`      // namespace UTS: "" ou "host"
	UsernsMode   string `json:"UsernsMode"`   // namespace de usuário: "" ou "host"
	CgroupnsMode string `json:"CgroupnsMode"` // namespace de cgroup: "private" ou "host"
	Cgroup       string `json:"Cgroup"`       // cgroup pai customizado

	// ── Comunicação ───────────────────────────────────────────────────────────
	Links    []string `json:"Links"`    // links para outros containers (legado)
	GroupAdd []string `json:"GroupAdd"` // grupos Unix adicionais para o processo

	// ── Memória compartilhada e IPC ───────────────────────────────────────────
	ShmSize int64             `json:"ShmSize"` // tamanho do /dev/shm em bytes
	Tmpfs   map[string]string `json:"Tmpfs,omitempty"` // mountpoints tmpfs: {"/tmp": "size=100m"}

	// ── Sysctls e runtime ─────────────────────────────────────────────────────
	Sysctls     map[string]string `json:"Sysctls,omitempty"`  // parâmetros do kernel (ex: net.ipv4.*)
	Runtime     string            `json:"Runtime"`             // runtime OCI (ex: "runc", "nvidia")
	Isolation   string            `json:"Isolation,omitempty"` // isolamento (principalmente Windows)
	ConsoleSize [2]uint           `json:"ConsoleSize"`         // tamanho do terminal [linhas, colunas]
	Annotations map[string]string `json:"Annotations,omitempty"` // anotações arbitrárias

	// ── Armazenamento ─────────────────────────────────────────────────────────
	StorageOpt map[string]string `json:"StorageOpt,omitempty"` // opções de armazenamento

	// ── Ajuste OOM ────────────────────────────────────────────────────────────
	OomScoreAdj int `json:"OomScoreAdj"` // ajuste do OOM killer: -1000 a 1000

	// ── Limites de recursos ───────────────────────────────────────────────────
	// CPU
	CPUShares         int64  `json:"CpuShares"`         // peso relativo de CPU (padrão: 1024)
	NanoCPUs          int64  `json:"NanoCpus"`          // limite de CPU em nanoCPUs (1e9 = 1 CPU)
	CPUPeriod         int64  `json:"CpuPeriod"`         // período do CFS scheduler (microssegundos)
	CPUQuota          int64  `json:"CpuQuota"`          // quota de CPU no período (microssegundos)
	CPURealtimePeriod int64  `json:"CpuRealtimePeriod"` // período de CPU realtime
	CPURealtimeRuntime int64 `json:"CpuRealtimeRuntime"` // runtime de CPU realtime
	CpusetCpus        string `json:"CpusetCpus"`        // CPUs permitidas (ex: "0-3", "0,2")
	CpusetMems        string `json:"CpusetMems"`        // nós NUMA permitidos
	CPUCount          int64  `json:"CpuCount"`          // número de CPUs (Windows)
	CPUPercent        int64  `json:"CpuPercent"`        // percentual de CPU (Windows)

	// Memória
	Memory            int64  `json:"Memory"`            // limite de memória em bytes (0 = sem limite)
	MemoryReservation int64  `json:"MemoryReservation"` // reserva flexível de memória
	MemorySwap        int64  `json:"MemorySwap"`        // memória + swap (-1 = sem limite)
	MemorySwappiness  *int64 `json:"MemorySwappiness"`  // tendência de swap (0-100, nil = padrão)
	OomKillDisable    *bool  `json:"OomKillDisable"`    // desativa o OOM killer (nil = padrão)
	CgroupParent      string `json:"CgroupParent"`      // cgroup pai (ex: "/docker")

	// I/O de bloco
	BlkioWeight          uint16           `json:"BlkioWeight"`          // peso global de I/O (10-1000)
	BlkioWeightDevice    []WeightDevice   `json:"BlkioWeightDevice"`    // peso por dispositivo
	BlkioDeviceReadBps   []ThrottleDevice `json:"BlkioDeviceReadBps"`   // limite leitura (bytes/s)
	BlkioDeviceWriteBps  []ThrottleDevice `json:"BlkioDeviceWriteBps"`  // limite escrita (bytes/s)
	BlkioDeviceReadIOps  []ThrottleDevice `json:"BlkioDeviceReadIOps"`  // limite leitura (ops/s)
	BlkioDeviceWriteIOps []ThrottleDevice `json:"BlkioDeviceWriteIOps"` // limite escrita (ops/s)

	// Dispositivos
	Devices           []DeviceMapping  `json:"Devices"`           // dispositivos mapeados no container
	DeviceCgroupRules []string         `json:"DeviceCgroupRules"` // regras cgroup para dispositivos
	DeviceRequests    []DeviceRequest  `json:"DeviceRequests"`    // requisições CDI (ex: GPU)

	// PIDs e I/O máximo
	PidsLimit          *int64 `json:"PidsLimit"`          // limite de processos (nil = sem limite)
	Ulimits            []Ulimit `json:"Ulimits"`          // limites do sistema (ulimit)
	IOMaximumIOps      uint64 `json:"IOMaximumIOps"`      // máximo de IOPS (Windows)
	IOMaximumBandwidth uint64 `json:"IOMaximumBandwidth"` // máximo de banda I/O (Windows)
}

// NetworkSettings contém a configuração completa de rede do container após o inspect.
type NetworkSettings struct {
	Bridge                 string                       `json:"Bridge"`                 // nome da bridge padrão
	SandboxID              string                       `json:"SandboxID"`              // ID do sandbox de rede
	HairpinMode            bool                         `json:"HairpinMode"`            // modo hairpin (tráfego local)
	LinkLocalIPv6Address   string                       `json:"LinkLocalIPv6Address"`   // endereço IPv6 link-local
	LinkLocalIPv6PrefixLen int                          `json:"LinkLocalIPv6PrefixLen"` // prefixo do IPv6 link-local
	Ports                  map[string][]PortBinding     `json:"Ports"`                  // portas mapeadas (porta/proto → binding)
	SandboxKey             string                       `json:"SandboxKey"`             // caminho do namespace de rede
	SecondaryIPAddresses   []Address                    `json:"SecondaryIPAddresses"`   // IPs IPv4 secundários
	SecondaryIPv6Addresses []Address                    `json:"SecondaryIPv6Addresses"` // IPs IPv6 secundários
	EndpointID             string                       `json:"EndpointID"`             // ID do endpoint na rede padrão
	Gateway                string                       `json:"Gateway"`                // gateway da rede padrão
	GlobalIPv6Address      string                       `json:"GlobalIPv6Address"`      // endereço IPv6 global
	GlobalIPv6PrefixLen    int                          `json:"GlobalIPv6PrefixLen"`    // prefixo do IPv6 global
	IPAddress              string                       `json:"IPAddress"`              // IP do container na rede padrão
	IPPrefixLen            int                          `json:"IPPrefixLen"`            // tamanho do prefixo da rede
	IPv6Gateway            string                       `json:"IPv6Gateway"`            // gateway IPv6
	MacAddress             string                       `json:"MacAddress"`             // endereço MAC do container
	Networks               map[string]*EndpointSettings `json:"Networks"`               // todas as redes conectadas
}

// EndpointSettings descreve a conexão do container a uma rede específica.
// Um container pode estar conectado a várias redes simultaneamente.
type EndpointSettings struct {
	IPAMConfig          *EndpointIPAMConfig `json:"IPAMConfig"`          // configuração de IP (nil = automático)
	Links               []string            `json:"Links"`               // links para outros containers
	Aliases             []string            `json:"Aliases"`             // nomes alternativos na rede
	NetworkID           string              `json:"NetworkID"`           // ID da rede
	EndpointID          string              `json:"EndpointID"`          // ID do endpoint nesta rede
	Gateway             string              `json:"Gateway"`             // gateway desta rede
	IPAddress           string              `json:"IPAddress"`           // IP do container nesta rede
	IPPrefixLen         int                 `json:"IPPrefixLen"`         // tamanho do prefixo CIDR
	IPv6Gateway         string              `json:"IPv6Gateway"`         // gateway IPv6
	GlobalIPv6Address   string              `json:"GlobalIPv6Address"`   // IP IPv6 global
	GlobalIPv6PrefixLen int                 `json:"GlobalIPv6PrefixLen"` // prefixo IPv6
	MacAddress          string              `json:"MacAddress"`          // endereço MAC nesta rede
	DriverOpts          map[string]string   `json:"DriverOpts"`          // opções do driver de rede
}

// ══════════════════════════════════════════════════════════════════════════════
// CONTAINER CREATE  — endpoint: POST /containers/create
// Usado por CreateContainer(). Define as opções de criação e a resposta.
// ══════════════════════════════════════════════════════════════════════════════

// ContainerCreateOptions define todas as opções para criar um container.
// O campo Name vai como query param na URL (?name=); o restante vai no corpo JSON.
// Apenas Image é obrigatório — todos os outros campos têm valores padrão no Docker.
type ContainerCreateOptions struct {
	Name         string              `json:"-"`                     // nome do container (ex: "meu-nginx") → vai na URL
	Image        string              `json:"Image"`                 // imagem a usar — obrigatório (ex: "nginx:latest")
	Cmd          []string            `json:"Cmd,omitempty"`         // comando a executar (ex: ["nginx", "-g", "daemon off;"])
	Entrypoint   []string            `json:"Entrypoint,omitempty"`  // substitui o ENTRYPOINT da imagem
	Env          []string            `json:"Env,omitempty"`         // variáveis de ambiente no formato "KEY=valor"
	WorkingDir   string              `json:"WorkingDir,omitempty"`  // diretório de trabalho dentro do container
	User         string              `json:"User,omitempty"`        // usuário:grupo para o processo (ex: "1000:1000")
	Hostname     string              `json:"Hostname,omitempty"`    // hostname do container
	Tty          bool                `json:"Tty,omitempty"`         // aloca pseudo-TTY (necessário para shells interativos)
	OpenStdin    bool                `json:"OpenStdin,omitempty"`   // mantém stdin aberto mesmo sem clientes conectados
	Labels       map[string]string   `json:"Labels,omitempty"`      // labels para organização (ex: {"app": "web"})
	ExposedPorts map[string]struct{} `json:"ExposedPorts,omitempty"` // portas a expor (ex: {"80/tcp": {}})
	Volumes      map[string]struct{} `json:"Volumes,omitempty"`     // pontos de montagem declarados (ex: {"/data": {}})
	StopSignal   string              `json:"StopSignal,omitempty"`  // sinal de parada (padrão: "SIGTERM")
	StopTimeout  *int                `json:"StopTimeout,omitempty"` // segundos de espera antes do SIGKILL
	HostConfig   *HostConfig         `json:"HostConfig,omitempty"`  // runtime: portas, volumes, restart, limites, etc.
}

// ContainerCreateResponse é a resposta do Docker ao criar um container com sucesso.
type ContainerCreateResponse struct {
	ID       string   `json:"Id"`       // ID gerado pelo Docker (64 caracteres hexadecimais)
	Warnings []string `json:"Warnings"` // avisos não fatais emitidos pelo Docker (geralmente vazio)
}

// ══════════════════════════════════════════════════════════════════════════════
// CONTAINER STATS  — endpoint: GET /containers/{id}/stats?stream=false
// Retornado por ContainerStats(). Contém métricas de uso de recursos.
// ══════════════════════════════════════════════════════════════════════════════

// Stats é o snapshot de métricas de um container em um momento específico.
// O Docker retorna dois momentos (CPUStats + PreCPUStats) para calcular deltas.
type Stats struct {
	Read         string                       `json:"read"`          // timestamp da leitura atual
	PreRead      string                       `json:"preread"`       // timestamp da leitura anterior
	PidsStats    PidsStats                    `json:"pids_stats"`    // contagem de processos/threads
	BlkioStats   BlkioStats                   `json:"blkio_stats"`   // métricas de I/O de disco
	NumProcs     uint32                       `json:"num_procs"`     // número de processos (Windows)
	StorageStats StorageStats                 `json:"storage_stats"` // métricas de armazenamento (Windows)
	CPUStats     CPUStats                     `json:"cpu_stats"`     // uso de CPU no momento atual
	PreCPUStats  CPUStats                     `json:"precpu_stats"`  // uso de CPU no momento anterior
	MemoryStats  MemoryStats                  `json:"memory_stats"`  // uso de memória
	Name         string                       `json:"name"`          // nome do container
	ID           string                       `json:"id"`            // ID do container
	Networks     map[string]NetworkStatsEntry `json:"networks,omitempty"` // métricas de rede por interface
}

// PidsStats contém informações sobre processos em execução no container.
type PidsStats struct {
	Current uint64 `json:"current"` // número atual de processos/threads
	Limit   uint64 `json:"limit"`   // limite máximo configurado
}

// BlkioStats contém métricas de I/O de disco do container.
// Cada campo é uma lista de operações por dispositivo e tipo (Read/Write).
type BlkioStats struct {
	IoServiceBytesRecursive []BlkioEntry `json:"io_service_bytes_recursive"` // bytes lidos/escritos
	IoServicedRecursive     []BlkioEntry `json:"io_serviced_recursive"`      // número de operações
	IoQueuedRecursive       []BlkioEntry `json:"io_queue_recursive"`         // operações na fila
	IoServiceTimeRecursive  []BlkioEntry `json:"io_service_time_recursive"`  // tempo de serviço
	IoWaitTimeRecursive     []BlkioEntry `json:"io_wait_time_recursive"`     // tempo de espera
	IoMergedRecursive       []BlkioEntry `json:"io_merged_recursive"`        // operações mescladas
	IoTimeRecursive         []BlkioEntry `json:"io_time_recursive"`          // tempo de uso do dispositivo
	SectorsRecursive        []BlkioEntry `json:"sectors_recursive"`          // setores acessados
}

// BlkioEntry é uma entrada de métrica de I/O para um dispositivo específico.
type BlkioEntry struct {
	Major uint64 `json:"major"` // número major do dispositivo (ex: 8 para /dev/sda)
	Minor uint64 `json:"minor"` // número minor do dispositivo (ex: 0 para /dev/sda)
	Op    string `json:"op"`   // operação: "Read", "Write", "Total", "Sync", "Async"
	Value uint64 `json:"value"` // valor da métrica (bytes ou contagem)
}

// StorageStats contém métricas de armazenamento específicas do Windows.
type StorageStats struct {
	ReadCountNormalized  uint64 `json:"read_count_normalized,omitempty"`  // contagem de leituras
	ReadSizeBytes        uint64 `json:"read_size_bytes,omitempty"`        // bytes lidos
	WriteCountNormalized uint64 `json:"write_count_normalized,omitempty"` // contagem de escritas
	WriteSizeBytes       uint64 `json:"write_size_bytes,omitempty"`       // bytes escritos
}

// CPUStats contém as métricas de uso de CPU em um momento específico.
// O Docker fornece dois snapshots (CPUStats e PreCPUStats) para calcular deltas.
type CPUStats struct {
	CPUUsage       CPUUsage       `json:"cpu_usage"`       // contadores de uso de CPU
	SystemCPUUsage uint64         `json:"system_cpu_usage"` // uso total do sistema (todos os processos)
	OnlineCPUs     int            `json:"online_cpus"`      // número de CPUs lógicas disponíveis
	ThrottlingData ThrottlingData `json:"throttling_data"`  // dados de throttling de CPU
}

// CPUUsage contém os contadores de tempo de CPU em nanosegundos.
type CPUUsage struct {
	TotalUsage        uint64   `json:"total_usage"`         // total de tempo de CPU usado (ns)
	PercpuUsage       []uint64 `json:"percpu_usage"`        // tempo por CPU lógica (ns)
	UsageInKernelmode uint64   `json:"usage_in_kernelmode"` // tempo no modo kernel (ns)
	UsageInUsermode   uint64   `json:"usage_in_usermode"`   // tempo no modo usuário (ns)
}

// ThrottlingData contém informações sobre throttling de CPU do container.
// Ocorre quando o container excede sua quota de CPU configurada.
type ThrottlingData struct {
	Periods          uint64 `json:"periods"`           // total de períodos de enforcement
	ThrottledPeriods uint64 `json:"throttled_periods"` // períodos onde houve throttling
	ThrottledTime    uint64 `json:"throttled_time"`    // tempo total throttled (nanosegundos)
}

// MemoryStats contém as métricas de uso de memória do container.
type MemoryStats struct {
	Usage    uint64            `json:"usage"`              // uso atual de memória (bytes)
	MaxUsage uint64            `json:"max_usage,omitempty"` // pico de uso de memória (bytes)
	Stats    map[string]uint64 `json:"stats"`              // métricas detalhadas do cgroup (cache, rss, etc.)
	Failcnt  uint64            `json:"failcnt,omitempty"`  // contagem de falhas de alocação
	Limit    uint64            `json:"limit"`              // limite máximo de memória (bytes)
}

// NetworkStatsEntry contém métricas de rede para uma interface de rede específica.
type NetworkStatsEntry struct {
	RxBytes   uint64 `json:"rx_bytes"`   // bytes recebidos
	RxPackets uint64 `json:"rx_packets"` // pacotes recebidos
	RxErrors  uint64 `json:"rx_errors"`  // erros de recepção
	RxDropped uint64 `json:"rx_dropped"` // pacotes descartados na recepção
	TxBytes   uint64 `json:"tx_bytes"`   // bytes transmitidos
	TxPackets uint64 `json:"tx_packets"` // pacotes transmitidos
	TxErrors  uint64 `json:"tx_errors"`  // erros de transmissão
	TxDropped uint64 `json:"tx_dropped"` // pacotes descartados na transmissão
}

// ══════════════════════════════════════════════════════════════════════════════
// IMAGE  — endpoint: GET /images/json
// Retornado por ListImages().
// ══════════════════════════════════════════════════════════════════════════════

// Image representa uma imagem Docker armazenada localmente.
type Image struct {
	ID          string            `json:"id"`                      // ID da imagem (sha256:...)
	//ParentID    string            `json:"ParentId"`                // ID da imagem pai (vazio para base)
	RepoTags    []string          `json:"RepoTags"`                // tags: ["nginx:latest", "nginx:1.25"]
	//RepoDigests []string          `json:"RepoDigests"`             // digests do registry
	Created     int64             `json:"Created"`                 // timestamp Unix de criação
	Size        int64             `json:"Size"`                    // tamanho real em bytes
	//SharedSize  int64             `json:"SharedSize"`              // tamanho compartilhado com outras imagens
	//VirtualSize int64             `json:"VirtualSize,omitempty"`   // tamanho virtual total
	Labels      map[string]string `json:"Labels"`                  // labels da imagem
	//Containers  int               `json:"Containers"`              // quantos containers usam esta imagem
}

// ══════════════════════════════════════════════════════════════════════════════
// VOLUME  — endpoint: GET /volumes
// Retornado por ListVolumes().
// ══════════════════════════════════════════════════════════════════════════════

// Volume representa um volume Docker para persistência de dados.
// Volumes sobrevivem ao ciclo de vida dos containers.
type Volume struct {
	CreatedAt  string            `json:"CreatedAt,omitempty"` // data de criação (RFC3339)
	Driver     string            `json:"Driver"`              // driver do volume (ex: "local")
	Labels     map[string]string `json:"Labels"`              // labels do volume
	Mountpoint string            `json:"Mountpoint"`          // caminho real no host
	Name       string            `json:"Name"`                // nome do volume
	Options    map[string]string `json:"Options"`             // opções do driver (ex: type, device)
	Scope      string            `json:"Scope"`               // "local" ou "global" (swarm)
	Status     map[string]any    `json:"Status,omitempty"`    // status do driver (depende do driver)
	UsageData  *VolumeUsageData  `json:"UsageData,omitempty"` // dados de uso (nil se não disponível)
}

// VolumeUsageData contém informações de uso do volume.
// Disponível apenas quando o Docker é chamado com --volumes (docker system df -v).
type VolumeUsageData struct {
	RefCount int64 `json:"RefCount"` // número de containers usando este volume
	Size     int64 `json:"Size"`     // tamanho total em bytes (-1 se desconhecido)
}

// VolumeListResponse é a resposta completa do endpoint /volumes.
// O Docker envolve o array em um objeto com warnings.
type VolumeListResponse struct {
	Volumes  []Volume `json:"Volumes"`  // lista de volumes
	Warnings []string `json:"Warnings"` // avisos sobre volumes inacessíveis
}

// ══════════════════════════════════════════════════════════════════════════════
// NETWORK  — endpoint: GET /networks
// Retornado por ListNetworks() e InspectNetwork().
// ══════════════════════════════════════════════════════════════════════════════

// Network representa uma rede Docker para comunicação entre containers.
type Network struct {
	ID         string                      `json:"Id"`         // ID único da rede
	Name       string                      `json:"Name"`       // nome da rede
	Created    string                      `json:"Created"`    // data de criação (RFC3339)
	Scope      string                      `json:"Scope"`      // "local" ou "swarm"
	Driver     string                      `json:"Driver"`     // "bridge", "overlay", "host", "none", etc.
	EnableIPv6 bool                        `json:"EnableIPv6"` // true se IPv6 está habilitado
	IPAM       NetworkIPAM                 `json:"IPAM"`       // configuração de gerenciamento de IPs
	Internal   bool                        `json:"Internal"`   // true = sem acesso externo
	Attachable bool                        `json:"Attachable"` // true = containers podem se conectar manualmente
	Ingress    bool                        `json:"Ingress"`    // true = rede de ingress do Swarm
	ConfigFrom ConfigReference             `json:"ConfigFrom"` // referência de configuração (Swarm)
	ConfigOnly bool                        `json:"ConfigOnly"` // true = apenas configuração, sem instância real
	Containers map[string]NetworkContainer `json:"Containers"` // containers conectados: ID → info
	Options    map[string]string           `json:"Options"`    // opções do driver de rede
	Labels     map[string]string           `json:"Labels"`     // labels da rede
}

// NetworkIPAM define o gerenciamento de endereços IP da rede.
type NetworkIPAM struct {
	Driver  string            `json:"Driver"`  // driver IPAM: "default" ou customizado
	Options map[string]string `json:"Options"` // opções do driver IPAM
	Config  []IPAMConfig      `json:"Config"`  // configurações de subnet (pode haver várias)
}

// IPAMConfig define a configuração de endereçamento de uma subnet da rede.
type IPAMConfig struct {
	Subnet     string            `json:"Subnet"`                     // ex: "172.18.0.0/16"
	Gateway    string            `json:"Gateway"`                    // ex: "172.18.0.1"
	IPRange    string            `json:"IPRange,omitempty"`          // range alocável (ex: "172.18.5.0/24")
	AuxAddress map[string]string `json:"AuxiliaryAddresses,omitempty"` // endereços reservados
}

// NetworkContainer é o resumo de um container conectado a uma rede.
type NetworkContainer struct {
	Name        string `json:"Name"`        // nome do container
	EndpointID  string `json:"EndpointID"`  // ID do endpoint nesta rede
	MacAddress  string `json:"MacAddress"`  // endereço MAC do container nesta rede
	IPv4Address string `json:"IPv4Address"` // IP do container com CIDR (ex: "172.18.0.2/16")
	IPv6Address string `json:"IPv6Address"` // IP IPv6 do container (se configurado)
}

// ConfigReference referencia uma configuração de rede (usado no Swarm).
type ConfigReference struct {
	Network string `json:"Network"` // nome da rede de configuração
}

// ── Tipos de entrada para criação e conexão de redes ─────────────────────────
// Estes tipos são usados como ENTRADA (parâmetros) nas chamadas de criação,
// diferente dos tipos acima que são SAÍDA (respostas do Docker).

// NetworkCreateOptions são os parâmetros para criar uma nova rede.
type NetworkCreateOptions struct {
	Name       string            `json:"Name"`                 // nome da rede (obrigatório)
	Driver     string            `json:"Driver,omitempty"`     // driver (padrão: "bridge")
	Internal   bool              `json:"Internal,omitempty"`   // sem acesso externo
	Attachable bool              `json:"Attachable,omitempty"` // containers podem se conectar manualmente
	Ingress    bool              `json:"Ingress,omitempty"`    // rede de ingress do Swarm
	EnableIPv6 bool              `json:"EnableIPv6,omitempty"` // habilita IPv6
	Options    map[string]string `json:"Options,omitempty"`    // opções do driver
	Labels     map[string]string `json:"Labels,omitempty"`     // labels
	IPAM       *NetworkIPAM      `json:"IPAM,omitempty"`       // configuração de IP (nil = automático)
}

// NetworkConnectOptions são os parâmetros para conectar um container a uma rede.
type NetworkConnectOptions struct {
	Container      string          `json:"Container"`                // ID ou nome do container
	EndpointConfig *EndpointConfig `json:"EndpointConfig,omitempty"` // configuração do endpoint (nil = automático)
}

// EndpointConfig configura o endpoint de um container em uma rede.
type EndpointConfig struct {
	IPAMConfig *EndpointIPAMConfig `json:"IPAMConfig,omitempty"` // IP fixo (nil = automático)
	Aliases    []string            `json:"Aliases,omitempty"`    // nomes alternativos nesta rede
}

// EndpointIPAMConfig permite especificar um IP fixo para o container na rede.
type EndpointIPAMConfig struct {
	IPv4Address string `json:"IPv4Address,omitempty"` // IP fixo (ex: "172.18.0.10")
	IPv6Address string `json:"IPv6Address,omitempty"` // IP IPv6 fixo
}

// ══════════════════════════════════════════════════════════════════════════════
// DAEMON INFO  — endpoint: GET /info
// Retornado por Info(). Contém o estado global do daemon Docker.
// ══════════════════════════════════════════════════════════════════════════════

// DockerInfo contém todas as informações do daemon Docker.
// Equivale à saída completa de `docker info`.
type DockerInfo struct {
	// ── Identificação ─────────────────────────────────────────────────────────
	ID            string `json:"ID"`            // ID único do daemon
	Name          string `json:"Name"`          // hostname da máquina
	ServerVersion string `json:"ServerVersion"` // versão do Docker (ex: "27.0.3")

	// ── Contadores ────────────────────────────────────────────────────────────
	Containers        int `json:"Containers"`        // total de containers
	ContainersRunning int `json:"ContainersRunning"` // containers em execução
	ContainersPaused  int `json:"ContainersPaused"`  // containers pausados
	ContainersStopped int `json:"ContainersStopped"` // containers parados
	Images            int `json:"Images"`            // total de imagens locais

	// ── Sistema ───────────────────────────────────────────────────────────────
	NCPU            int    `json:"NCPU"`            // número de CPUs lógicas
	MemTotal        int64  `json:"MemTotal"`        // memória total em bytes
	KernelVersion   string `json:"KernelVersion"`   // versão do kernel Linux
	OperatingSystem string `json:"OperatingSystem"` // nome do OS (ex: "Ubuntu 22.04.3 LTS")
	OSVersion       string `json:"OSVersion"`       // versão do OS
	OSType          string `json:"OSType"`          // tipo: "linux" ou "windows"
	Architecture    string `json:"Architecture"`    // arquitetura: "x86_64", "aarch64", etc.

	// ── Storage ───────────────────────────────────────────────────────────────
	StorageDriver string     `json:"Driver"`       // driver de storage: "overlay2", "btrfs", etc.
	DriverStatus  [][]string `json:"DriverStatus"` // status detalhado do driver (pares chave-valor)
	DockerRootDir string     `json:"DockerRootDir"` // diretório raiz do Docker (ex: "/var/lib/docker")

	// ── Capacidades do sistema (cgroup) ───────────────────────────────────────
	MemoryLimit     bool `json:"MemoryLimit"`    // suporta limite de memória
	SwapLimit       bool `json:"SwapLimit"`      // suporta limite de swap
	KernelMemoryTCP bool `json:"KernelMemoryTCP,omitempty"` // suporta limite de memória TCP do kernel
	CpuCfsPeriod    bool `json:"CpuCfsPeriod"`   // suporta CFS period (controle fino de CPU)
	CpuCfsQuota     bool `json:"CpuCfsQuota"`    // suporta CFS quota (limite de CPU)
	CPUShares       bool `json:"CPUShares"`       // suporta CPU shares (peso relativo)
	CPUSet          bool `json:"CPUSet"`          // suporta cpuset (fixar em CPU específica)
	PidsLimit       bool `json:"PidsLimit"`       // suporta limite de PIDs
	OomKillDisable  bool `json:"OomKillDisable,omitempty"` // suporta desabilitar OOM killer
	IPv4Forwarding  bool `json:"IPv4Forwarding"`  // encaminhamento IPv4 habilitado
	BridgeNfIptables  bool `json:"BridgeNfIptables"`  // bridge + netfilter/iptables ativo
	BridgeNfIp6tables bool `json:"BridgeNfIp6tables"` // bridge + ip6tables ativo

	// ── Debug e goroutines ────────────────────────────────────────────────────
	Debug       bool `json:"Debug"`       // daemon em modo debug
	NGoroutines int  `json:"NGoroutines"` // goroutines ativas no daemon
	OomScoreAdj int  `json:"OomScoreAdj"` // ajuste OOM do daemon

	// ── Logging ───────────────────────────────────────────────────────────────
	LoggingDriver   string `json:"LoggingDriver"`           // driver de log padrão (ex: "json-file")
	CgroupDriver    string `json:"CgroupDriver"`            // driver de cgroup: "cgroupfs" ou "systemd"
	CgroupVersion   string `json:"CgroupVersion,omitempty"` // versão do cgroup: "1" ou "2"
	NEventsListener int    `json:"NEventsListener"`         // número de listeners de eventos

	// ── Rede e proxy ─────────────────────────────────────────────────────────
	IndexServerAddress string `json:"IndexServerAddress"` // endereço do registry (ex: "https://index.docker.io/v1/")
	HttpProxy          string `json:"HttpProxy"`          // proxy HTTP configurado
	HttpsProxy         string `json:"HttpsProxy"`         // proxy HTTPS configurado
	NoProxy            string `json:"NoProxy"`            // endereços que ignoram proxy

	// ── Labels e configurações ────────────────────────────────────────────────
	Labels            []string           `json:"Labels"`            // labels do daemon (chave=valor)
	ExperimentalBuild bool               `json:"ExperimentalBuild"` // features experimentais ativas
	Runtimes          map[string]Runtime `json:"Runtimes"`          // runtimes disponíveis (runc, nvidia, etc.)
	DefaultRuntime    string             `json:"DefaultRuntime"`    // runtime padrão (normalmente "runc")
	LiveRestoreEnabled bool              `json:"LiveRestoreEnabled"` // containers sobrevivem ao restart do daemon
	Isolation         string             `json:"Isolation,omitempty"` // isolamento (Windows: "hyperv", "process")
	InitBinary        string             `json:"InitBinary"`          // binário init (ex: "docker-init")

	// ── Versões de componentes ────────────────────────────────────────────────
	ContainerdCommit Commit `json:"ContainerdCommit"` // versão do containerd
	RuncCommit       Commit `json:"RuncCommit"`       // versão do runc
	InitCommit       Commit `json:"InitCommit"`       // versão do docker-init

	// ── Segurança e plugins ───────────────────────────────────────────────────
	SecurityOptions     []string      `json:"SecurityOptions"`              // opções ativas (seccomp, apparmor, etc.)
	Plugins             PluginsInfo   `json:"Plugins"`                      // plugins disponíveis por categoria
	ProductLicense      string        `json:"ProductLicense,omitempty"`     // licença (ex: "Community Engine")
	DefaultAddressPools []AddressPool `json:"DefaultAddressPools,omitempty"` // pools de IP padrão para novas redes

	// ── Swarm ─────────────────────────────────────────────────────────────────
	Swarm SwarmInfo `json:"Swarm"` // estado do cluster Swarm (inactive se não configurado)

	// ── Avisos ────────────────────────────────────────────────────────────────
	Warnings []string `json:"Warnings"` // avisos do daemon (ex: swap desabilitado, etc.)
}

// PluginsInfo lista os plugins disponíveis por categoria.
type PluginsInfo struct {
	Volume        []string `json:"Volume"`                // plugins de volume (ex: ["local"])
	Network       []string `json:"Network"`               // plugins de rede (ex: ["bridge","host","overlay"])
	Authorization []string `json:"Authorization,omitempty"` // plugins de autorização
	Log           []string `json:"Log"`                   // plugins de log (ex: ["awslogs","json-file"])
}

// SwarmInfo contém o estado do nó no cluster Docker Swarm.
type SwarmInfo struct {
	NodeID           string       `json:"NodeID"`           // ID do nó neste cluster (vazio se inativo)
	NodeAddr         string       `json:"NodeAddr"`         // endereço anunciado do nó
	LocalNodeState   string       `json:"LocalNodeState"`   // "inactive", "pending", "active", "error", "locked"
	ControlAvailable bool         `json:"ControlAvailable"` // true se este nó é um manager
	Error            string       `json:"Error"`            // mensagem de erro (se houver)
	RemoteManagers   []PeerNode   `json:"RemoteManagers"`   // outros managers conhecidos
	Nodes            int          `json:"Nodes,omitempty"`  // total de nós no cluster
	Managers         int          `json:"Managers,omitempty"` // total de managers no cluster
	Cluster          *ClusterInfo `json:"Cluster,omitempty"` // informações do cluster (nil se não manager)
}

// PeerNode representa outro nó manager conhecido no cluster Swarm.
type PeerNode struct {
	NodeID string `json:"NodeID"` // ID do nó
	Addr   string `json:"Addr"`   // endereço (host:porta)
}

// ClusterInfo contém informações detalhadas do cluster Swarm.
// Disponível apenas em nós manager (ControlAvailable=true).
type ClusterInfo struct {
	ID                     string        `json:"ID"`                     // ID do cluster
	Version                ObjectVersion `json:"Version"`                // versão do objeto Raft
	CreatedAt              string        `json:"CreatedAt"`              // data de criação do cluster
	UpdatedAt              string        `json:"UpdatedAt"`              // data da última atualização
	Spec                   SwarmSpec     `json:"Spec"`                   // especificação do cluster
	RootRotationInProgress bool          `json:"RootRotationInProgress"` // rotação de certificado raiz em progresso
	DefaultAddrPool        []string      `json:"DefaultAddrPool"`        // pools de endereço padrão
	SubnetSize             uint32        `json:"SubnetSize"`             // tamanho da subnet por serviço
	DataPathPort           uint32        `json:"DataPathPort"`           // porta do plano de dados (VXLAN)
}

// ObjectVersion é o número de versão de um objeto Raft no Swarm.
type ObjectVersion struct {
	Index uint64 `json:"Index"` // índice monotonicamente crescente do Raft
}

// SwarmSpec contém a especificação básica do cluster.
type SwarmSpec struct {
	Name string `json:"Name"` // nome do cluster (ex: "default")
}

// Runtime representa um runtime OCI disponível no daemon.
type Runtime struct {
	Path        string   `json:"path,omitempty"`        // caminho do binário do runtime
	RuntimeArgs []string `json:"runtimeArgs,omitempty"` // argumentos extras para o runtime
}

// Commit representa a versão de um componente do Docker (containerd, runc, init).
type Commit struct {
	ID       string `json:"ID"`       // hash do commit atual
	Expected string `json:"Expected"` // hash esperado (para verificação de integridade)
}

// AddressPool define um pool de endereços IP para atribuição automática a redes.
type AddressPool struct {
	Base string `json:"Base"` // rede base em CIDR (ex: "10.0.0.0/8")
	Size int    `json:"Size"` // tamanho de cada subnet alocada (ex: 24 para /24)
}
