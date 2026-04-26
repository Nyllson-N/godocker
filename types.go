package godocker

// ══════════════════════════════════════════════════════════════════════════════
// SHARED / PRIMITIVOS
// ══════════════════════════════════════════════════════════════════════════════

type Port struct {
	IP          string `json:"IP,omitempty"`
	PrivatePort uint16 `json:"PrivatePort"`
	PublicPort  uint16 `json:"PublicPort,omitempty"`
	Type        string `json:"Type"`
}

// PortBinding representa um bind de porta host ↔ container.
type PortBinding struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

// MountPoint é um ponto de montagem dentro de um container.
type MountPoint struct {
	Type        string `json:"Type"`
	Name        string `json:"Name,omitempty"`
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	Driver      string `json:"Driver,omitempty"`
	Mode        string `json:"Mode"`
	RW          bool   `json:"RW"`
	Propagation string `json:"Propagation"`
}

type RestartPolicy struct {
	Name              string `json:"Name"`
	MaximumRetryCount int    `json:"MaximumRetryCount"`
}

type LogConfig struct {
	Type   string            `json:"Type"`
	Config map[string]string `json:"Config"`
}

type Ulimit struct {
	Name string `json:"Name"`
	Soft int64  `json:"Soft"`
	Hard int64  `json:"Hard"`
}

type WeightDevice struct {
	Path   string `json:"Path"`
	Weight uint16 `json:"Weight"`
}

type ThrottleDevice struct {
	Path string `json:"Path"`
	Rate uint64 `json:"Rate"`
}

type DeviceMapping struct {
	PathOnHost        string `json:"PathOnHost"`
	PathInContainer   string `json:"PathInContainer"`
	CgroupPermissions string `json:"CgroupPermissions"`
}

type DeviceRequest struct {
	Driver       string            `json:"Driver"`
	Count        int               `json:"Count"`
	DeviceIDs    []string          `json:"DeviceIDs"`
	Capabilities [][]string        `json:"Capabilities"`
	Options      map[string]string `json:"Options"`
}

type Address struct {
	Addr      string `json:"Addr"`
	PrefixLen int    `json:"PrefixLen"`
}

// ══════════════════════════════════════════════════════════════════════════════
// CONTAINER LIST  (/containers/json)
// ══════════════════════════════════════════════════════════════════════════════

type Container struct {
	ID              string                     `json:"Id"`
	Names           []string                   `json:"Names"`
	Image           string                     `json:"Image"`
	ImageID         string                     `json:"ImageID"`
	Command         string                     `json:"Command"`
	Created         int64                      `json:"Created"`
	Ports           []Port                     `json:"Ports"`
	SizeRw          int64                      `json:"SizeRw,omitempty"`
	SizeRootFs      int64                      `json:"SizeRootFs,omitempty"`
	Labels          map[string]string          `json:"Labels"`
	State           string                     `json:"State"`
	Status          string                     `json:"Status"`
	HostConfig      ContainerHostConfigSummary `json:"HostConfig"`
	NetworkSettings ContainerNetworkSummary    `json:"NetworkSettings"`
	Mounts          []MountPoint               `json:"Mounts"`
}

type ContainerHostConfigSummary struct {
	NetworkMode string `json:"NetworkMode"`
}

type ContainerNetworkSummary struct {
	Networks map[string]*EndpointSettings `json:"Networks"`
}

// ══════════════════════════════════════════════════════════════════════════════
// CONTAINER INSPECT  (/containers/{id}/json)
// ══════════════════════════════════════════════════════════════════════════════

