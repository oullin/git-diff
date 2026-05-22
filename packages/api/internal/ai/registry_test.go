package ai

import (
	"context"
	"strings"
	"testing"
)

type stubProvider struct {
	id      string
	model   string
	support bool
}

func (s stubProvider) ID() string                { return s.id }
func (s stubProvider) DefaultModel() string      { return s.model }
func (s stubProvider) SupportsModel(string) bool { return s.support }
func (s stubProvider) Generate(context.Context, GenerateRequest) (GenerateResponse, error) {
	return GenerateResponse{ProviderID: s.id, ModelID: s.model, Text: "ok"}, nil
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	r.Register(stubProvider{id: "alpha", model: "m1", support: true})

	got, err := r.Get("alpha")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.ID() != "alpha" {
		t.Fatalf("id = %q", got.ID())
	}
}

func TestRegistryGetUnknownIDListsAvailable(t *testing.T) {
	r := NewRegistry()
	r.Register(stubProvider{id: "alpha"})
	r.Register(stubProvider{id: "beta"})

	_, err := r.Get("gamma")

	if err == nil {
		t.Fatal("expected error for unknown id")
	}

	if !strings.Contains(err.Error(), "alpha") || !strings.Contains(err.Error(), "beta") {
		t.Fatalf("error should list available providers, got %q", err.Error())
	}
}

func TestRegistryRegisterReplacesAndReturnsPrevious(t *testing.T) {
	r := NewRegistry()
	r.Register(stubProvider{id: "alpha", model: "v1"})

	prev := r.Register(stubProvider{id: "alpha", model: "v2"})

	if prev == nil || prev.DefaultModel() != "v1" {
		t.Fatalf("expected previous v1 provider, got %#v", prev)
	}

	got, _ := r.Get("alpha")

	if got.DefaultModel() != "v2" {
		t.Fatalf("replacement not applied, model = %q", got.DefaultModel())
	}
}

func TestRegistryListSorted(t *testing.T) {
	r := NewRegistry()
	r.Register(stubProvider{id: "beta"})
	r.Register(stubProvider{id: "alpha"})
	r.Register(stubProvider{id: "gamma"})

	ids := r.List()

	if len(ids) != 3 || ids[0] != "alpha" || ids[1] != "beta" || ids[2] != "gamma" {
		t.Fatalf("list = %v, want sorted [alpha beta gamma]", ids)
	}
}
