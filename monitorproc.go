package main

import (
	"GoTalk/nc"
	"log"
	"sync"
	"time"

	"github.com/billgraziano/dpapi"
)

type monitorProcData struct {
	instanceName  string
	ncCredentials nc.AuthCredentials
	ncInstance    *nc.Instance
	ncMonitor     *nc.Monitor
	cache         InstanceCache
	org           OrgInstanceSettings
	user          UserInstanceSettings
}

type LoginFlowResult int64

const (
	LoginFlowWaiting LoginFlowResult = iota
	LoginFlowCanceledOrTimeout
	LoginFlowSuccessful
)

func newMonitorProc(instanceName string) monitorProcData {
	return monitorProcData{
		instanceName: instanceName,
	}
}

func (p *monitorProcData) readCredentials() nc.AuthCredentials {
	var decPassword string
	var err error
	if decPassword, err = dpapi.Decrypt(p.cache.EncryptedAppPassword); err != nil {
		decPassword = ""
	}

	return nc.AuthCredentials{
		LoginName:   p.cache.Username,
		AppPassword: decPassword,
	}
}

func (p *monitorProcData) saveCredentials(cred nc.AuthCredentials) {
	p.cache.Username = cred.LoginName
	p.cache.EncryptedAppPassword = ""

	if cred.AppPassword != "" {
		s, err := dpapi.Encrypt(cred.AppPassword)
		if err != nil {
			return
		}
		p.cache.EncryptedAppPassword = s
	}
}

func (p *monitorProcData) getNotificationSettings() nc.NotificationSettings {
	return p.user.NotificationSettings
}

func (p *monitorProcData) run(wg *sync.WaitGroup, closeChan chan interface{}) {
	defer wg.Done()

	p.cache = cache.InstanceData[p.instanceName]
	p.org = org.InstanceData[p.instanceName]
	p.user = user.InstanceData[p.instanceName]

	defer func() { cache.InstanceData[p.instanceName] = p.cache }()

	p.ncInstance = nc.NewInstance(p.org.InstanceURL)
	p.ncInstance.SetCredentials(p.readCredentials())
	p.ncInstance.OnCredentialsUpdated(p.saveCredentials)

	p.ncMonitor = nc.NewMonitor(p.ncInstance)
	p.ncMonitor.SetNotificationSender(sendMessageNotification)
	p.ncMonitor.SetNotificationSettingsGetter(p.getNotificationSettings)

	messageCheckTime := org.MessageCheckTime
	if messageCheckTime <= 5 {
		messageCheckTime = 5
	}

}

// Returns false, error in case of error.
// Returns false, nil in case that the credentials are still valid.
// Returns true, nil in case that the credentials expired / are invalid.
func (p *monitorProcData) handleLoginSuccessful() (bool, error) {
	if err := p.ncMonitor.ProcessMessages(); err != nil {
		return false, err
	}

	return p.ncInstance.NeedsLogin()
}

func (p *monitorProcData) handleLoginRequired() (LoginFlowResult, error) {

}

// Returns false, error in case of error.
// Returns false, nil in case that the credentials are still valid.
// Returns true, nil in case that the credentials expired / are invalid.
func (p *monitorProcData) handleFirstLoginCheck() (bool, error) {
	return p.ncInstance.NeedsLogin()
}

func runNextcloudMonitor(wg *sync.WaitGroup, closeChan chan interface{}, instanceName string) {

	needsLogin, err := ncMonitor.NeedsLogin()
	if err != nil {
		log.Print(err)
		return
	}

	var rightClickLoginFlow *nc.LoginFlow

	for {
		if needsLogin {
			setInstanceLoginMenuOption(instanceName, func() {
				if rightClickLoginFlow != nil {
					return
				}

				rightClickLoginFlow = ncInstance.NewLoginFlow()
				defer func() { rightClickLoginFlow = nil }()

				if err := rightClickLoginFlow.Start(); err != nil {
					log.Print(err)
					return
				}

				cred, err := rightClickLoginFlow.WaitFlow()
				if err != nil {
					log.Print(err)
					return
				}

				if cred == nil {
					log.Print("Login failed.")
					return
				}

				ncInstance.SetCredentials(*cred)
				needsLogin = false
			})

			switch instanceOrgSettings.Startup {
			case ImmediateLogin:
				loginFlow := ncInstance.NewLoginFlow()

				if err := loginFlow.Start(); err != nil {
					log.Print(err)
					return
				}

				cred, err := loginFlow.WaitFlow()
				if err != nil {
					log.Print(err)
					return
				}

				if cred == nil {
					log.Print("Login failed.")
					return
				}

				ncInstance.SetCredentials(*cred)
				needsLogin = false

			case NotificationLogin:
				loginFlow := ncInstance.NewLoginFlow()

				url, err := loginFlow.StartManual()
				if err != nil {
					log.Print(err)
					return
				}

				go sendMessageNotification(instanceName, "Log In", "Log In to receive Nextcloud notifications", url, true)
				cred, err := loginFlow.WaitFlow()
				if err != nil {
					log.Print(err)
					return
				}

				if cred == nil {
					log.Print("Login failed.")
					return
				}

				ncInstance.SetCredentials(*cred)
				needsLogin = false
			}
		} else {
			setInstanceLoginMenuOption(instanceName, nil)
		}

		select {
		case <-time.After(time.Second * time.Duration(messageCheckTime)):

			if err := ncMonitor.ProcessMessages(); err != nil {
				log.Print(err)
				return
			}

			needsLogin, err = ncMonitor.NeedsLogin()
			if err != nil {
				log.Print(err)
				return
			}

		case _, ok := <-closeChan:
			if !ok {
				return
			}
		}
	}
}
