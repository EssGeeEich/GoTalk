package nc

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckServerURL(t *testing.T) {
	secure := NewInstance("test", "https://cloud.example.org/")
	insecure := NewInstance("test", "http://cloud.example.org/")

	cases := []struct {
		instance *Instance
		url      string
		valid    bool
	}{
		{secure, "https://cloud.example.org/index.php/login/v2/flow/abc", true},
		{secure, "http://cloud.example.org/index.php/login/v2/flow/abc", false},
		{secure, "file:///C:/Windows/System32/calc.exe", false},
		{secure, `C:\Windows\System32\calc.exe`, false},
		{secure, "javascript:alert(1)", false},
		{secure, "/relative/path", false},
		{insecure, "http://cloud.example.org/index.php/login/v2/flow/abc", true},
		{insecure, "https://cloud.example.org/index.php/login/v2/flow/abc", true},
	}

	for _, c := range cases {
		err := c.instance.checkServerURL(c.url)
		if (err == nil) != c.valid {
			t.Errorf("checkServerURL(%q) on %q: got error %v, expected valid=%v", c.url, c.instance.baseUrl, err, c.valid)
		}
	}
}

func TestRedirectDowngradeIsRefused(t *testing.T) {
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the request must not reach the HTTP server")
	}))
	defer plain.Close()

	secure := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, plain.URL, http.StatusFound)
	}))
	defer secure.Close()

	client := newHTTPClient()
	client.Transport = secure.Client().Transport

	if _, err := client.Get(secure.URL); err == nil {
		t.Error("expected the redirect to be refused")
	}
}

func TestLoginFlowInvalidPollResponse(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/index.php/login/v2":
			w.Write([]byte(`{"poll":{"token":"tok","endpoint":"` + server.URL + `/poll"},"login":"` + server.URL + `/flow"}`))
		case "/poll":
			if err := r.ParseForm(); err != nil || r.PostForm.Get("token") != "tok" {
				t.Errorf("unexpected poll request: %v %v", r.PostForm, err)
			}
			w.Write([]byte(`not json`))
		}
	}))
	defer server.Close()

	instance := NewInstance("test", server.URL)
	flow := instance.NewLoginFlow()
	flow.updateTime = time.Millisecond * 10

	resp, url, err := flow.Start()
	if resp != APISuccess || err != nil || url != server.URL+"/flow" {
		t.Fatalf("unexpected start result: %v %q %v", resp, url, err)
	}

	resp, result, cred, err := flow.Wait()
	if resp == APISuccess || result == LoginSuccessful || cred != nil || err == nil {
		t.Errorf("an invalid response must not be a successful login: %v %v %v %v", resp, result, cred, err)
	}

	// Canceling after the flow ended must not panic.
	flow.Cancel()
}

func TestLoginFlowRejectsUnsafeLoginURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"poll":{"token":"tok","endpoint":"http://example.org/poll"},"login":"file:///C:/Windows/System32/calc.exe"}`))
	}))
	defer server.Close()

	resp, _, err := NewInstance("test", server.URL).NewLoginFlow().Start()
	if resp == APISuccess || err == nil {
		t.Errorf("expected the login URL to be rejected, got %v %v", resp, err)
	}
}

func TestGetUserConversationsWithoutLastMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ocs":{"meta":{"status":"ok"},"data":[
			{"id":1,"token":"aaa","displayName":"Empty","lastMessage":[]},
			{"id":2,"token":"bbb","displayName":"Changed type","sipEnabled":"yes","lastMessage":{"id":5,"message":"hi {user}","messageParameters":{"user":{"name":"Bob"}}}},
			{"id":3,"token":"ccc","displayName":"Normal","lastMessage":{"id":7,"message":"hello","messageParameters":[]}}
		]}}`))
	}))
	defer server.Close()

	conversations, resp, err := NewInstance("test", server.URL).GetUserConversations()
	if resp != APISuccess || err != nil {
		t.Fatalf("unexpected result: %v %v", resp, err)
	}

	if len(*conversations) != 3 {
		t.Fatalf("expected 3 conversations, got %d", len(*conversations))
	}

	if msg := (*conversations)[1].LastMessage; msg.Id != 5 || msg.format() != "hi Bob" {
		t.Errorf("unexpected last message: %+v", msg)
	}

	if msg := (*conversations)[2].LastMessage; msg.Id != 7 || msg.format() != "hello" {
		t.Errorf("unexpected last message: %+v", msg)
	}
}

func TestGetUserConversationsMaintenance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	if _, resp, _ := NewInstance("test", server.URL).GetUserConversations(); resp != APIMaintenance {
		t.Errorf("expected APIMaintenance, got %v", resp)
	}
}
