package dto

// Container List
type ContainerListInput struct {
	All bool `query:"all" doc:"Return all containers, including stopped ones"`
}

type ContainerListOutput struct {
	Body []ContainerSummaryResponse
}

type ContainerSummaryResponse struct {
	ID      string            `json:"id"`
	Names   []string          `json:"names"`
	Image   string            `json:"image"`
	ImageID string            `json:"imageId"`
	Command string            `json:"command"`
	Created string            `json:"created"`
	State   string            `json:"state"`
	Status  string            `json:"status"`
	Ports   []PortResponse    `json:"ports,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
}

type PortResponse struct {
	IP          string `json:"ip,omitempty"`
	PrivatePort uint16 `json:"privatePort"`
	PublicPort  uint16 `json:"publicPort,omitempty"`
	Type        string `json:"type"`
}

// Container Inspect
type ContainerInspectInput struct {
	ID string `path:"id" doc:"Container ID or name"`
}

type ContainerInspectOutput struct {
	Body ContainerResponse
}

type ContainerResponse struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Image       string                     `json:"image"`
	ImageID     string                     `json:"imageId"`
	Command     string                     `json:"command"`
	Created     string                     `json:"created"`
	State       ContainerStateResponse     `json:"state"`
	Env         []string                   `json:"env"`
	Labels      map[string]string          `json:"labels"`
	Mounts      []MountResponse            `json:"mounts"`
	Healthcheck *HealthcheckConfigResponse `json:"healthcheck,omitempty"`
}

type ContainerStateResponse struct {
	Status     string          `json:"status"`
	Running    bool            `json:"running"`
	Paused     bool            `json:"paused"`
	Restarting bool            `json:"restarting"`
	Dead       bool            `json:"dead"`
	Pid        int             `json:"pid"`
	ExitCode   int             `json:"exitCode"`
	StartedAt  string          `json:"startedAt"`
	FinishedAt string          `json:"finishedAt"`
	Health     *HealthResponse `json:"health,omitempty"`
}

type HealthResponse struct {
	Status        string              `json:"status"`
	FailingStreak int                 `json:"failingStreak"`
	Log           []HealthLogResponse `json:"log,omitempty"`
}

type HealthLogResponse struct {
	Start    string `json:"start"`
	End      string `json:"end"`
	ExitCode int    `json:"exitCode"`
	Output   string `json:"output"`
}

type HealthcheckConfigResponse struct {
	Test        []string `json:"test"`
	Interval    string   `json:"interval"`
	Timeout     string   `json:"timeout"`
	StartPeriod string   `json:"startPeriod"`
	Retries     int      `json:"retries"`
}

type MountResponse struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	RW          bool   `json:"rw"`
}

// Container Create
type ContainerCreateInput struct {
	Body ContainerCreateRequest
}

type ContainerCreateRequest struct {
	Name            string                          `json:"name"`
	Image           string                          `json:"image" required:"true"`
	Cmd             []string                        `json:"cmd"`
	Env             []string                        `json:"env"`
	Labels          map[string]string               `json:"labels"`
	WorkingDir      string                          `json:"workingDir"`
	User            string                          `json:"user"`
	ExposedPorts    map[string]struct{}             `json:"exposedPorts"`
	PortBindings    map[string][]PortBindingRequest `json:"portBindings"`
	Mounts          []MountRequest                  `json:"mounts"`
	NetworkMode     string                          `json:"networkMode"`
	RestartPolicy   string                          `json:"restartPolicy"`
	Memory          int64                           `json:"memory"`
	MemorySwap      int64                           `json:"memorySwap"`
	CPUShares       int64                           `json:"cpuShares"`
	CPUPeriod       int64                           `json:"cpuPeriod"`
	CPUQuota        int64                           `json:"cpuQuota"`
	Privileged      bool                            `json:"privileged"`
	AutoRemove      bool                            `json:"autoRemove"`
	PublishAllPorts bool                            `json:"publishAllPorts"`
}

type PortBindingRequest struct {
	HostIP   string `json:"hostIp"`
	HostPort string `json:"hostPort"`
}

type MountRequest struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	RW          bool   `json:"rw"`
}

type ContainerCreateOutput struct {
	Body struct {
		ID string `json:"id"`
	}
}

// Container Actions
type ContainerActionInput struct {
	ID string `path:"id" doc:"Container ID or name"`
}

type ContainerActionOutput struct{}

type ContainerStopInput struct {
	ID      string `path:"id" doc:"Container ID or name"`
	Timeout int    `query:"timeout" doc:"Seconds to wait before killing"`
}

type ContainerKillInput struct {
	ID     string `path:"id" doc:"Container ID or name"`
	Signal string `query:"signal" doc:"Signal to send (default: SIGKILL)"`
}

type ContainerRemoveInput struct {
	ID            string `path:"id" doc:"Container ID or name"`
	Force         bool   `query:"force" doc:"Force remove even if running"`
	RemoveVolumes bool   `query:"v" doc:"Remove associated volumes"`
}

type ContainerRenameInput struct {
	ID   string `path:"id" doc:"Container ID or name"`
	Name string `query:"name" required:"true" doc:"New container name"`
}

// Container Logs
type ContainerLogsInput struct {
	ID         string `path:"id" doc:"Container ID or name"`
	Stdout     bool   `query:"stdout" default:"true"`
	Stderr     bool   `query:"stderr" default:"true"`
	Since      string `query:"since"`
	Until      string `query:"until"`
	Timestamps bool   `query:"timestamps"`
	Tail       string `query:"tail" default:"all"`
}

type ContainerLogsOutput struct {
	Body struct {
		Logs string `json:"logs"`
	}
}

// Container Stats
type ContainerStatsInput struct {
	ID string `path:"id" doc:"Container ID or name"`
}

type ContainerStatsOutput struct {
	Body ContainerStatsResponse
}

type ContainerStatsResponse struct {
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryUsage   int64   `json:"memoryUsage"`
	MemoryLimit   int64   `json:"memoryLimit"`
	MemoryPercent float64 `json:"memoryPercent"`
	NetworkRx     int64   `json:"networkRx"`
	NetworkTx     int64   `json:"networkTx"`
	BlockRead     int64   `json:"blockRead"`
	BlockWrite    int64   `json:"blockWrite"`
	PIDs          int64   `json:"pids"`
}

// Container Exec
type ContainerExecInput struct {
	ID   string `path:"id" doc:"Container ID or name"`
	Body ExecRequest
}

type ExecRequest struct {
	Cmd             []string `json:"cmd" required:"true"`
	Env             []string `json:"env,omitempty" required:"false"`
	WorkingDir      string   `json:"workingDir,omitempty" required:"false"`
	User            string   `json:"user,omitempty" required:"false"`
	Privileged      bool     `json:"privileged,omitempty" required:"false"`
	Tty             bool     `json:"tty,omitempty" required:"false"`
	ConfirmPassword string   `json:"confirmPassword,omitempty" doc:"Current password — required when privileged is true"`
}

type ContainerExecOutput struct {
	Body ExecResultResponse
}

type ExecResultResponse struct {
	ExitCode int    `json:"exitCode"`
	Output   string `json:"output"`
}

// Container Update Environment
type ContainerUpdateEnvInput struct {
	ID   string `path:"id" doc:"Container ID or name"`
	Body ContainerUpdateEnvRequest
}

type ContainerUpdateEnvRequest struct {
	Env             []string `json:"env" required:"true" doc:"Environment variables in KEY=VALUE format"`
	ConfirmPassword string   `json:"confirmPassword" required:"true" doc:"Current password to confirm identity"`
}

type ContainerUpdateEnvResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type ContainerUpdateEnvOutput struct {
	Body ContainerUpdateEnvResponse
}

// Container Top
type ContainerTopInput struct {
	ID     string `path:"id" doc:"Container ID or name"`
	PsArgs string `query:"ps_args" doc:"Arguments for ps command"`
}

type ContainerTopOutput struct {
	Body ContainerTopResponse
}

type ContainerTopResponse struct {
	Titles    []string   `json:"titles"`
	Processes [][]string `json:"processes"`
}