type ContainerInspect struct {
	ID              string          `json:"Id"`
	Created         string          `json:"Created"`
	Path            string          `json:"Path"`
	Args            []string        `json:"Args"`
	State           ContainerState  `json:"State"`
	Image           string          `json:"Image"`
	ResolvConfPath  string          `json:"ResolvConfPath"`
	HostnamePath    string          `json:"HostnamePath"`
	HostsPath       string          `json:"HostsPath"`
	LogPath         string          `json:"LogPath"`
	Name            string          `json:"Name"`
	RestartCount    int             `json:"RestartCount"`
	Driver          string          `json:"Driver"`
	Platform        string          `json:"Platform"`
	MountLabel      string          `json:"MountLabel"`
	ProcessLabel    string          `json:"ProcessLabel"`
	AppArmorProfile string          `json:"AppArmorProfile"`
	ExecIDs         []string        `json:"ExecIDs"`
	HostConfig      HostConfig      `json:"HostConfig"`
	GraphDriver     GraphDriverData `json:"GraphDriver"`
	SizeRw          *int64          `json:"SizeRw,omitempty"`
	SizeRootFs      *int64          `json:"SizeRootFs,omitempty"`
	Mounts          []MountPoint    `json:"Mounts"`
	Config          ContainerConfig `json:"Config"`
	NetworkSettings NetworkSettings `json:"NetworkSettings"`
}

type ContainerState struct {
	Status     string  `json:"Status"`
	Running    bool    `json:"Running"`
	Paused     bool    `json:"Paused"`
	Restarting bool    `json:"Restarting"`
	OOMKilled  bool    `json:"OOMKilled"`
	Dead       bool    `json:"Dead"`
	Pid        int     `json:"Pid"`
	ExitCode   int     `json:"ExitCode"`
	Error      string  `json:"Error"`
	StartedAt  string  `json:"StartedAt"`
	FinishedAt string  `json:"FinishedAt"`
	Health     *Health `json:"Health,omitempty"`
}

type Health struct {
	Status        string      `json:"Status"`
	FailingStreak int         `json:"FailingStreak"`
	Log           []HealthLog `json:"Log"`
}

type HealthLog struct {
	Start    string `json:"Start"`
	End      string `json:"End"`
	ExitCode int    `json:"ExitCode"`
	Output   string `json:"Output"`
}

type ContainerConfig struct {
	Hostname     string              `json:"Hostname"`
	Domainname   string              `json:"Domainname"`
	User         string              `json:"User"`
	AttachStdin  bool                `json:"AttachStdin"`
	AttachStdout bool                `json:"AttachStdout"`
	AttachStderr bool                `json:"AttachStderr"`
	ExposedPorts map[string]struct{} `json:"ExposedPorts"`
	Tty          bool                `json:"Tty"`
	OpenStdin    bool                `json:"OpenStdin"`
	StdinOnce    bool                `json:"StdinOnce"`
	Env          []string            `json:"Env"`
	Cmd          []string            `json:"Cmd"`
	Healthcheck  *HealthcheckConfig  `json:"Healthcheck,omitempty"`
	ArgsEscaped  bool                `json:"ArgsEscaped,omitempty"`
	Image        string              `json:"Image"`
	Volumes      map[string]struct{} `json:"Volumes"`
	WorkingDir   string              `json:"WorkingDir"`
	Entrypoint   []string            `json:"Entrypoint"`
	OnBuild      []string            `json:"OnBuild"`
	Labels       map[string]string   `json:"Labels"`
	StopSignal   string              `json:"StopSignal,omitempty"`
	StopTimeout  *int                `json:"StopTimeout,omitempty"`
	Shell        []string            `json:"Shell,omitempty"`
}

type HealthcheckConfig struct {
	Test        []string `json:"Test"`
	Interval    int64    `json:"Interval"`
	Timeout     int64    `json:"Timeout"`
	Retries     int      `json:"Retries"`
	StartPeriod int64    `json:"StartPeriod"`
}

type GraphDriverData struct {
	Name string            `json:"Name"`
	Data map[string]string `json:"Data"`
}

