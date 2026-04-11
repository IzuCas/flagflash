package dto

// Network List
type NetworkListInput struct {
	Filters map[string]string `query:"filters"`
}

type NetworkListOutput struct {
	Body []NetworkSummaryResponse
}

type NetworkSummaryResponse struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Driver         string       `json:"driver"`
	Scope          string       `json:"scope"`
	Internal       bool         `json:"internal"`
	Attachable     bool         `json:"attachable"`
	Ingress        bool         `json:"ingress"`
	Created        string       `json:"created"`
	ContainerCount int          `json:"containerCount"`
	IPAM           IPAMResponse `json:"ipam"`
}

// Network Inspect
type NetworkInspectInput struct {
	ID string `path:"id" doc:"Network ID or name"`
}

type NetworkInspectOutput struct {
	Body NetworkResponse
}

type NetworkResponse struct {
	ID         string                      `json:"id"`
	Name       string                      `json:"name"`
	Created    string                      `json:"created"`
	Scope      string                      `json:"scope"`
	Driver     string                      `json:"driver"`
	EnableIPv6 bool                        `json:"enableIPv6"`
	IPAM       IPAMResponse                `json:"ipam"`
	Internal   bool                        `json:"internal"`
	Attachable bool                        `json:"attachable"`
	Ingress    bool                        `json:"ingress"`
	Options    map[string]string           `json:"options"`
	Labels     map[string]string           `json:"labels"`
	Containers map[string]EndpointResponse `json:"containers"`
}

type IPAMResponse struct {
	Driver  string               `json:"driver"`
	Config  []IPAMConfigResponse `json:"config"`
	Options map[string]string    `json:"options"`
}

type IPAMConfigResponse struct {
	Subnet     string            `json:"subnet"`
	IPRange    string            `json:"ipRange"`
	Gateway    string            `json:"gateway"`
	AuxAddress map[string]string `json:"auxAddress"`
}

type EndpointResponse struct {
	Name        string `json:"name"`
	EndpointID  string `json:"endpointId"`
	MacAddress  string `json:"macAddress"`
	IPv4Address string `json:"ipv4Address"`
	IPv6Address string `json:"ipv6Address"`
}

// Network Create
type NetworkCreateInput struct {
	Body NetworkCreateRequest
}

type NetworkCreateRequest struct {
	Name       string            `json:"name" required:"true"`
	Driver     string            `json:"driver"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
	Ingress    bool              `json:"ingress"`
	EnableIPv6 bool              `json:"enableIPv6"`
	IPAM       *IPAMRequest      `json:"ipam"`
	Options    map[string]string `json:"options"`
	Labels     map[string]string `json:"labels"`
}

type IPAMRequest struct {
	Driver  string              `json:"driver"`
	Config  []IPAMConfigRequest `json:"config"`
	Options map[string]string   `json:"options"`
}

type IPAMConfigRequest struct {
	Subnet     string            `json:"subnet"`
	IPRange    string            `json:"ipRange"`
	Gateway    string            `json:"gateway"`
	AuxAddress map[string]string `json:"auxAddress"`
}

type NetworkCreateOutput struct {
	Body struct {
		ID string `json:"id"`
	}
}

// Network Remove
type NetworkRemoveInput struct {
	ID string `path:"id" doc:"Network ID or name"`
}

type NetworkRemoveOutput struct{}

// Network Connect
type NetworkConnectInput struct {
	ID   string `path:"id" doc:"Network ID or name"`
	Body NetworkConnectRequest
}

type NetworkConnectRequest struct {
	Container      string                 `json:"container" required:"true"`
	EndpointConfig *EndpointConfigRequest `json:"endpointConfig" required:"false"`
}

type EndpointConfigRequest struct {
	Links     []string `json:"links"`
	Aliases   []string `json:"aliases"`
	NetworkID string   `json:"networkId"`
}

type NetworkConnectOutput struct{}

// Network Disconnect
type NetworkDisconnectInput struct {
	ID   string `path:"id" doc:"Network ID or name"`
	Body NetworkDisconnectRequest
}

type NetworkDisconnectRequest struct {
	Container string `json:"container" required:"true"`
	Force     bool   `json:"force"`
}

type NetworkDisconnectOutput struct{}

// Network Prune
type NetworkPruneInput struct{}

type NetworkPruneOutput struct {
	Body NetworkPruneResponse
}

type NetworkPruneResponse struct {
	NetworksDeleted []string `json:"networksDeleted"`
}
