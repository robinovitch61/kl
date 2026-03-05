package model_test

import (
	"strings"
	"testing"

	"github.com/robinovitch61/kl/internal/k8s/container"
	"github.com/robinovitch61/kl/internal/model"
)

func TestNewMatcher_EmptyArgs(t *testing.T) {
	m, err := model.NewMatcher(model.NewMatcherArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// empty matcher should never match any container
	if m.MatchesContainer(matcherTestContainer()) {
		t.Error("expected empty matcher to not match any container")
	}
}

func TestNewMatcher_InvalidRegex(t *testing.T) {
	fields := []struct {
		name string
		args model.NewMatcherArgs
	}{
		{"cluster", model.NewMatcherArgs{Cluster: "[invalid"}},
		{"namespace", model.NewMatcherArgs{Namespace: "[invalid"}},
		{"podOwner", model.NewMatcherArgs{PodOwner: "[invalid"}},
		{"pod", model.NewMatcherArgs{Pod: "[invalid"}},
		{"container", model.NewMatcherArgs{Container: "[invalid"}},
	}
	for _, f := range fields {
		t.Run(f.name, func(t *testing.T) {
			_, err := model.NewMatcher(f.args)
			if err == nil {
				t.Fatal("expected error for invalid regex")
			}
			if !strings.Contains(err.Error(), f.name) {
				t.Errorf("expected error to contain field name %q, got %q", f.name, err.Error())
			}
		})
	}
}

func TestNewMatcher_ValidRegex(t *testing.T) {
	m, err := model.NewMatcher(model.NewMatcherArgs{
		Cluster:   "prod-.*",
		Namespace: "default",
		PodOwner:  "my-app",
		Pod:       "my-app-.*",
		Container: "web",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// non-empty matcher should match a matching container
	if !m.MatchesContainer(matcherTestContainer()) {
		t.Error("expected non-empty matcher to match a matching container")
	}
}

func matcherTestContainer() container.Container {
	return container.Container{
		Cluster:   "prod-us",
		Namespace: "default",
		PodOwner:  "my-app",
		Pod:       "my-app-abc123",
		Name:      "web",
	}
}

func TestMatchesContainer_EmptyMatcher(t *testing.T) {
	m, err := model.NewMatcher(model.NewMatcherArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.MatchesContainer(matcherTestContainer()) {
		t.Error("empty matcher should never match")
	}
}

func TestMatchesContainer_AllFieldsMatch(t *testing.T) {
	m, err := model.NewMatcher(model.NewMatcherArgs{
		Cluster:   "prod-.*",
		Namespace: "default",
		PodOwner:  "my-app",
		Pod:       "my-app-.*",
		Container: "web",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.MatchesContainer(matcherTestContainer()) {
		t.Error("expected match when all fields match")
	}
}

func TestMatchesContainer_OneFieldMismatch(t *testing.T) {
	cases := []struct {
		name string
		args model.NewMatcherArgs
	}{
		{"cluster mismatch", model.NewMatcherArgs{Cluster: "^staging$", Namespace: "default", PodOwner: "my-app", Pod: "my-app-.*", Container: "web"}},
		{"namespace mismatch", model.NewMatcherArgs{Cluster: "prod-.*", Namespace: "^kube-system$", PodOwner: "my-app", Pod: "my-app-.*", Container: "web"}},
		{"podOwner mismatch", model.NewMatcherArgs{Cluster: "prod-.*", Namespace: "default", PodOwner: "^other-app$", Pod: "my-app-.*", Container: "web"}},
		{"pod mismatch", model.NewMatcherArgs{Cluster: "prod-.*", Namespace: "default", PodOwner: "my-app", Pod: "^other-pod$", Container: "web"}},
		{"container mismatch", model.NewMatcherArgs{Cluster: "prod-.*", Namespace: "default", PodOwner: "my-app", Pod: "my-app-.*", Container: "^sidecar$"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := model.NewMatcher(tc.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.MatchesContainer(matcherTestContainer()) {
				t.Error("expected no match when one field mismatches")
			}
		})
	}
}

func TestMatchesContainer_EmptyPatternMatchesAll(t *testing.T) {
	m, err := model.NewMatcher(model.NewMatcherArgs{Container: "web"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.MatchesContainer(matcherTestContainer()) {
		t.Error("empty pattern fields should match any value")
	}
}

func TestMatchesContainer_PartialRegex(t *testing.T) {
	m, err := model.NewMatcher(model.NewMatcherArgs{Pod: "app"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.MatchesContainer(matcherTestContainer()) {
		t.Error("partial regex should match as substring")
	}
}