// HostConfig contém todas as configurações de runtime do container.
type HostConfig struct {
	Binds           []string                   `json:"Binds"`
	ContainerIDFile string                     `json:"ContainerIDFile"`
	LogConfig       LogConfig                  `json:"LogConfig"`
	NetworkMode     string                     `json:"NetworkMode"`
	PortBindings    map[string][]PortBinding   `json:"PortBindings"`
	RestartPolicy   RestartPolicy              `json:"RestartPolicy"`
	AutoRemove      bool                       `json:"AutoRemove"`
	VolumeDriver    string                     `json:"VolumeDriver"`
	VolumesFrom     []string                   `json:"VolumesFrom"`
	ConsoleSize     [2]uint                    `json:"ConsoleSize"`
	Annotations     map[string]string          `json:"Annotations,omitempty"`
	CapAdd          []string                   `json:"CapAdd"`
	CapDrop         []string                   `json:"CapDrop"`
	CgroupnsMode    string                     `json:"CgroupnsMode"`
	DNS             []string                   `json:"Dns"`
	DNSOptions      []string                   `json:"DnsOptions"`
	DNSSearch       []string                   `json:"DnsSearch"`
	ExtraHosts      []string                   `json:"ExtraHosts"`
	GroupAdd        []string                   `json:"GroupAdd"`
	IpcMode         string                     `json:"IpcMode"`
	Cgroup          string                     `json:"Cgroup"`
	Links           []string                   `json:"Links"`
	OomScoreAdj     int                        `json:"OomScoreAdj"`
	PidMode         string                     `json:"PidMode"`
	Privileged      bool                       `json:"Privileged"`
	PublishAllPorts bool                       `json:"PublishAllPorts"`
	ReadonlyRootfs  bool                       `json:"ReadonlyRootfs"`
	SecurityOpt     []string                   `json:"SecurityOpt"`
	StorageOpt      map[string]string          `json:"StorageOpt,omitempty"`
	Tmpfs           map[string]string          `json:"Tmpfs,omitempty"`
	UTSMode         string                     `json:"UTSMode"`
	UsernsMode      string                     `json:"UsernsMode"`
	ShmSize         int64                      `json:"ShmSize"`
	Sysctls         map[string]string          `json:"Sysctls,omitempty"`
	Runtime         string                     `json:"Runtime"`
	Isolation       string                     `json:"Isolation,omitempty"`
	MaskedPaths     []string                   `json:"MaskedPaths"`
	ReadonlyPaths   []string                   `json:"ReadonlyPaths"`
	// Resource limits
	CPUShares            int64          `json:"CpuShares"`
	Memory               int64          `json:"Memory"`
	NanoCPUs             int64          `json:"NanoCpus"`
	CgroupParent         string         `json:"CgroupParent"`
	BlkioWeight          uint16         `json:"BlkioWeight"`
	BlkioWeightDevice    []WeightDevice `json:"BlkioWeightDevice"`
	BlkioDeviceReadBps   []ThrottleDevice `json:"BlkioDeviceReadBps"`
	BlkioDeviceWriteBps  []ThrottleDevice `json:"BlkioDeviceWriteBps"`
	BlkioDeviceReadIOps  []ThrottleDevice `json:"BlkioDeviceReadIOps"`
	BlkioDeviceWriteIOps []ThrottleDevice `json:"BlkioDeviceWriteIOps"`
	CPUPeriod            int64          `json:"CpuPeriod"`
	CPUQuota             int64          `json:"CpuQuota"`
	CPURealtimePeriod    int64          `json:"CpuRealtimePeriod"`
	CPURealtimeRuntime   int64          `json:"CpuRealtimeRuntime"`
	CpusetCpus           string         `json:"CpusetCpus"`
	CpusetMems           string         `json:"CpusetMems"`
	Devices              []DeviceMapping  `json:"Devices"`
	DeviceCgroupRules    []string         `json:"DeviceCgroupRules"`
	DeviceRequests       []DeviceRequest  `json:"DeviceRequests"`
	MemoryReservation    int64          `json:"MemoryReservation"`
	MemorySwap           int64          `json:"MemorySwap"`
	MemorySwappiness     *int64         `json:"MemorySwappiness"`
	OomKillDisable       *bool          `json:"OomKillDisable"`
	PidsLimit            *int64         `json:"PidsLimit"`
	Ulimits              []Ulimit       `json:"Ulimits"`
	CPUCount             int64          `json:"CpuCount"`
	CPUPercent           int64          `json:"CpuPercent"`
	IOMaximumIOps        uint64         `json:"IOMaximumIOps"`
	IOMaximumBandwidth   uint64         `json:"IOMaximumBandwidth"`
}

