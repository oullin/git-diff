package ai

import "testing"

func TestAnthropicSupportsModel(t *testing.T) {
	p := NewAnthropicProvider()

	cases := map[string]bool{
		"claude-sonnet-4-5": true,
		"claude-opus-9":     true,
		"CLAUDE-3":          true,
		"gpt-5":             false,
		"":                  false,
		"   ":               false,
	}

	for model, want := range cases {
		if got := p.SupportsModel(model); got != want {
			t.Fatalf("SupportsModel(%q) = %v, want %v", model, got, want)
		}
	}
}

func TestAnthropicProviderIdentity(t *testing.T) {
	p := NewAnthropicProvider()

	if p.ID() != "anthropic" {
		t.Fatalf("id = %q", p.ID())
	}

	if p.DefaultModel() != anthropicDefault {
		t.Fatalf("default model = %q", p.DefaultModel())
	}
}
