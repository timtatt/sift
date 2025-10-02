package sift

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestGetIndentLevel(t *testing.T) {
	tests := []struct {
		name     string
		testName string
		want     int
	}{
		{
			name:     "no slashes",
			testName: "TestSimple",
			want:     0,
		},
		{
			name:     "one slash",
			testName: "TestParent/TestChild",
			want:     1,
		},
		{
			name:     "two slashes",
			testName: "TestGrandparent/TestParent/TestChild",
			want:     2,
		},
		{
			name:     "three slashes",
			testName: "TestRoot/TestLevel1/TestLevel2/TestLevel3",
			want:     3,
		},
		{
			name:     "empty string",
			testName: "",
			want:     0,
		},
		{
			name:     "trailing slash",
			testName: "TestParent/TestChild/",
			want:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getIndentLevel(tt.testName)
			if got != tt.want {
				t.Errorf("getIndentLevel(%q) = %d, want %d", tt.testName, got, tt.want)
			}
		})
	}
}

func TestGetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		testName string
		want     string
	}{
		{
			name:     "no slashes",
			testName: "TestSimple",
			want:     "TestSimple",
		},
		{
			name:     "one slash",
			testName: "TestParent/TestChild",
			want:     "TestChild",
		},
		{
			name:     "two slashes",
			testName: "TestGrandparent/TestParent/TestChild",
			want:     "TestChild",
		},
		{
			name:     "empty string",
			testName: "",
			want:     "",
		},
		{
			name:     "trailing slash",
			testName: "TestParent/TestChild/",
			want:     "",
		},
		{
			name:     "starts with slash",
			testName: "/TestChild",
			want:     "TestChild",
		},
		{
			name:     "multiple levels deep",
			testName: "A/B/C/D/E/F",
			want:     "F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDisplayName(tt.testName)
			if got != tt.want {
				t.Errorf("getDisplayName(%q) = %q, want %q", tt.testName, got, tt.want)
			}
		})
	}
}

func TestGetIndentWithLines(t *testing.T) {
	tests := []struct {
		name        string
		indentLevel int
		wantEmpty   bool
		wantCount   int
	}{
		{
			name:        "zero indent",
			indentLevel: 0,
			wantEmpty:   true,
			wantCount:   0,
		},
		{
			name:        "one indent",
			indentLevel: 1,
			wantEmpty:   false,
			wantCount:   1,
		},
		{
			name:        "two indents",
			indentLevel: 2,
			wantEmpty:   false,
			wantCount:   2,
		},
		{
			name:        "five indents",
			indentLevel: 5,
			wantEmpty:   false,
			wantCount:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getIndentWithLines(tt.indentLevel)
			
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("getIndentWithLines(%d) = %q, want empty string", tt.indentLevel, got)
				}
				return
			}

			if len(got) == 0 {
				t.Errorf("getIndentWithLines(%d) returned empty string, want non-empty", tt.indentLevel)
			}
		})
	}
}

func TestLastKeysMatch(t *testing.T) {
	tests := []struct {
		name       string
		keyBuffer  []string
		binding    key.Binding
		want       bool
	}{
		{
			name:      "match two-char sequence zA",
			keyBuffer: []string{"z", "A"},
			binding: key.NewBinding(
				key.WithKeys("zA"),
			),
			want: true,
		},
		{
			name:      "match two-char sequence zR",
			keyBuffer: []string{"z", "R"},
			binding: key.NewBinding(
				key.WithKeys("zR"),
			),
			want: true,
		},
		{
			name:      "no match wrong sequence",
			keyBuffer: []string{"z", "X"},
			binding: key.NewBinding(
				key.WithKeys("zA"),
			),
			want: false,
		},
		{
			name:      "match with history before",
			keyBuffer: []string{"k", "j", "z", "a"},
			binding: key.NewBinding(
				key.WithKeys("za"),
			),
			want: true,
		},
		{
			name:      "match single char at end",
			keyBuffer: []string{"z", "a"},
			binding: key.NewBinding(
				key.WithKeys("a"),
			),
			want: true,
		},
		{
			name:      "no match buffer too short",
			keyBuffer: []string{"z"},
			binding: key.NewBinding(
				key.WithKeys("zA"),
			),
			want: false,
		},
		{
			name:      "match multiple possible keys",
			keyBuffer: []string{"z", "a"},
			binding: key.NewBinding(
				key.WithKeys("za", "enter", " "),
			),
			want: true,
		},
		{
			name:      "empty buffer no match",
			keyBuffer: []string{"", ""},
			binding: key.NewBinding(
				key.WithKeys("zA"),
			),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &siftModel{
				keyBuffer: tt.keyBuffer,
			}
			got := m.LastKeysMatch(tt.binding)
			if got != tt.want {
				t.Errorf("LastKeysMatch() with buffer %v and binding keys %v = %v, want %v",
					tt.keyBuffer, tt.binding.Keys(), got, tt.want)
			}
		})
	}
}

