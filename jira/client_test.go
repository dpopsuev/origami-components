package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetIssue_Done(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issue := Issue{
			Key: "OCPBUGS-70233",
		}
		issue.Fields.Summary = "phc2sys process state"
		issue.Fields.Status = IssueStatus{Name: "Closed"}
		issue.Fields.Status.StatusCategory.Key = "done"
		json.NewEncoder(w).Encode(issue)
	}))
	defer server.Close()

	c, err := New(server.URL, "")
	if err != nil {
		t.Fatal(err)
	}

	issue, err := c.GetIssue(context.Background(), "OCPBUGS-70233")
	if err != nil {
		t.Fatal(err)
	}
	if !issue.IsDone() {
		t.Errorf("expected IsDone=true, got status=%s", issue.StatusName())
	}
}

func TestGetIssue_Open(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issue := Issue{Key: "OCPBUGS-99999"}
		issue.Fields.Status = IssueStatus{Name: "In Progress"}
		issue.Fields.Status.StatusCategory.Key = "indeterminate"
		json.NewEncoder(w).Encode(issue)
	}))
	defer server.Close()

	c, _ := New(server.URL, "")
	issue, err := c.GetIssue(context.Background(), "OCPBUGS-99999")
	if err != nil {
		t.Fatal(err)
	}
	if issue.IsDone() {
		t.Error("expected IsDone=false for In Progress")
	}
}

func TestNew_RequiresHTTPS(t *testing.T) {
	_, err := New("http://evil.com", "tok")
	if err == nil {
		t.Error("expected HTTPS enforcement error")
	}
}
