package store

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitRepo makes a bare "remote" with one commit on main, so Clone and
// Checkout run for real without a network.
func gitRepo(t *testing.T) string {
	t.Helper()
	work := t.TempDir()
	remote := filepath.Join(t.TempDir(), "remote.git")
	runGit(t, work, "init", "-q", "-b", "main")
	runGit(t, work, "commit", "-q", "--allow-empty", "-m", "one")
	runGit(t, t.TempDir(), "clone", "-q", "--bare", work, remote)
	return remote
}

// storeFor returns a store whose every repo clones from remote.
func storeFor(t *testing.T, remote string) *Store {
	t.Helper()
	s := New(t.TempDir(), "github.com")
	s.Remote = func(Repo) string { return remote }
	return s
}

func TestCloneUsesADeclaredCheckout(t *testing.T) {
	declared := t.TempDir()
	runGit(t, declared, "init", "-q", "-b", "main")
	runGit(t, declared, "commit", "-q", "--allow-empty", "-m", "one")

	s := storeFor(t, filepath.Join(t.TempDir(), "never-used"))
	s.Declared["o-r"] = declared

	got, created, err := s.Clone(t.Context(), Repo{Owner: "o", Name: "r", Base: "main"})
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	if created {
		t.Fatal("a declared checkout must not be cloned")
	}
	// git resolves symlinks, so /tmp becomes /private/tmp on macOS.
	if got != gitResolved(t, declared) {
		t.Fatalf("Clone = %q, want %q", got, gitResolved(t, declared))
	}
}

func TestCloneFetchesAMissingRepo(t *testing.T) {
	s := storeFor(t, gitRepo(t))

	main, created, err := s.Clone(t.Context(), Repo{Owner: "o", Name: "r", Base: "main"})
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	if !created {
		t.Fatal("first Clone should create the clone")
	}
	if _, err := os.Stat(filepath.Join(main, ".git")); err != nil {
		t.Fatalf("clone has no .git: %v", err)
	}

	again, createdAgain, err := s.Clone(t.Context(), Repo{Owner: "o", Name: "r", Base: "main"})
	if err != nil {
		t.Fatalf("second Clone: %v", err)
	}
	if createdAgain {
		t.Fatal("second Clone should reuse the clone")
	}
	if again != main {
		t.Fatalf("second Clone = %q, want %q", again, main)
	}
}

func TestCheckoutIsReusedAndReleased(t *testing.T) {
	s := storeFor(t, gitRepo(t))
	repo := Repo{Owner: "o", Name: "r", Base: "main"}
	main, _, err := s.Clone(t.Context(), repo)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}

	wt, created, err := s.Checkout(t.Context(), repo, main, "feat-a")
	if err != nil {
		t.Fatalf("Checkout: %v", err)
	}
	if !created {
		t.Fatal("first Checkout should create a worktree")
	}

	again, createdAgain, err := s.Checkout(t.Context(), repo, main, "feat-a")
	if err != nil {
		t.Fatalf("second Checkout: %v", err)
	}
	if createdAgain {
		t.Fatal("second Checkout should reuse the worktree")
	}
	if again != wt {
		t.Fatalf("second Checkout = %q, want %q", again, wt)
	}
	if ws := s.git().Worktrees(t.Context(), main); len(ws) != 2 {
		t.Fatalf("want main plus one worktree, got %d", len(ws))
	}

	if err := s.Release(t.Context(), main, wt); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if ws := s.git().Worktrees(t.Context(), main); len(ws) != 1 {
		t.Fatalf("want only main after release, got %d", len(ws))
	}
}

func TestCheckoutUsesTheBaseBranch(t *testing.T) {
	remote := gitRepo(t)
	s := storeFor(t, remote)
	repo := Repo{Owner: "o", Name: "r", Base: "main"}
	main, _, err := s.Clone(t.Context(), repo)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	wt, _, err := s.Checkout(t.Context(), repo, main, "feat-b")
	if err != nil {
		t.Fatalf("Checkout: %v", err)
	}
	if branch, ok := s.git().Branch(t.Context(), wt); !ok || branch != "feat-b" {
		t.Fatalf("worktree branch = %q (ok=%v), want feat-b", branch, ok)
	}
}

func TestReleaseLeavesTheMainCheckout(t *testing.T) {
	declared := t.TempDir()
	runGit(t, declared, "init", "-q", "-b", "main")
	runGit(t, declared, "commit", "-q", "--allow-empty", "-m", "one")
	s := storeFor(t, filepath.Join(t.TempDir(), "never-used"))
	if err := s.Release(t.Context(), declared, declared); err != nil {
		t.Fatalf("Release of the main checkout should be a no-op: %v", err)
	}
	if _, err := os.Stat(filepath.Join(declared, ".git")); err != nil {
		t.Fatalf("main checkout was removed: %v", err)
	}
}

func TestDeclaredRepoMustBeACheckout(t *testing.T) {
	s := storeFor(t, "")
	s.Declared["o-r"] = filepath.Join(t.TempDir(), "not-a-repo")
	if _, _, err := s.Clone(t.Context(), Repo{Owner: "o", Name: "r", Base: "main"}); err == nil {
		t.Fatal("a declared path that is not a repo should be rejected")
	}
}

// gitResolved resolves a path the way git reports it.
func gitResolved(t *testing.T, path string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-parse in %s: %v", path, err)
	}
	return strings.TrimSpace(string(out))
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