// NetworkSettings contém a configuração completa de rede do container.
type NetworkSettings struct {
	Bridge                 string                       `json:"Bridge"`
	SandboxID              string                       `json:"SandboxID"`
	HairpinMode            bool                         `json:"HairpinMode"`
	LinkLocalIPv6Address   string                       `json:"LinkLocalIPv6Address"`
	LinkLocalIPv6PrefixLen int                          `json:"LinkLocalIPv6PrefixLen"`
	Ports                  map[string][]PortBinding     `json:"Ports"`
	SandboxKey             string                       `json:"SandboxKey"`
	SecondaryIPAddresses   []Address                    `json:"SecondaryIPAddresses"`
	SecondaryIPv6Addresses []Address                    `json:"SecondaryIPv6Addresses"`
	EndpointID             string                       `json:"EndpointID"`
	Gateway                string                       `json:"Gateway"`
	GlobalIPv6Address      string                       `json:"GlobalIPv6Address"`
	GlobalIPv6PrefixLen    int                          `json:"GlobalIPv6PrefixLen"`
	IPAddress              string                       `json:"IPAddress"`
	IPPrefixLen            int                          `json:"IPPrefixLen"`
	IPv6Gateway            string                       `json:"IPv6Gateway"`
	MacAddress             string                       `json:"MacAddress"`
	Networks               map[string]*EndpointSettings `json:"Networks"`
}

// EndpointSettings descreve a conexão de um container a uma rede específica.
type EndpointSettings struct {
	IPAMConfig          *EndpointIPAMConfig `json:"IPAMConfig"`
	Links               []string            `json:"Links"`
	Aliases             []string            `json:"Aliases"`
	NetworkID           string              `json:"NetworkID"`
	EndpointID          string              `json:"EndpointID"`
	Gateway             string              `json:"Gateway"`
	IPAddress           string              `json:"IPAddress"`
	IPPrefixLen         int                 `json:"IPPrefixLen"`
	IPv6Gateway         string              `json:"IPv6Gateway"`
	GlobalIPv6Address   string              `json:"GlobalIPv6Address"`
	GlobalIPv6PrefixLen int                 `json:"GlobalIPv6PrefixLen"`
	MacAddress          string              `json:"MacAddress"`
	DriverOpts          map[string]string   `json:"DriverOpts"`
}

// ══════════════════════════════════════════════════════════════════════════════
// CONTAINER STATS  (/containers/{id}/stats?stream=false)
// ══════════════════════════════════════════════════════════════════════════════

type Stats struct {
	Read         string                       `json:"read"`
	PreRead      string                       `json:"preread"`
	PidsStats    PidsStats                    `json:"pids_stats"`
	BlkioStats   BlkioStats                   `json:"blkio_stats"`
	NumProcs     uint32                       `json:"num_procs"`
	StorageStats StorageStats                 `json:"storage_stats"`
	CPUStats     CPUStats                     `json:"cpu_stats"`
	PreCPUStats  CPUStats                     `json:"precpu_stats"`
	MemoryStats  MemoryStats                  `json:"memory_stats"`
	Name         string                       `json:"name"`
	ID           string                       `json:"id"`
	Networks     map[string]NetworkStatsEntry `json:"networks,omitempty"`
}

type PidsStats struct {
	Current uint64 `json:"current"`
	Limit   uint64 `json:"limit"`
}

type BlkioStats struct {
	IoServiceBytesRecursive []BlkioEntry `json:"io_service_bytes_recursive"`
	IoServicedRecursive     []BlkioEntry `json:"io_serviced_recursive"`
	IoQueuedRecursive       []BlkioEntry `json:"io_queue_recursive"`
	IoServiceTimeRecursive  []BlkioEntry `json:"io_service_time_recursive"`
	IoWaitTimeRecursive     []BlkioEntry `json:"io_wait_time_recursive"`
	IoMergedRecursive       []BlkioEntry `json:"io_merged_recursive"`
	IoTimeRecursive         []BlkioEntry `json:"io_time_recursive"`
	SectorsRecursive        []BlkioEntry `json:"sectors_recursive"`
}

