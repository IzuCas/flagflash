package dto

// System Info
type SystemInfoInput struct{}

type SystemInfoOutput struct {
	Body SystemInfoResponse
}

type SystemInfoResponse struct {
	ID                string `json:"id"`
	Containers        int    `json:"containers"`
	ContainersRunning int    `json:"containersRunning"`
	ContainersPaused  int    `json:"containersPaused"`
	ContainersStopped int    `json:"containersStopped"`
	Images            int    `json:"images"`
	Driver            string `json:"driver"`
	MemoryLimit       bool   `json:"memoryLimit"`
	SwapLimit         bool   `json:"swapLimit"`
	KernelVersion     string `json:"kernelVersion"`
	OperatingSystem   string `json:"operatingSystem"`
	OSType            string `json:"osType"`
	Architecture      string `json:"architecture"`
	NCPU              int    `json:"ncpu"`
	MemTotal          int64  `json:"memTotal"`
	DockerRootDir     string `json:"dockerRootDir"`
	Name              string `json:"name"`
	ServerVersion     string `json:"serverVersion"`
}

// System Version
type SystemVersionInput struct{}

type SystemVersionOutput struct {
	Body VersionResponse
}

type VersionResponse struct {
	Version       string `json:"version"`
	APIVersion    string `json:"apiVersion"`
	MinAPIVersion string `json:"minApiVersion"`
	GitCommit     string `json:"gitCommit"`
	GoVersion     string `json:"goVersion"`
	Os            string `json:"os"`
	Arch          string `json:"arch"`
	KernelVersion string `json:"kernelVersion"`
	BuildTime     string `json:"buildTime"`
}

// System Disk Usage
type SystemDiskUsageInput struct{}

type SystemDiskUsageOutput struct {
	Body DiskUsageResponse
}

type DiskUsageResponse struct {
	LayersSize int64                      `json:"layersSize"`
	Images     []ImageDiskUsageResponse   `json:"images"`
	Containers []ContainerDiskUsageResponse `json:"containers"`
	Volumes    []VolumeDiskUsageResponse  `json:"volumes"`
}

type ImageDiskUsageResponse struct {
	ID         string   `json:"id"`
	RepoTags   []string `json:"repoTags"`
	Created    int64    `json:"created"`
	Size       int64    `json:"size"`
	SharedSize int64    `json:"sharedSize"`
	Containers int64    `json:"containers"`
}

type ContainerDiskUsageResponse struct {
	ID         string   `json:"id"`
	Names      []string `json:"names"`
	Image      string   `json:"image"`
	SizeRw     int64    `json:"sizeRw"`
	SizeRootFs int64    `json:"sizeRootFs"`
	Created    int64    `json:"created"`
	State      string   `json:"state"`
}

type VolumeDiskUsageResponse struct {
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Mountpoint string `json:"mountpoint"`
	Size       int64  `json:"size"`
	RefCount   int64  `json:"refCount"`
}

// System Ping
type SystemPingInput struct{}

type SystemPingOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}
