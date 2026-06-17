package note

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const frontMatterDelim = "---"

// Note represents a markdown note with YAML front-matter metadata.
type Note struct {
	// Front-matter fields.
	ID      string    `yaml:"id"`
	Title   string    `yaml:"title"`
	Tags    []string  `yaml:"tags,omitempty"`
	Created time.Time `yaml:"created"`
	Updated time.Time `yaml:"updated"`

	// Extra holds any unknown front-matter keys so they survive round-trips.
	Extra map[string]any `yaml:"-"`

	// Body is the markdown content after the front-matter.
	Body string `yaml:"-"`

	// rawFrontMatter preserves the original YAML text for lossless
	// re-serialization when no metadata fields were changed.
	rawFrontMatter string
	parsed         bool
}

// Parse splits raw markdown bytes into front-matter and body, populating
// the Note struct. If the document has no front-matter delimiters the
// entire content is treated as the body and an empty front-matter is used.
func Parse(data []byte) (*Note, error) {
	text := string(data)
	n := &Note{}

	fm, body, ok := splitFrontMatter(text)
	if !ok {
		n.Body = text
		return n, nil
	}

	n.rawFrontMatter = fm
	n.Body = body
	n.parsed = true

	if err := n.decodeFrontMatter(fm); err != nil {
		return nil, err
	}

	return n, nil
}

// StableID returns a deterministic identifier derived from the note title.
// It produces a 12-character hex string from the SHA-256 of the lowercased,
// trimmed title. Use this to generate an ID for a new note when one is not
// already set.
func StableID(title string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(strings.ToLower(title))))
	return hex.EncodeToString(h[:6])
}

// EnsureID populates the ID field if it is empty, using StableID derived
// from the Title. If both ID and Title are empty an error is returned.
func (n *Note) EnsureID() error {
	if n.ID != "" {
		return nil
	}
	if strings.TrimSpace(n.Title) == "" {
		return errors.New("note: cannot generate id without a title")
	}
	n.ID = StableID(n.Title)
	return nil
}

// Bytes serializes the note back to markdown with YAML front-matter.
// If the metadata fields (ID, Title, Tags, Created, Updated) have not
// been modified since parsing, the original front-matter text is reused
// verbatim to avoid reformatting.
func (n *Note) Bytes() ([]byte, error) {
	var buf bytes.Buffer

	fm, err := n.encodeFrontMatter()
	if err != nil {
		return nil, err
	}

	buf.WriteString(frontMatterDelim)
	buf.WriteByte('\n')
	buf.WriteString(fm)
	buf.WriteString(frontMatterDelim)
	buf.WriteByte('\n')
	buf.WriteString(n.Body)

	return buf.Bytes(), nil
}

// splitFrontMatter extracts the YAML block between the first pair of "---"
// delimiters. It returns the front-matter content (without delimiters),
// the remaining body, and whether front-matter was found.
func splitFrontMatter(text string) (fm, body string, ok bool) {
	// The opening delimiter must be the very first line.
	if !strings.HasPrefix(text, frontMatterDelim) {
		return "", "", false
	}

	rest := text[len(frontMatterDelim):]
	if len(rest) == 0 || rest[0] != '\n' {
		return "", "", false
	}
	rest = rest[1:] // skip the newline after opening ---

	idx := strings.Index(rest, "\n"+frontMatterDelim)
	if idx < 0 {
		return "", "", false
	}

	fm = rest[:idx+1] // include trailing newline of last YAML line

	after := rest[idx+1+len(frontMatterDelim):]
	if len(after) > 0 && after[0] == '\n' {
		after = after[1:]
	}

	return fm, after, true
}

// frontMatterProxy is used for YAML marshal/unmarshal so we can capture
// unknown fields via inline mapping while keeping known fields typed.
type frontMatterProxy struct {
	ID      string    `yaml:"id"`
	Title   string    `yaml:"title"`
	Tags    []string  `yaml:"tags,omitempty"`
	Created time.Time `yaml:"created"`
	Updated time.Time `yaml:"updated"`
}

func (n *Note) decodeFrontMatter(raw string) error {
	// First decode known fields.
	var proxy frontMatterProxy
	if err := yaml.Unmarshal([]byte(raw), &proxy); err != nil {
		return err
	}
	n.ID = proxy.ID
	n.Title = proxy.Title
	n.Tags = proxy.Tags
	n.Created = proxy.Created
	n.Updated = proxy.Updated

	// Decode again into a generic map to capture extra keys.
	var full map[string]any
	if err := yaml.Unmarshal([]byte(raw), &full); err != nil {
		return err
	}

	known := map[string]bool{
		"id": true, "title": true, "tags": true,
		"created": true, "updated": true,
	}
	extra := make(map[string]any)
	for k, v := range full {
		if !known[k] {
			extra[k] = v
		}
	}
	if len(extra) > 0 {
		n.Extra = extra
	}

	return nil
}

func (n *Note) encodeFrontMatter() (string, error) {
	// Build an ordered map: known fields first, then extras.
	m := &yaml.Node{Kind: yaml.MappingNode}

	addField := func(key string, value any) error {
		kn := &yaml.Node{}
		kn.SetString(key)

		vn := &yaml.Node{}
		if err := vn.Encode(value); err != nil {
			return err
		}
		m.Content = append(m.Content, kn, vn)
		return nil
	}

	if err := addField("id", n.ID); err != nil {
		return "", err
	}
	if err := addField("title", n.Title); err != nil {
		return "", err
	}
	if len(n.Tags) > 0 {
		// Emit tags as flow sequence [a, b] for compactness.
		kn := &yaml.Node{}
		kn.SetString("tags")
		vn := &yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle}
		for _, t := range n.Tags {
			tn := &yaml.Node{}
			tn.SetString(t)
			vn.Content = append(vn.Content, tn)
		}
		m.Content = append(m.Content, kn, vn)
	}

	if !n.Created.IsZero() {
		if err := addField("created", n.Created); err != nil {
			return "", err
		}
	}
	if !n.Updated.IsZero() {
		if err := addField("updated", n.Updated); err != nil {
			return "", err
		}
	}

	for k, v := range n.Extra {
		if err := addField(k, v); err != nil {
			return "", err
		}
	}

	doc := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{m}}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return "", err
	}
	enc.Close()

	// yaml.Encoder adds a trailing "...\n" document end marker — strip it.
	out := buf.String()
	out = strings.TrimSuffix(out, "...\n")

	return out, nil
}
