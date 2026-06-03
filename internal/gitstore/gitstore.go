package gitstore

import (
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Store is a concurrency-safe abstraction over a bare or non-bare git
// repository on disk. Read operations are lock-free; every mutation
// (write / delete) acquires an exclusive lock and produces a commit
// attributed to the given author.
type Store struct {
	repo *git.Repository
	mu   sync.Mutex
}

// Open opens an existing git repository at the given path.
func Open(repoPath string) (*Store, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("gitstore: open %q: %w", repoPath, err)
	}
	return &Store{repo: repo}, nil
}

// Init initialises a new git repository at the given path and returns a Store.
func Init(repoPath string) (*Store, error) {
	repo, err := git.PlainInit(repoPath, false)
	if err != nil {
		return nil, fmt.Errorf("gitstore: init %q: %w", repoPath, err)
	}
	return &Store{repo: repo}, nil
}

// Author identifies the person to attribute commits to.
type Author struct {
	Name  string
	Email string
}

// ReadFile returns the contents of a file at the given path from HEAD.
// The path must be slash-separated and relative to the repository root.
func (s *Store) ReadFile(filePath string) ([]byte, error) {
	filePath = normPath(filePath)

	head, err := s.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("gitstore: head: %w", err)
	}

	commit, err := s.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("gitstore: commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("gitstore: tree: %w", err)
	}

	f, err := tree.File(filePath)
	if err != nil {
		return nil, fmt.Errorf("gitstore: file %q: %w", filePath, err)
	}

	rc, err := f.Reader()
	if err != nil {
		return nil, fmt.Errorf("gitstore: reader %q: %w", filePath, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("gitstore: read %q: %w", filePath, err)
	}

	return data, nil
}

// WriteFile creates or overwrites a file in the worktree, stages it, and
// commits with the given message and author. It acquires an exclusive lock
// for the duration of the operation.
func (s *Store) WriteFile(filePath string, data []byte, msg string, author Author) error {
	filePath = normPath(filePath)

	s.mu.Lock()
	defer s.mu.Unlock()

	wt, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("gitstore: worktree: %w", err)
	}

	fs := wt.Filesystem

	// Ensure parent directories exist.
	if dir := path.Dir(filePath); dir != "." {
		if err := fs.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("gitstore: mkdir %q: %w", dir, err)
		}
	}

	f, err := fs.Create(filePath)
	if err != nil {
		return fmt.Errorf("gitstore: create %q: %w", filePath, err)
	}

	if _, err := f.Write(data); err != nil {
		f.Close()
		return fmt.Errorf("gitstore: write %q: %w", filePath, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("gitstore: close %q: %w", filePath, err)
	}

	if _, err := wt.Add(filePath); err != nil {
		return fmt.Errorf("gitstore: add %q: %w", filePath, err)
	}

	now := time.Now()
	_, err = wt.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  author.Name,
			Email: author.Email,
			When:  now,
		},
	})
	if err != nil {
		return fmt.Errorf("gitstore: commit: %w", err)
	}

	return nil
}

// DeleteFile removes a file from the worktree, stages the removal, and
// commits. It acquires an exclusive lock for the duration of the operation.
func (s *Store) DeleteFile(filePath string, msg string, author Author) error {
	filePath = normPath(filePath)

	s.mu.Lock()
	defer s.mu.Unlock()

	wt, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("gitstore: worktree: %w", err)
	}

	if _, err := wt.Remove(filePath); err != nil {
		return fmt.Errorf("gitstore: remove %q: %w", filePath, err)
	}

	now := time.Now()
	_, err = wt.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  author.Name,
			Email: author.Email,
			When:  now,
		},
	})
	if err != nil {
		return fmt.Errorf("gitstore: commit: %w", err)
	}

	return nil
}

// FileInfo describes a file tracked in the repository at HEAD.
type FileInfo struct {
	Path string
	Size int64
}

// ListFiles returns all files currently tracked at HEAD.
// If prefix is non-empty only files whose path starts with that prefix are
// returned (slash-separated, no leading slash).
func (s *Store) ListFiles(prefix string) ([]FileInfo, error) {
	prefix = normPath(prefix)

	head, err := s.repo.Head()
	if err != nil {
		// No commits yet — empty list.
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("gitstore: head: %w", err)
	}

	commit, err := s.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("gitstore: commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("gitstore: tree: %w", err)
	}

	var files []FileInfo
	err = tree.Files().ForEach(func(f *object.File) error {
		if prefix != "" && !strings.HasPrefix(f.Name, prefix) {
			return nil
		}
		files = append(files, FileInfo{
			Path: f.Name,
			Size: f.Size,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("gitstore: list: %w", err)
	}

	return files, nil
}

// normPath cleans a file path: trims leading/trailing slashes, collapses
// sequences, and returns a clean slash-separated relative path.
func normPath(p string) string {
	p = strings.TrimSpace(p)
	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")
	if p == "." {
		return ""
	}
	return p
}
