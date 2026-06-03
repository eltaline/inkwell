package gitstore

import (
	"os"
	"sync"
	"testing"
)

var testAuthor = Author{Name: "Test User", Email: "test@example.com"}

func tmpStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := Init(dir)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	return s
}

func TestWriteAndReadFile(t *testing.T) {
	s := tmpStore(t)

	content := []byte("hello world")
	if err := s.WriteFile("docs/readme.txt", content, "add readme", testAuthor); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := s.ReadFile("docs/readme.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestOverwriteFile(t *testing.T) {
	s := tmpStore(t)

	if err := s.WriteFile("a.txt", []byte("v1"), "v1", testAuthor); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteFile("a.txt", []byte("v2"), "v2", testAuthor); err != nil {
		t.Fatal(err)
	}

	got, err := s.ReadFile("a.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "v2" {
		t.Errorf("got %q, want %q", got, "v2")
	}
}

func TestDeleteFile(t *testing.T) {
	s := tmpStore(t)

	if err := s.WriteFile("rm-me.txt", []byte("bye"), "add", testAuthor); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteFile("rm-me.txt", "remove", testAuthor); err != nil {
		t.Fatal(err)
	}

	_, err := s.ReadFile("rm-me.txt")
	if err == nil {
		t.Fatal("expected error reading deleted file")
	}
}

func TestListFiles(t *testing.T) {
	s := tmpStore(t)

	files := []string{"a.txt", "dir/b.txt", "dir/c.txt", "other/d.txt"}
	for _, f := range files {
		if err := s.WriteFile(f, []byte(f), "add "+f, testAuthor); err != nil {
			t.Fatal(err)
		}
	}

	all, err := s.ListFiles("")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 4 {
		t.Errorf("ListFiles(\"\") returned %d files, want 4", len(all))
	}

	dirOnly, err := s.ListFiles("dir/")
	if err != nil {
		t.Fatal(err)
	}
	if len(dirOnly) != 2 {
		t.Errorf("ListFiles(\"dir/\") returned %d files, want 2", len(dirOnly))
	}
}

func TestListFilesEmptyRepo(t *testing.T) {
	s := tmpStore(t)

	files, err := s.ListFiles("")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestOpenExistingRepo(t *testing.T) {
	dir := t.TempDir()
	s1, err := Init(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s1.WriteFile("f.txt", []byte("data"), "init", testAuthor); err != nil {
		t.Fatal(err)
	}

	s2, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got, err := s2.ReadFile("f.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "data" {
		t.Errorf("got %q, want %q", got, "data")
	}
}

func TestOpenNonExistent(t *testing.T) {
	_, err := Open(os.TempDir() + "/nonexistent-repo-abc123")
	if err == nil {
		t.Fatal("expected error opening nonexistent repo")
	}
}

func TestConcurrentWrites(t *testing.T) {
	s := tmpStore(t)

	var wg sync.WaitGroup
	errs := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			name := "file" + string(rune('A'+n)) + ".txt"
			if err := s.WriteFile(name, []byte("data"), "add "+name, testAuthor); err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent write error: %v", err)
	}

	files, err := s.ListFiles("")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 10 {
		t.Errorf("expected 10 files, got %d", len(files))
	}
}

func TestNormPath(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"/a/b/c", "a/b/c"},
		{"a//b/../c", "a/c"},
		{"  /foo/  ", "foo"},
		{".", ""},
		{"", ""},
	}
	for _, tc := range cases {
		got := normPath(tc.in)
		if got != tc.want {
			t.Errorf("normPath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
