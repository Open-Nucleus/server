package gitstore

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Store provides Git operations for the clinical data repository.
type Store interface {
	// WriteAndCommit writes a file and creates a commit. Returns commit hash.
	WriteAndCommit(path string, data []byte, msg CommitMessage) (string, error)
	// Read returns file contents at HEAD.
	Read(path string) ([]byte, error)
	// LogPath returns commit history for a path prefix.
	LogPath(pathPrefix string, limit int) ([]CommitInfo, error)
	// Head returns current HEAD commit hash.
	Head() (string, error)
	// TreeWalk calls fn for every file in the tree at HEAD.
	TreeWalk(fn func(path string, data []byte) error) error
	// Rollback resets the working tree to HEAD (for error recovery).
	Rollback() error
}

type gitStore struct {
	repo       *git.Repository
	worktree   *git.Worktree
	repoPath   string
	authorName string
	authorEmail string
}

// NewStore opens or initialises a Git repository at repoPath.
func NewStore(repoPath, authorName, authorEmail string) (Store, error) {
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		return nil, fmt.Errorf("create repo dir: %w", err)
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		// Init new repo
		repo, err = git.PlainInit(repoPath, false)
		if err != nil {
			return nil, fmt.Errorf("git init: %w", err)
		}
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("worktree: %w", err)
	}

	return &gitStore{
		repo:       repo,
		worktree:   wt,
		repoPath:   repoPath,
		authorName: authorName,
		authorEmail: authorEmail,
	}, nil
}

func (s *gitStore) WriteAndCommit(path string, data []byte, msg CommitMessage) (string, error) {
	fullPath := filepath.Join(s.repoPath, path)

	// Ensure parent directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create dir %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	// Stage file
	if _, err := s.worktree.Add(path); err != nil {
		return "", fmt.Errorf("git add: %w", err)
	}

	// Commit
	commitHash, err := s.worktree.Commit(msg.Format(), &git.CommitOptions{
		Author: &object.Signature{
			Name:  s.authorName,
			Email: s.authorEmail,
			When:  msg.Timestamp,
		},
	})
	if err != nil {
		return "", fmt.Errorf("git commit: %w", err)
	}

	return commitHash.String(), nil
}

func (s *gitStore) Read(path string) ([]byte, error) {
	fullPath := filepath.Join(s.repoPath, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return data, nil
}

func (s *gitStore) LogPath(pathPrefix string, limit int) ([]CommitInfo, error) {
	ref, err := s.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("get HEAD: %w", err)
	}

	logOpts := &git.LogOptions{
		From: ref.Hash(),
	}
	if pathPrefix != "" {
		logOpts.PathFilter = func(p string) bool {
			return len(p) >= len(pathPrefix) && p[:len(pathPrefix)] == pathPrefix
		}
	}

	iter, err := s.repo.Log(logOpts)
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	var infos []CommitInfo
	err = iter.ForEach(func(c *object.Commit) error {
		if limit > 0 && len(infos) >= limit {
			return fmt.Errorf("limit reached")
		}
		infos = append(infos, CommitInfo{
			Hash:      c.Hash.String(),
			Timestamp: c.Author.When,
			Message:   c.Message,
		})
		return nil
	})
	// "limit reached" is not a real error
	if err != nil && err.Error() != "limit reached" {
		return nil, err
	}

	return infos, nil
}

func (s *gitStore) Head() (string, error) {
	ref, err := s.repo.Head()
	if err != nil {
		if err == plumbing.ErrReferenceNotFound {
			return "", nil // empty repo
		}
		return "", fmt.Errorf("get HEAD: %w", err)
	}
	return ref.Hash().String(), nil
}

func (s *gitStore) TreeWalk(fn func(path string, data []byte) error) error {
	ref, err := s.repo.Head()
	if err != nil {
		if err == plumbing.ErrReferenceNotFound {
			return nil // empty repo, nothing to walk
		}
		return fmt.Errorf("get HEAD: %w", err)
	}

	commit, err := s.repo.CommitObject(ref.Hash())
	if err != nil {
		return fmt.Errorf("get commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return fmt.Errorf("get tree: %w", err)
	}

	return tree.Files().ForEach(func(f *object.File) error {
		contents, err := f.Contents()
		if err != nil {
			return fmt.Errorf("read file %s: %w", f.Name, err)
		}
		return fn(f.Name, []byte(contents))
	})
}

func (s *gitStore) Rollback() error {
	return s.worktree.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})
}

// TimeNow is exposed for testing; production code should use time.Now().
var TimeNow = time.Now
