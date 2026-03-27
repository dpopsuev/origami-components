package github

import "testing"

func TestParsePRRef_Valid(t *testing.T) {
	owner, repo, num, err := ParsePRRef("redhat-cne/cloud-event-proxy#633")
	if err != nil {
		t.Fatal(err)
	}
	if owner != "redhat-cne" {
		t.Errorf("owner = %q, want redhat-cne", owner)
	}
	if repo != "cloud-event-proxy" {
		t.Errorf("repo = %q, want cloud-event-proxy", repo)
	}
	if num != 633 {
		t.Errorf("number = %d, want 633", num)
	}
}

func TestParsePRRef_Invalid(t *testing.T) {
	for _, bad := range []string{"no-hash", "org#abc", "just-a-string", "#123", "/"} {
		_, _, _, err := ParsePRRef(bad)
		if err == nil {
			t.Errorf("ParsePRRef(%q) should fail", bad)
		}
	}
}
