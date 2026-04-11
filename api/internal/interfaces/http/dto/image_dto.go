package dto

// Image List
type ImageListInput struct {
	All     bool              `query:"all" doc:"Return all images"`
	Filters map[string]string `query:"filters"`
}

type ImageListOutput struct {
	Body []ImageSummaryResponse
}

type ImageSummaryResponse struct {
	ID          string            `json:"id"`
	RepoTags    []string          `json:"repoTags"`
	RepoDigests []string          `json:"repoDigests"`
	Created     string            `json:"created"`
	Size        int64             `json:"size"`
	VirtualSize int64             `json:"virtualSize"`
	Labels      map[string]string `json:"labels"`
}

// Image Inspect
type ImageInspectInput struct {
	ID string `path:"id" doc:"Image ID or name"`
}

type ImageInspectOutput struct {
	Body ImageInspectResponse
}

type ImageInspectResponse struct {
	ID            string              `json:"id"`
	RepoTags      []string            `json:"repoTags"`
	RepoDigests   []string            `json:"repoDigests"`
	Created       string              `json:"created"`
	Size          int64               `json:"size"`
	VirtualSize   int64               `json:"virtualSize"`
	Author        string              `json:"author"`
	Config        ImageConfigResponse `json:"config"`
	Architecture  string              `json:"architecture"`
	Os            string              `json:"os"`
	DockerVersion string              `json:"dockerVersion"`
}

type ImageConfigResponse struct {
	Hostname     string              `json:"hostname"`
	User         string              `json:"user"`
	Env          []string            `json:"env"`
	Cmd          []string            `json:"cmd"`
	WorkingDir   string              `json:"workingDir"`
	Entrypoint   []string            `json:"entrypoint"`
	Labels       map[string]string   `json:"labels"`
	ExposedPorts map[string]struct{} `json:"exposedPorts"`
}

// Image Pull
type ImagePullInput struct {
	Body ImagePullRequest
}

type ImagePullRequest struct {
	Image    string  `json:"image" required:"true"`
	Tag      *string `json:"tag,omitempty"`
	Platform *string `json:"platform,omitempty"`
}

type ImagePullOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

// Image Remove
type ImageRemoveInput struct {
	ID            string `path:"id" doc:"Image ID or name"`
	Force         bool   `query:"force"`
	PruneChildren bool   `query:"prune"`
}

type ImageRemoveOutput struct {
	Body ImageRemoveResponse
}

type ImageRemoveResponse struct {
	Deleted  []string `json:"deleted"`
	Untagged []string `json:"untagged"`
}

// Image Tag
type ImageTagInput struct {
	ID   string `path:"id" doc:"Image ID or name"`
	Body ImageTagRequest
}

type ImageTagRequest struct {
	Repo string `json:"repo" required:"true"`
	Tag  string `json:"tag"`
}

type ImageTagOutput struct{}

// Image History
type ImageHistoryInput struct {
	ID string `path:"id" doc:"Image ID or name"`
}

type ImageHistoryOutput struct {
	Body []ImageHistoryItem
}

type ImageHistoryItem struct {
	ID        string   `json:"id"`
	Created   string   `json:"created"`
	CreatedBy string   `json:"createdBy"`
	Size      int64    `json:"size"`
	Comment   string   `json:"comment"`
	Tags      []string `json:"tags"`
}

// Image Search
type ImageSearchInput struct {
	Term  string `query:"term" required:"true" doc:"Search term"`
	Limit int    `query:"limit" default:"25"`
}

type ImageSearchOutput struct {
	Body []ImageSearchResult
}

type ImageSearchResult struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	StarCount   int    `json:"starCount"`
	IsOfficial  bool   `json:"isOfficial"`
	IsAutomated bool   `json:"isAutomated"`
}

// Image Prune
type ImagePruneInput struct {
	All bool `query:"all" doc:"Remove all unused images"`
}

type ImagePruneOutput struct {
	Body ImagePruneResponse
}

type ImagePruneResponse struct {
	ImagesDeleted  []string `json:"imagesDeleted"`
	SpaceReclaimed int64    `json:"spaceReclaimed"`
}
