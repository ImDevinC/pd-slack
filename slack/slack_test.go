package slack_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pdslack "github.com/imdevinc/pd-slack/slack"
	slackapi "github.com/slack-go/slack"
)

// TestClientUpdateUserGroupMembersResolvesEmails verifies that member emails are resolved to unique Slack user IDs.
func TestClientUpdateUserGroupMembersResolvesEmails(t *testing.T) {
	t.Parallel()

	emailToID := map[string]string{
		"one@example.com": "U001",
		"two@example.com": "U002",
	}
	var updatedGroup string
	var updatedUsers string
	mux := http.NewServeMux()
	mux.HandleFunc("/users.lookupByEmail", func(response http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			t.Errorf("failed to parse lookup form: %v", err)
			return
		}
		ID, ok := emailToID[request.Form.Get("email")]
		if !ok {
			t.Errorf("unexpected email lookup: %q", request.Form.Get("email"))
			return
		}
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(response, `{"ok":true,"user":{"id":%q}}`, ID)
	})
	mux.HandleFunc("/usergroups.users.update", func(response http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			t.Errorf("failed to parse update form: %v", err)
			return
		}
		updatedGroup = request.Form.Get("usergroup")
		updatedUsers = request.Form.Get("users")
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"ok":true,"usergroup":{"id":"S001"}}`)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client := pdslack.New("token", slackapi.OptionAPIURL(server.URL+"/"))

	err := client.UpdateUserGroupMembers(context.Background(), "S001", []string{
		"one@example.com",
		"two@example.com",
		"one@example.com",
	})
	if err != nil {
		t.Fatalf("UpdateUserGroupMembers returned an error: %v", err)
	}
	if updatedGroup != "S001" {
		t.Errorf("updated group = %q, want %q", updatedGroup, "S001")
	}
	if updatedUsers != "U001,U002" {
		t.Errorf("updated users = %q, want %q", updatedUsers, "U001,U002")
	}
}

// TestClientUpdateUserGroupMembersLookupFailure verifies that a failed email lookup prevents a partial group update.
func TestClientUpdateUserGroupMembersLookupFailure(t *testing.T) {
	t.Parallel()

	updateCalled := false
	mux := http.NewServeMux()
	mux.HandleFunc("/users.lookupByEmail", func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"ok":false,"error":"users_not_found"}`)
	})
	mux.HandleFunc("/usergroups.users.update", func(response http.ResponseWriter, _ *http.Request) {
		updateCalled = true
		fmt.Fprint(response, `{"ok":true}`)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client := pdslack.New("token", slackapi.OptionAPIURL(server.URL+"/"))

	err := client.UpdateUserGroupMembers(context.Background(), "S001", []string{"missing@example.com"})
	if err == nil {
		t.Fatal("UpdateUserGroupMembers returned nil, want an error")
	}
	if !strings.Contains(err.Error(), `failed to find Slack user for "missing@example.com": users_not_found`) {
		t.Errorf("error = %q, want email lookup context", err)
	}
	if updateCalled {
		t.Error("user group update was called after a failed lookup")
	}
}

// TestClientUpdateUserGroupMembersRejectsEmptyMembers verifies that Slack is not called with an invalid empty member list.
func TestClientUpdateUserGroupMembersRejectsEmptyMembers(t *testing.T) {
	t.Parallel()

	client := pdslack.New("token")
	err := client.UpdateUserGroupMembers(context.Background(), "S001", nil)
	if err == nil {
		t.Fatal("UpdateUserGroupMembers returned nil, want an error")
	}
	if !strings.Contains(err.Error(), "without members") {
		t.Errorf("error = %q, want empty member context", err)
	}
}
