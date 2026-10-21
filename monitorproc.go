package main

import (
	"GoTalk/nc"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/pkg/browser"
)

type monitorProcData struct {
	instanceName string
	ncInstance   *nc.Instance
	ncMonitor    *nc.Monitor
	cache        InstanceCache
	org          OrgInstanceSettings
	user         UserInstanceSettings
}

type LoginFlowResult int64

const (
	LoginFlowCanceledOrTimeout LoginFlowResult = iota
	LoginFlowError
	LoginFlowSuccessful
)

type loginFlowOutcome struct {
	LoginFlowResult
	nc.AuthCredentials
}

func newMonitorProc(instanceName string) monitorProcData {
	return monitorProcData{
		instanceName: instanceName,
	}
}

func (p *monitorProcData) readCredentials() nc.AuthCredentials {
	var decPassword string
	var err error
	if decPassword, err = crypto_decrypt(p.cache.EncryptedAppPassword); err != nil {
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
		s, err := crypto_encrypt(cred.AppPassword)
		if err != nil {
			return
		}
		p.cache.EncryptedAppPassword = s
	}

	settingsMutex.Lock()
	defer settingsMutex.Unlock()

	cache.InstanceData[p.instanceName] = p.cache
	saveSettings()
}

func (p *monitorProcData) getNotificationSettings() nc.NotificationSettings {
	settingsMutex.Lock()
	defer settingsMutex.Unlock()

	return user.InstanceData[p.instanceName].NotificationSettings
}

func (p *monitorProcData) sendNotification(instance string, title string, message string, url string, playAudio bool) error {
	return sendMessageNotification(instance, title, message, url, playAudio)
}

// Login requests are shown even if the user disabled message notifications,
// otherwise there would be no way to log in again.
func (p *monitorProcData) sendLoginNotification(url string) error {
	return pushNotification(p.instanceName, p.instanceName, "Login to NextCloud", url, true)
}

func (p *monitorProcData) setLoginMenuOption(callback func()) {
	if p.org.setInstanceLoginMenuOption != nil {
		p.org.setInstanceLoginMenuOption(callback)
	}
}

func (p *monitorProcData) run(wg *sync.WaitGroup, closeChan chan interface{}) {
	defer wg.Done()

	settingsMutex.Lock()
	p.cache = cache.InstanceData[p.instanceName]
	p.org = org.InstanceData[p.instanceName]
	p.user = user.InstanceData[p.instanceName]
	messageCheckTime := org.MessageCheckTime
	settingsMutex.Unlock()

	p.ncInstance = nc.NewInstance(p.instanceName, p.org.InstanceURL)
	p.ncInstance.SetCredentials(p.readCredentials())
	p.ncInstance.OnCredentialsUpdated(p.saveCredentials)

	p.ncMonitor = nc.NewMonitor(p.ncInstance, p.org.NotificationRepeatTime)
	p.ncMonitor.SetNotificationSettingsGetter(p.getNotificationSettings)
	p.ncMonitor.SetNotificationSender(p.sendNotification)

	if !p.ncInstance.IsSecure() {
		log.Printf("%s: the instance URL does not use HTTPS, credentials will be sent unencrypted", p.instanceName)
	}

	if messageCheckTime <= 5 {
		messageCheckTime = 5
	}

	shouldLogin := true

LoginCheck:
	for {
		result, resp, err := p.handleFirstLoginCheck()

		if err != nil {
			log.Print(err)
		}

		switch resp {
		case nc.APIMaintenance:
			// Nextcloud is in maintenance: Wait 5 minutes and retry.
			select {
			case <-time.After(time.Minute * 5):
			case <-closeChan:
				break LoginCheck
			}
		case nc.APIUnreachable:
			// Nextcloud is unreachable: Wait 5 minutes and retry.
			select {
			case <-time.After(time.Minute * 5):
			case <-closeChan:
				break LoginCheck
			}
		case nc.APISuccess:
			// Request successful, but should we login?
			shouldLogin = !(result == nc.CredentialsValid)
			break LoginCheck
		case nc.APILoginExpired:
			// The login expired somehow.
			// This shouldn't happen, but if it does...
			shouldLogin = true
			break LoginCheck
		}
	}

	defer p.setLoginMenuOption(nil)

RunLoop:
	for {
		// First of all, check if the app is quitting.
		select {
		case <-closeChan:
			break RunLoop
		default:
		}

		// Should we login?
		if shouldLogin {
			chanWaitLogin, resp, err := p.handleLoginRequired()

			if err != nil {
				log.Print(err)
			}

			if resp != nc.APISuccess || err != nil {
				// Error trying to start the Login Flow:
				// Retry in a bit.
				select {
				case <-time.After(time.Minute * 5):
				case <-closeChan:
					break RunLoop
				}
				continue
			}

			// Waiting for the login flow to end.
			// Alternative: The app is quitting, get out.
			var result loginFlowOutcome
			select {
			case result = <-chanWaitLogin:
			case <-closeChan:
				break RunLoop
			}
			p.setLoginMenuOption(nil)

			if result.LoginFlowResult == LoginFlowSuccessful {
				shouldLogin = false
				p.ncInstance.SetCredentials(result.AuthCredentials)
			} else if result.LoginFlowResult == LoginFlowError {
				// Don't start a new Login Flow (and possibly open the browser) right away.
				select {
				case <-time.After(time.Second * 30):
				case <-closeChan:
					break RunLoop
				}
			}
		} else {
			// We're logged in, process our request.
			resp, err := p.handleLoginSuccessful()
			if err != nil {
				log.Print(err)
			}

			switch resp {
			case nc.APILoginExpired:
				// We got logged out. Try logging in again, but wait a bit before trying.
				shouldLogin = true
				select {
				case <-time.After(time.Second * 5):
				case <-closeChan:
					break RunLoop
				}
			case nc.APIMaintenance:
				// Nextcloud is in maintenance: Wait 5 minutes and retry.
				select {
				case <-time.After(time.Minute * 5):
				case <-closeChan:
					break RunLoop
				}
			case nc.APIUnreachable:
				// Nextcloud is unreachable: Wait a bit and retry.
				select {
				case <-time.After(time.Second * 20):
				case <-closeChan:
					break RunLoop
				}
			case nc.APISuccess:
				// API Request successful: Wait some time before running another request.
				select {
				case <-time.After(time.Second * time.Duration(messageCheckTime)):
				case <-closeChan:
					break RunLoop
				}
			}
		}
	}
}