type BlkioEntry struct {
	Major uint64 `json:"major"`
	Minor uint64 `json:"minor"`
	Op    string `json:"op"`
	Value uint64 `json:"value"`
}

// StorageStats é usado apenas no Windows.
type StorageStats struct {
	ReadCountNormalized  uint64 `json:"read_count_normalized,omitempty"`
	ReadSizeBytes        uint64 `json:"read_size_bytes,omitempty"`
	WriteCountNormalized uint64 `json:"write_count_normalized,omitempty"`
	WriteSizeBytes       uint64 `json:"write_size_bytes,omitempty"`
}

type CPUStats struct {
	CPUUsage       CPUUsage       `json:"cpu_usage"`
	SystemCPUUsage uint64         `json:"system_cpu_usage"`
	OnlineCPUs     int            `json:"online_cpus"`
	ThrottlingData ThrottlingData `json:"throttling_data"`
}

type CPUUsage struct {
	TotalUsage        uint64   `json:"total_usage"`
	PercpuUsage       []uint64 `json:"percpu_usage"`
	UsageInKernelmode uint64   `json:"usage_in_kernelmode"`
	UsageInUsermode   uint64   `json:"usage_in_usermode"`
}

type ThrottlingData struct {
	Periods          uint64 `json:"periods"`
	ThrottledPeriods uint64 `json:"throttled_periods"`
	ThrottledTime    uint64 `json:"throttled_time"`
}

type MemoryStats struct {
	Usage    uint64            `json:"usage"`
	MaxUsage uint64            `json:"max_usage,omitempty"`
	Stats    map[string]uint64 `json:"stats"`
	Failcnt  uint64            `json:"failcnt,omitempty"`
	Limit    uint64            `json:"limit"`
}

