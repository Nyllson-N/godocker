package godocker

// ══════════════════════════════════════════════════════════════════════════════
// CONTAINER
// ══════════════════════════════════════════════════════════════════════════════

type Container struct {
	ID      string            `json:"Id"`
	Names   []string          `json:"Names"`
	Image   string            `json:"Image"`
	State   string            `json:"State"`
	Status  string            `json:"Status"`
	Labels  map[string]string `json:"Labels"`
	Ports   []Port            `json:"Ports"`
	Created int64             `json:"Created"`
}

type Port struct {
	IP          string `json:"IP"`
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
}

type ContainerInspect struct {
	ID      string `json:"Id"`
	Name    string `json:"Name"`
	Created string `json:"Created"`
	State   struct {
		Status     string `json:"Status"`
		Running    bool   `json:"Running"`
		Paused     bool   `json:"Paused"`
		Restarting bool   `json:"Restarting"`
		Pid        int    `json:"Pid"`
		StartedAt  string `json:"StartedAt"`
		FinishedAt string `json:"FinishedAt"`
	} `json:"State"`
	Config struct {
		Image  string            `json:"Image"`
		Env    []string          `json:"Env"`
		Labels map[string]string `json:"Labels"`
	} `json:"Config"`
	NetworkSettings struct {
		Networks map[string]struct {
			IPAddress string `json:"IPAddress"`
			Gateway   string `json:"Gateway"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
	HostConfig struct {
		Memory       int64 `json:"Memory"`
		CPUShares    int64 `json:"CpuShares"`
		RestartPolicy struct {
			Name string `json:"Name"`
		} `json:"RestartPolicy"`
	} `json:"HostConfig"`
}

type Stats struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage  uint64   `json:"total_usage"`
			PercpuUsage []uint64 `json:"percpu_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     int    `json:"online_cpus"`
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
}

// ══════════════════════════════════════════════════════════════════════════════
// IMAGE
// ══════════════════════════════════════════════════════════════════════════════

type Image struct {
	ID       string   `json:"Id"`
	RepoTags []string `json:"RepoTags"`
	Size     int64    `json:"Size"`
	Created  int64    `json:"Created"`
}

// ══════════════════════════════════════════════════════════════════════════════
// VOLUME
// ══════════════════════════════════════════════════════════════════════════════

type Volume struct {
	Name       string            `json:"Name"`
	Driver     string            `json:"Driver"`
	Mountpoint string            `json:"Mountpoint"`
	Labels     map[string]string `json:"Labels"`
}

// ══════════════════════════════════════════════════════════════════════════════
// NETWORK
// ══════════════════════════════════════════════════════════════════════════════

type Network struct {
	ID         string                      `json:"Id"`
	Name       string                      `json:"Name"`
	Driver     string                      `json:"Driver"`
	Scope      string                      `json:"Scope"`
	Internal   bool                        `json:"Internal"`
	Attachable bool                        `json:"Attachable"`
	Ingress    bool                        `json:"Ingress"`
	EnableIPv6 bool                        `json:"EnableIPv6"`
	IPAM       NetworkIPAM                 `json:"IPAM"`
	Options    map[string]string           `json:"Options"`
	Labels     map[string]string           `json:"Labels"`
	Containers map[string]NetworkContainer `json:"Containers"`
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

// NetworkCreateOptions são os parâmetros para criar uma rede.
// Driver padrão: "bridge". Outros: overlay | host | none | macvlan.
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

// NetworkConnectOptions são os parâmetros para conectar um container a uma rede.
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
// DAEMON
// ══════════════════════════════════════════════════════════════════════════════

type DockerInfo struct {
	ID                string `json:"ID"`
	Containers        int    `json:"Containers"`
	ContainersRunning int    `json:"ContainersRunning"`
	ContainersPaused  int    `json:"ContainersPaused"`
	ContainersStopped int    `json:"ContainersStopped"`
	Images            int    `json:"Images"`
	ServerVersion     string `json:"ServerVersion"`
	MemTotal          int64  `json:"MemTotal"`
	NCPU              int    `json:"NCPU"`
	OperatingSystem   string `json:"OperatingSystem"`
	OSType            string `json:"OSType"`
	Architecture      string `json:"Architecture"`
}