// Returns false, error in case of error.
// Returns false, nil in case that the credentials are still valid.
// Returns true, nil/error case that the credentials expired / are invalid.
func (p *monitorProcData) handleLoginSuccessful() (nc.APIResponse, error) {
	status, err := p.ncMonitor.ProcessMessages()

	if status != nc.APISuccess || err != nil {
		return status, err
	}

	// Placeholder: Do something?

	return nc.APISuccess, nil
}

// Waits for a started login flow to end.
func waitLoginFlow(loginFlow *nc.LoginFlow) loginFlowOutcome {
	resp, res, cred, err := loginFlow.Wait()
	if err != nil {
		log.Print(err)
	}

	if resp != nc.APISuccess || err != nil {
		return loginFlowOutcome{LoginFlowError, nc.AuthCredentials{}}
	}

	if res == nc.LoginSuccessful && cred != nil {
		return loginFlowOutcome{LoginFlowSuccessful, *cred}
	}

	// Login Flow Canceled / Timeout
	return loginFlowOutcome{LoginFlowCanceledOrTimeout, nc.AuthCredentials{}}
}

func (p *monitorProcData) openLoginPage(url string) error {
	if err := browser.OpenURL(url); err != nil {
		// Could not open the browser: Send a notification.
		return p.sendLoginNotification(url)
	}
	return nil
}

// Runs a whole login flow, started from the context menu.
func (p *monitorProcData) runContextMenuLogin() loginFlowOutcome {
	loginFlow := p.ncInstance.NewLoginFlow()

	resp, url, err := loginFlow.Start()
	if resp != nc.APISuccess || err != nil {
		if err != nil {
			log.Print(err)
		}
		return loginFlowOutcome{LoginFlowError, nc.AuthCredentials{}}
	}

	if err = p.openLoginPage(url); err != nil {
		log.Print(err)
		loginFlow.Cancel()
		return loginFlowOutcome{LoginFlowError, nc.AuthCredentials{}}
	}

	return waitLoginFlow(loginFlow)
}

// The returned channel receives the outcome of the login flow,
// and is only valid if the function returns nc.APISuccess and no error.
func (p *monitorProcData) handleLoginRequired() (chan loginFlowOutcome, nc.APIResponse, error) {
	// Buffered, so that the outcome can be delivered even if nobody is waiting for it anymore.
	chanLoginFlow := make(chan loginFlowOutcome, 1)

	switch p.org.Login {
	case LoginImmediately, LoginWithNotification:
		loginFlow := p.ncInstance.NewLoginFlow()

		resp, url, err := loginFlow.Start()
		if resp != nc.APISuccess || err != nil {
			return nil, resp, err
		}

		if p.org.Login == LoginImmediately {
			err = p.openLoginPage(url)
		} else {
			err = p.sendLoginNotification(url)
		}
		if err != nil {
			loginFlow.Cancel()
			return nil, resp, err
		}

		go func() {
			chanLoginFlow <- waitLoginFlow(loginFlow)
		}()

		return chanLoginFlow, nc.APISuccess, nil

	case LoginWithContextMenu:
		var loginStarted sync.Once
		p.setLoginMenuOption(func() {
			// Menu actions run on the UI goroutine: never block it while the user logs in,
			// and only ever start one login flow.
			loginStarted.Do(func() {
				go func() {
					chanLoginFlow <- p.runContextMenuLogin()
				}()
			})
		})
		return chanLoginFlow, nc.APISuccess, nil
	}

	return nil, nc.APIUnreachable, errors.New("invalid login mode")
}

// Returns false, error in case of error.
// Returns false, nil in case that the credentials are still valid.
// Returns true, nil in case that the credentials expired / are invalid.
func (p *monitorProcData) handleFirstLoginCheck() (nc.CredentialValidationResult, nc.APIResponse, error) {
	result, resp, err := p.ncInstance.ValidateCredentials(p.ncInstance.GetCredentials())
	if result != nc.CredentialsValid || resp != nc.APISuccess || err != nil {
		return result, resp, err
	}

	// Placeholder: Do something?

	return nc.CredentialsValid, nc.APISuccess, nil
}