func TestBufferKey(t *testing.T) {
	tests := []struct {
		name           string
		initialBuffer  []string
		keyToAdd       string
		expectedBuffer []string
	}{
		{
			name:           "add to empty buffer",
			initialBuffer:  []string{"", ""},
			keyToAdd:       "a",
			expectedBuffer: []string{"", "a"},
		},
		{
			name:           "add to partially filled buffer",
			initialBuffer:  []string{"", "z"},
			keyToAdd:       "a",
			expectedBuffer: []string{"z", "a"},
		},
		{
			name:           "add to full buffer",
			initialBuffer:  []string{"z", "a"},
			keyToAdd:       "b",
			expectedBuffer: []string{"a", "b"},
		},
		{
			name:           "add multiple times",
			initialBuffer:  []string{"", ""},
			keyToAdd:       "x",
			expectedBuffer: []string{"", "x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &siftModel{
				keyBuffer: make([]string, len(tt.initialBuffer)),
			}
			copy(m.keyBuffer, tt.initialBuffer)

			msg := tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune(tt.keyToAdd),
			}

			m.BufferKey(msg)

			if len(m.keyBuffer) != len(tt.expectedBuffer) {
				t.Errorf("Buffer length = %d, want %d", len(m.keyBuffer), len(tt.expectedBuffer))
				return
			}

			for i := range m.keyBuffer {
				if m.keyBuffer[i] != tt.expectedBuffer[i] {
					t.Errorf("Buffer[%d] = %q, want %q", i, m.keyBuffer[i], tt.expectedBuffer[i])
				}
			}
		})
	}
}

func TestBufferKeySequence(t *testing.T) {
	m := &siftModel{
		keyBuffer: make([]string, 2),
	}

	keys := []struct {
		key      string
		expected []string
	}{
		{"z", []string{"", "z"}},
		{"A", []string{"z", "A"}},
		{"k", []string{"A", "k"}},
	}

	for _, k := range keys {
		msg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune(k.key),
		}
		m.BufferKey(msg)

		for i := range m.keyBuffer {
			if m.keyBuffer[i] != k.expected[i] {
				t.Errorf("After adding %q, buffer[%d] = %q, want %q",
					k.key, i, m.keyBuffer[i], k.expected[i])
			}
		}
	}
}

func TestBufferKeyAndLastKeysMatchIntegration(t *testing.T) {
	m := &siftModel{
		keyBuffer: make([]string, 2),
	}

	binding := key.NewBinding(key.WithKeys("zA"))

	if m.LastKeysMatch(binding) {
		t.Error("Should not match with empty buffer")
	}

	m.BufferKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")})
	if m.LastKeysMatch(binding) {
		t.Error("Should not match with only 'z' in buffer")
	}

	m.BufferKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("A")})
	if !m.LastKeysMatch(binding) {
		t.Error("Should match with 'zA' in buffer")
	}

	m.BufferKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if m.LastKeysMatch(binding) {
		t.Error("Should not match after buffer shifted")
	}
}
