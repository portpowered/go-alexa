package alexa_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func TestGetUserInfoMapsCallerOptionsToRESTRequest(t *testing.T) {
	t.Parallel()

	type observedRequest struct {
		path   string
		cookie string
	}

	observed := make(chan observedRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		observed <- observedRequest{
			path:   request.URL.RequestURI(),
			cookie: request.Header.Get("Cookie"),
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"countryOfResidence":"US","id":"synthetic-user"}`))
	}))
	t.Cleanup(server.Close)

	client, err := alexa.NewClient(
		alexa.WithAlexaWebBaseURL(server.URL),
		alexa.WithRESTHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	session, err := client.NewSession(alexa.WithBearerToken("synthetic-access-token"))
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	t.Cleanup(func() {
		err := session.Close()
		if err != nil {
			t.Errorf("close session: %v", err)
		}
	})

	userInfo, err := session.GetUserInfo(context.Background(), alexaapimodels.UserInfoRequest{
		Platform:  "synthetic-platform",
		Version:   "1.2.3",
		CSRFToken: "synthetic-csrf-token",
	})
	if err != nil {
		t.Fatalf("get user info: %v", err)
	}

	if userInfo.ID != "synthetic-user" {
		t.Fatalf("user ID = %q, want synthetic-user", userInfo.ID)
	}

	request := <-observed
	if request.path != "/api/users/me?platform=synthetic-platform&version=1.2.3" {
		t.Errorf("request URI = %q", request.path)
	}

	if request.cookie != "csrf=synthetic-csrf-token" {
		t.Errorf("cookie = %q", request.cookie)
	}
}
