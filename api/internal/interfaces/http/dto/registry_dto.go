package dto

// Registry Login
type RegistryLoginInput struct {
	Body RegistryLoginRequest
}

type RegistryLoginRequest struct {
	Username      string `json:"username" doc:"Registry username" minLength:"1"`
	Password      string `json:"password" doc:"Registry password" minLength:"1"`
	ServerAddress string `json:"serverAddress,omitempty" doc:"Registry server address (default: Docker Hub)"`
}

type RegistryLoginOutput struct {
	Body RegistryLoginResponse
}

type RegistryLoginResponse struct {
	Status        string `json:"status" doc:"Login status message"`
	IdentityToken string `json:"identityToken,omitempty" doc:"Identity token if provided"`
}

// Registry Logout
type RegistryLogoutInput struct {
	Body RegistryLogoutRequest
}

type RegistryLogoutRequest struct {
	ServerAddress string `json:"serverAddress,omitempty" doc:"Registry server address (default: Docker Hub)"`
}

type RegistryLogoutOutput struct {
	Body StatusResponse
}

type StatusResponse struct {
	Status string `json:"status" doc:"Operation status"`
}

// Proxy Configuration
type ProxyGetInput struct{}

type ProxyGetOutput struct {
	Body ProxyConfigResponse
}

type ProxyConfigResponse struct {
	HTTPProxy  string `json:"httpProxy" doc:"HTTP proxy URL"`
	HTTPSProxy string `json:"httpsProxy" doc:"HTTPS proxy URL"`
	NoProxy    string `json:"noProxy" doc:"Comma-separated list of hosts to exclude from proxy"`
	FTPProxy   string `json:"ftpProxy,omitempty" doc:"FTP proxy URL"`
}

type ProxySetInput struct {
	Body ProxySetRequest
}

type ProxySetRequest struct {
	HTTPProxy  *string `json:"httpProxy,omitempty" doc:"HTTP proxy URL"`
	HTTPSProxy *string `json:"httpsProxy,omitempty" doc:"HTTPS proxy URL"`
	NoProxy    *string `json:"noProxy,omitempty" doc:"Comma-separated list of hosts to exclude from proxy"`
	FTPProxy   *string `json:"ftpProxy,omitempty" doc:"FTP proxy URL"`
}

type ProxySetOutput struct {
	Body StatusResponse
}

// Settings Info
type SettingsInfoInput struct{}

type SettingsInfoOutput struct {
	Body SettingsInfoResponse
}

type SettingsInfoResponse struct {
	Registries []RegistryInfoResponse `json:"registries" doc:"Configured registries"`
	Proxy      ProxyConfigResponse    `json:"proxy" doc:"Current proxy configuration"`
}

type RegistryInfoResponse struct {
	ServerAddress string `json:"serverAddress" doc:"Registry server address"`
	Username      string `json:"username" doc:"Logged in username"`
	IsLoggedIn    bool   `json:"isLoggedIn" doc:"Whether currently logged in"`
}
