package dto

// Volume List
type VolumeListInput struct {
	Filters map[string]string `query:"filters"`
}

type VolumeListOutput struct {
	Body []VolumeResponse
}

type VolumeResponse struct {
	Name       string               `json:"name"`
	Driver     string               `json:"driver"`
	Mountpoint string               `json:"mountpoint"`
	CreatedAt  string               `json:"createdAt"`
	Labels     map[string]string    `json:"labels"`
	Scope      string               `json:"scope"`
	Options    map[string]string    `json:"options"`
	UsageData  *VolumeUsageResponse `json:"usageData,omitempty"`
}

type VolumeUsageResponse struct {
	Size     int64 `json:"size"`
	RefCount int64 `json:"refCount"`
}

// Volume Inspect
type VolumeInspectInput struct {
	Name string `path:"name" doc:"Volume name"`
}

type VolumeInspectOutput struct {
	Body VolumeResponse
}

// Volume Create
type VolumeCreateInput struct {
	Body VolumeCreateRequest
}

type VolumeCreateRequest struct {
	Name       string            `json:"name" required:"true"`
	Driver     string            `json:"driver"`
	DriverOpts map[string]string `json:"driverOpts"`
	Labels     map[string]string `json:"labels"`
}

type VolumeCreateOutput struct {
	Body VolumeResponse
}

// Volume Remove
type VolumeRemoveInput struct {
	Name  string `path:"name" doc:"Volume name"`
	Force bool   `query:"force"`
}

type VolumeRemoveOutput struct{}

// Volume Prune
type VolumePruneInput struct{}

type VolumePruneOutput struct {
	Body VolumePruneResponse
}

type VolumePruneResponse struct {
	VolumesDeleted []string `json:"volumesDeleted"`
	SpaceReclaimed uint64   `json:"spaceReclaimed"`
}
