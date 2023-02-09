package nc

import (
	"net/http"
	"strings"
	"time"
)

type Instance struct {
	instanceName string
	baseUrl      string
	client       *http.Client
	userAgent    string
	credentials  AuthCredentials

	credentialUpdateProc func(AuthCredentials)
}

func NewInstance(instanceName string, url string) *Instance {
	return &Instance{
		instanceName:         instanceName,
		baseUrl:              strings.TrimRight(url, "/"),
		client:               http.DefaultClient,
		userAgent:            "Nextcloud Talk Client (GoTalk)",
		credentialUpdateProc: nil,
	}
}

func (i *Instance) GetBaseURL() string {
	return i.baseUrl
}

func (i *Instance) SetCredentials(credentials AuthCredentials) {
	i.credentials = credentials
	if i.credentialUpdateProc != nil {
		i.credentialUpdateProc(credentials)
	}
}

func (i *Instance) OnCredentialsUpdated(credUpdateProc func(AuthCredentials)) {
	i.credentialUpdateProc = credUpdateProc
}

func (i *Instance) GetCredentials() AuthCredentials {
	return i.credentials
}

func (i *Instance) NewLoginFlow() *LoginFlow {
	return &LoginFlow{
		instance:   i,
		updateTime: time.Second * 5,
	}
}
