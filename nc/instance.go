package nc

import (
	"net/http"
)

type Instance struct {
	baseUrl     string
	client      *http.Client
	userAgent   string
	credentials AuthCredentials
}

func NewInstance(url string) *Instance {
	return &Instance{
		baseUrl:   url,
		client:    http.DefaultClient,
		userAgent: "Nextcloud Talk Client (GoTalk)",
	}
}

func (i *Instance) GetBaseURL() string {
	return i.baseUrl
}

func (i *Instance) SetCredentials(credentials AuthCredentials) {
	i.credentials = credentials
}

func (i *Instance) GetCredentials() AuthCredentials {
	return i.credentials
}

func (i *Instance) NewLoginFlow() *LoginFlow {
	return &LoginFlow{
		instance: i,
	}
}
