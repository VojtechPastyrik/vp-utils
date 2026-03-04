package vault

import (
	"testing"
)

func TestExtractVaultKeys_JSON(t *testing.T) {
	jsonData := `{
		"unseal_keys_b64": ["key1base64", "key2base64", "key3base64"],
		"unseal_threshold": 2
	}`

	keys, threshold := ExtractVaultKeys(jsonData)

	if threshold != 2 {
		t.Errorf("expected threshold 2, got %d", threshold)
	}
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	expected := []string{"key1base64", "key2base64", "key3base64"}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("key %d: expected %q, got %q", i, expected[i], key)
		}
	}
}

func TestExtractVaultKeys_Text(t *testing.T) {
	textData := `Unseal Key 1: abc123def456
Unseal Key 2: ghi789jkl012
Unseal Key 3: mno345pqr678

Initial Root Token: hvs.xxxxxxxxxxxx

Vault initialized with 3 key shares and a key threshold of 2.`

	keys, threshold := ExtractVaultKeys(textData)

	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if threshold != 3 {
		t.Errorf("expected threshold 3 (all keys), got %d", threshold)
	}
	expected := []string{"abc123def456", "ghi789jkl012", "mno345pqr678"}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("key %d: expected %q, got %q", i, expected[i], key)
		}
	}
}

func TestParseVaultKeysText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
		wantErr  bool
	}{
		{
			name: "standard format",
			input: `Unseal Key 1: abc123
Unseal Key 2: def456
Unseal Key 3: ghi789`,
			wantKeys: []string{"abc123", "def456", "ghi789"},
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no keys",
			input:   "some random text\nno keys here",
			wantErr: true,
		},
		{
			name: "single key",
			input: `Unseal Key 1: onlykey

Initial Root Token: hvs.token`,
			wantKeys: []string{"onlykey"},
		},
		{
			name: "five keys",
			input: `Unseal Key 1: k1
Unseal Key 2: k2
Unseal Key 3: k3
Unseal Key 4: k4
Unseal Key 5: k5`,
			wantKeys: []string{"k1", "k2", "k3", "k4", "k5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys, err := ParseVaultKeysText(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(keys) != len(tt.wantKeys) {
				t.Fatalf("expected %d keys, got %d", len(tt.wantKeys), len(keys))
			}
			for i, key := range keys {
				if key != tt.wantKeys[i] {
					t.Errorf("key %d: expected %q, got %q", i, tt.wantKeys[i], key)
				}
			}
		})
	}
}