type NetworkStatsEntry struct {
	RxBytes   uint64 `json:"rx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	RxErrors  uint64 `json:"rx_errors"`
	RxDropped uint64 `json:"rx_dropped"`
	TxBytes   uint64 `json:"tx_bytes"`
	TxPackets uint64 `json:"tx_packets"`
	TxErrors  uint64 `json:"tx_errors"`
	TxDropped uint64 `json:"tx_dropped"`
}

// ══════════════════════════════════════════════════════════════════════════════
// IMAGE  (/images/json)
// ══════════════════════════════════════════════════════════════════════════════

type Image struct {
	ID          string            `json:"Id"`
	ParentID    string            `json:"ParentId"`
	RepoTags    []string          `json:"RepoTags"`
	RepoDigests []string          `json:"RepoDigests"`
	Created     int64             `json:"Created"`
	Size        int64             `json:"Size"`
	SharedSize  int64             `json:"SharedSize"`
	VirtualSize int64             `json:"VirtualSize,omitempty"`
	Labels      map[string]string `json:"Labels"`
	Containers  int               `json:"Containers"`
}

// ══════════════════════════════════════════════════════════════════════════════
// VOLUME  (/volumes)
// ══════════════════════════════════════════════════════════════════════════════

type Volume struct {
	CreatedAt  string            `json:"CreatedAt,omitempty"`
	Driver     string            `json:"Driver"`
	Labels     map[string]string `json:"Labels"`
	Mountpoint string            `json:"Mountpoint"`
	Name       string            `json:"Name"`
	Options    map[string]string `json:"Options"`
	Scope      string            `json:"Scope"`
	Status     map[string]any    `json:"Status,omitempty"`
	UsageData  *VolumeUsageData  `json:"UsageData,omitempty"`
}

type VolumeUsageData struct {
	RefCount int64 `json:"RefCount"`
	Size     int64 `json:"Size"`
}

type VolumeListResponse struct {
	Volumes  []Volume `json:"Volumes"`
	Warnings []string `json:"Warnings"`
}

// ══════════════════════════════════════════════════════════════════════════════
// NETWORK  (/networks)
// ══════════════════════════════════════════════════════════════════════════════

type Network struct {
	ID         string                      `json:"Id"`
	Name       string                      `json:"Name"`
	Created    string                      `json:"Created"`
	Scope      string                      `json:"Scope"`
	Driver     string                      `json:"Driver"`
	EnableIPv6 bool                        `json:"EnableIPv6"`
	IPAM       NetworkIPAM                 `json:"IPAM"`
	Internal   bool                        `json:"Internal"`
	Attachable bool                        `json:"Attachable"`
	Ingress    bool                        `json:"Ingress"`
	ConfigFrom ConfigReference             `json:"ConfigFrom"`
	ConfigOnly bool                        `json:"ConfigOnly"`
	Containers map[string]NetworkContainer `json:"Containers"`
	Options    map[string]string           `json:"Options"`
	Labels     map[string]string           `json:"Labels"`
}

type NetworkIPAM struct {
	Driver  string            `json:"Driver"`
	Options map[string]string `json:"Options"`
	Config  []IPAMConfig      `json:"Config"`
}

type IPAMConfig struct {
	Subnet     string            `json:"Subnet"`
	Gateway    string            `json:"Gateway"`
	IPRange    string            `json:"IPRange,omitempty"`
	AuxAddress map[string]string `json:"AuxiliaryAddresses,omitempty"`
}

type NetworkContainer struct {
	Name        string `json:"Name"`
	EndpointID  string `json:"EndpointID"`
	MacAddress  string `json:"MacAddress"`
	IPv4Address string `json:"IPv4Address"`
	IPv6Address string `json:"IPv6Address"`
}

type ConfigReference struct {
	Network string `json:"Network"`
}

// ── Tipos de entrada (input) para criação/conexão de redes ───────────────────

type NetworkCreateOptions struct {
	Name       string            `json:"Name"`
	Driver     string            `json:"Driver,omitempty"`
	Internal   bool              `json:"Internal,omitempty"`
	Attachable bool              `json:"Attachable,omitempty"`
	Ingress    bool              `json:"Ingress,omitempty"`
	EnableIPv6 bool              `json:"EnableIPv6,omitempty"`
	Options    map[string]string `json:"Options,omitempty"`
	Labels     map[string]string `json:"Labels,omitempty"`
	IPAM       *NetworkIPAM      `json:"IPAM,omitempty"`
}

type NetworkConnectOptions struct {
	Container      string          `json:"Container"`
	EndpointConfig *EndpointConfig `json:"EndpointConfig,omitempty"`
}

type EndpointConfig struct {
	IPAMConfig *EndpointIPAMConfig `json:"IPAMConfig,omitempty"`
	Aliases    []string            `json:"Aliases,omitempty"`
}

type EndpointIPAMConfig struct {
	IPv4Address string `json:"IPv4Address,omitempty"`
	IPv6Address string `json:"IPv6Address,omitempty"`
}

// ══════════════════════════════════════════════════════════════════════════════
// DAEMON INFO  (/info)
// ══════════════════════════════════════════════════════════════════════════════

type DockerInfo struct {
	ID                  string            `json:"ID"`
	Containers          int               `json:"Containers"`
	ContainersRunning   int               `json:"ContainersRunning"`
	ContainersPaused    int               `json:"ContainersPaused"`
	ContainersStopped   int               `json:"ContainersStopped"`
	Images              int               `json:"Images"`
	StorageDriver       string            `json:"Driver"`
	DriverStatus        [][]string        `json:"DriverStatus"`
	DockerRootDir       string            `json:"DockerRootDir"`
	MemoryLimit         bool              `json:"MemoryLimit"`
	SwapLimit           bool              `json:"SwapLimit"`
	KernelMemoryTCP     bool              `json:"KernelMemoryTCP,omitempty"`
	CpuCfsPeriod        bool              `json:"CpuCfsPeriod"`
	CpuCfsQuota         bool              `json:"CpuCfsQuota"`
	CPUShares           bool              `json:"CPUShares"`
	CPUSet              bool              `json:"CPUSet"`
	PidsLimit           bool              `json:"PidsLimit"`
	OomKillDisable      bool              `json:"OomKillDisable,omitempty"`
	IPv4Forwarding      bool              `json:"IPv4Forwarding"`
	BridgeNfIptables    bool              `json:"BridgeNfIptables"`
	BridgeNfIp6tables   bool              `json:"BridgeNfIp6tables"`
	Debug               bool              `json:"Debug"`
	OomScoreAdj         int               `json:"OomScoreAdj"`
	NGoroutines         int               `json:"NGoroutines"`
	LoggingDriver       string            `json:"LoggingDriver"`
	CgroupDriver        string            `json:"CgroupDriver"`
	CgroupVersion       string            `json:"CgroupVersion,omitempty"`
	NEventsListener     int               `json:"NEventsListener"`
	KernelVersion       string            `json:"KernelVersion"`
	OperatingSystem     string            `json:"OperatingSystem"`
	OSVersion           string            `json:"OSVersion"`
	OSType              string            `json:"OSType"`
	Architecture        string            `json:"Architecture"`
	NCPU                int               `json:"NCPU"`
	MemTotal            int64             `json:"MemTotal"`
	IndexServerAddress  string            `json:"IndexServerAddress"`
	HttpProxy           string            `json:"HttpProxy"`
	HttpsProxy          string            `json:"HttpsProxy"`
	NoProxy             string            `json:"NoProxy"`
	Name                string            `json:"Name"`
	Labels              []string          `json:"Labels"`
	ExperimentalBuild   bool              `json:"ExperimentalBuild"`
	ServerVersion       string            `json:"ServerVersion"`
	Runtimes            map[string]Runtime `json:"Runtimes"`
	DefaultRuntime      string            `json:"DefaultRuntime"`
	Swarm               SwarmInfo         `json:"Swarm"`
	LiveRestoreEnabled  bool              `json:"LiveRestoreEnabled"`
	Isolation           string            `json:"Isolation,omitempty"`
	InitBinary          string            `json:"InitBinary"`
	ContainerdCommit    Commit            `json:"ContainerdCommit"`
	RuncCommit          Commit            `json:"RuncCommit"`
	InitCommit          Commit            `json:"InitCommit"`
	SecurityOptions     []string          `json:"SecurityOptions"`
	Plugins             PluginsInfo       `json:"Plugins"`
	ProductLicense      string            `json:"ProductLicense,omitempty"`
	DefaultAddressPools []AddressPool     `json:"DefaultAddressPools,omitempty"`
	Warnings            []string          `json:"Warnings"`
}

type PluginsInfo struct {
	Volume        []string `json:"Volume"`
	Network       []string `json:"Network"`
	Authorization []string `json:"Authorization,omitempty"`
	Log           []string `json:"Log"`
}

type SwarmInfo struct {
	NodeID           string       `json:"NodeID"`
	NodeAddr         string       `json:"NodeAddr"`
	LocalNodeState   string       `json:"LocalNodeState"`
	ControlAvailable bool         `json:"ControlAvailable"`
	Error            string       `json:"Error"`
	RemoteManagers   []PeerNode   `json:"RemoteManagers"`
	Nodes            int          `json:"Nodes,omitempty"`
	Managers         int          `json:"Managers,omitempty"`
	Cluster          *ClusterInfo `json:"Cluster,omitempty"`
}

type PeerNode struct {
	NodeID string `json:"NodeID"`
	Addr   string `json:"Addr"`
}

type ClusterInfo struct {
	ID                     string        `json:"ID"`
	Version                ObjectVersion `json:"Version"`
	CreatedAt              string        `json:"CreatedAt"`
	UpdatedAt              string        `json:"UpdatedAt"`
	Spec                   SwarmSpec     `json:"Spec"`
	RootRotationInProgress bool          `json:"RootRotationInProgress"`
	DefaultAddrPool        []string      `json:"DefaultAddrPool"`
	SubnetSize             uint32        `json:"SubnetSize"`
	DataPathPort           uint32        `json:"DataPathPort"`
}

type ObjectVersion struct {
	Index uint64 `json:"Index"`
}

type SwarmSpec struct {
	Name string `json:"Name"`
}

type Runtime struct {
	Path        string   `json:"path,omitempty"`
	RuntimeArgs []string `json:"runtimeArgs,omitempty"`
}

type Commit struct {
	ID       string `json:"ID"`
	Expected string `json:"Expected"`
}

type AddressPool struct {
	Base string `json:"Base"`
	Size int    `json:"Size"`
}
