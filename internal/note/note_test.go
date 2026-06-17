package note

import (
	"strings"
	"testing"
	"time"
)

const sampleNote = `---
id: abc123
title: My First Note
tags: [go, notes]
created: 2025-01-15T10:30:00Z
updated: 2025-06-01T14:00:00Z
---
# Hello

This is the **body** of the note.
`

func TestParseBasic(t *testing.T) {
	n, err := Parse([]byte(sampleNote))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if n.ID != "abc123" {
		t.Errorf("ID = %q, want %q", n.ID, "abc123")
	}
	if n.Title != "My First Note" {
		t.Errorf("Title = %q, want %q", n.Title, "My First Note")
	}
	if len(n.Tags) != 2 || n.Tags[0] != "go" || n.Tags[1] != "notes" {
		t.Errorf("Tags = %v, want [go notes]", n.Tags)
	}
	if n.Created.Year() != 2025 || n.Created.Month() != time.January {
		t.Errorf("Created = %v, unexpected", n.Created)
	}
	if n.Updated.Year() != 2025 || n.Updated.Month() != time.June {
		t.Errorf("Updated = %v, unexpected", n.Updated)
	}
	if !strings.Contains(n.Body, "# Hello") {
		t.Errorf("Body missing heading, got: %q", n.Body)
	}
	if !strings.Contains(n.Body, "**body**") {
		t.Errorf("Body missing bold text, got: %q", n.Body)
	}
}

func TestParseNoFrontMatter(t *testing.T) {
	raw := "# Just markdown\n\nNo front-matter here.\n"
	n, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if n.ID != "" {
		t.Errorf("ID = %q, want empty", n.ID)
	}
	if n.Body != raw {
		t.Errorf("Body = %q, want %q", n.Body, raw)
	}
}

func TestParseExtraFields(t *testing.T) {
	raw := `---
id: x1
title: Extra
author: someone
draft: true
created: 2025-01-01T00:00:00Z
updated: 2025-01-01T00:00:00Z
---
Body.
`
	n, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if n.Extra["author"] != "someone" {
		t.Errorf("Extra[author] = %v, want 'someone'", n.Extra["author"])
	}
	if n.Extra["draft"] != true {
		t.Errorf("Extra[draft] = %v, want true", n.Extra["draft"])
	}
}

func TestStableID(t *testing.T) {
	id1 := StableID("My Note Title")
	id2 := StableID("my note title")
	id3 := StableID("  My Note Title  ")

	if id1 != id2 {
		t.Errorf("StableID not case-insensitive: %q != %q", id1, id2)
	}
	if id1 != id3 {
		t.Errorf("StableID not trim-insensitive: %q != %q", id1, id3)
	}
	if len(id1) != 12 {
		t.Errorf("StableID length = %d, want 12", len(id1))
	}

	other := StableID("Different Title")
	if id1 == other {
		t.Error("StableID collision for different titles")
	}
}

func TestEnsureID(t *testing.T) {
	n := &Note{Title: "Test Note"}
	if err := n.EnsureID(); err != nil {
		t.Fatalf("EnsureID: %v", err)
	}
	if n.ID == "" {
		t.Error("EnsureID did not set ID")
	}
	if n.ID != StableID("Test Note") {
		t.Errorf("EnsureID = %q, want %q", n.ID, StableID("Test Note"))
	}

	// Should not overwrite existing ID.
	n2 := &Note{ID: "custom", Title: "Test Note"}
	if err := n2.EnsureID(); err != nil {
		t.Fatalf("EnsureID: %v", err)
	}
	if n2.ID != "custom" {
		t.Errorf("EnsureID overwrote existing ID: %q", n2.ID)
	}

	// Error without title or ID.
	n3 := &Note{}
	if err := n3.EnsureID(); err == nil {
		t.Error("EnsureID should fail without title and id")
	}
}

func TestRoundTrip(t *testing.T) {
	n, err := Parse([]byte(sampleNote))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	out, err := n.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	// Re-parse the output.
	n2, err := Parse(out)
	if err != nil {
		t.Fatalf("re-Parse: %v", err)
	}

	if n2.ID != n.ID {
		t.Errorf("round-trip ID: %q != %q", n2.ID, n.ID)
	}
	if n2.Title != n.Title {
		t.Errorf("round-trip Title: %q != %q", n2.Title, n.Title)
	}
	if len(n2.Tags) != len(n.Tags) {
		t.Errorf("round-trip Tags count: %d != %d", len(n2.Tags), len(n.Tags))
	}
	if !n2.Created.Equal(n.Created) {
		t.Errorf("round-trip Created: %v != %v", n2.Created, n.Created)
	}
	if !n2.Updated.Equal(n.Updated) {
		t.Errorf("round-trip Updated: %v != %v", n2.Updated, n.Updated)
	}
	if n2.Body != n.Body {
		t.Errorf("round-trip Body:\ngot:  %q\nwant: %q", n2.Body, n.Body)
	}
}

func TestRoundTripExtraFields(t *testing.T) {
	raw := `---
id: x1
title: Extra
author: someone
created: 2025-01-01T00:00:00Z
updated: 2025-01-01T00:00:00Z
---
Body.
`
	n, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	out, err := n.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	n2, err := Parse(out)
	if err != nil {
		t.Fatalf("re-Parse: %v", err)
	}

	if n2.Extra["author"] != "someone" {
		t.Errorf("round-trip Extra[author] = %v, want 'someone'", n2.Extra["author"])
	}
}

func TestBytesNewNote(t *testing.T) {
	now := time.Date(2025, 6, 17, 12, 0, 0, 0, time.UTC)
	n := &Note{
		ID:      "test123",
		Title:   "Brand New",
		Tags:    []string{"fresh"},
		Created: now,
		Updated: now,
		Body:    "Content here.\n",
	}

	out, err := n.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	s := string(out)
	if !strings.HasPrefix(s, "---\n") {
		t.Error("output should start with ---")
	}
	if !strings.Contains(s, "id: test123") {
		t.Error("output missing id field")
	}
	if !strings.Contains(s, "title: Brand New") {
		t.Error("output missing title field")
	}
	if !strings.Contains(s, "Content here.") {
		t.Error("output missing body")
	}
}
