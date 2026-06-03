// Copyright 2026 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pull

import (
	"context"
	"time"

	"github.com/google/go-github/v85/github"
	"github.com/palantir/policy-bot/commit"
	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
)

// GitHubMergeGroupContext is a commit.Context implementation that gets
// information from GitHub for the head commit of a merge group. A new
// instance must be created for each request.
type GitHubMergeGroupContext struct {
	*GitHubCommitContext

	baseSHA    string
	baseBranch string
	headBranch string

	// cached fields
	files   []*commit.File
	commits []*commit.Commit
}

var _ commit.Context = (*GitHubMergeGroupContext)(nil)

// NewGitHubMergeGroupContext creates a new commit.Context that makes GitHub
// requests to obtain information about a merge group. It caches responses for
// the lifetime of the context. baseBranch and headBranch should be the
// unprefixed branch names (without the leading "refs/heads/").
func NewGitHubMergeGroupContext(
	ctx context.Context,
	mbrCtx commit.MembershipContext,
	globalCache GlobalCache,
	client *github.Client,
	v4client *githubv4.Client,
	owner, repo string,
	repoID int64,
	baseSHA, headSHA string,
	baseBranch, headBranch string,
) commit.Context {
	return &GitHubMergeGroupContext{
		GitHubCommitContext: newGitHubCommitContext(
			ctx, mbrCtx, globalCache, client, v4client,
			owner, repo, repoID, headSHA,
		),
		baseSHA:    baseSHA,
		baseBranch: baseBranch,
		headBranch: headBranch,
	}
}

func (mgc *GitHubMergeGroupContext) Branches() (base string, head string) {
	return mgc.baseBranch, mgc.headBranch
}

func (mgc *GitHubMergeGroupContext) ChangedFiles() ([]*commit.File, error) {
	if mgc.files == nil {
		if err := mgc.loadChangedFiles(); err != nil {
			return nil, err
		}
	}
	return mgc.files, nil
}

func (mgc *GitHubMergeGroupContext) Commits() ([]*commit.Commit, error) {
	if mgc.commits == nil {
		if err := mgc.loadCommits(); err != nil {
			return nil, err
		}
	}
	return mgc.commits, nil
}

// PushedAt for a merge group looks up the push time for sha using only the
// shared single-sha sources (local cache, global cache, GitHub status
// timestamps). Unlike the pull request implementation, there is no
// candidate-commit walk: merge group contexts do not have a notion of a
// pending batch of pushed commits beyond the head.
func (mgc *GitHubMergeGroupContext) PushedAt(sha string) (time.Time, error) {
	if mgc.pushedAt == nil {
		mgc.pushedAt = make(map[string]time.Time)
	}

	pushedAt, err := mgc.tryPushedAt(sha)
	if err != nil {
		return time.Time{}, err
	}
	if pushedAt.IsZero() {
		pushedAt = mgc.EvaluationTimestamp()
	}

	mgc.pushedAt[sha] = pushedAt
	if gc := mgc.globalCache; gc != nil {
		gc.SetPushedAt(mgc.repoID, sha, pushedAt)
	}
	return pushedAt, nil
}

// loadChangedFiles fetches the files changed between baseSHA and headSHA via
// the REST compare API. The REST endpoint is used because file information
// does not depend on signature data, so the simpler API is sufficient.
func (mgc *GitHubMergeGroupContext) loadChangedFiles() error {
	opt := &github.ListOptions{PerPage: 100}

	var allFiles []*github.CommitFile
	for {
		cmp, resp, err := mgc.client.Repositories.CompareCommits(
			mgc.ctx, mgc.owner, mgc.repo, mgc.baseSHA, mgc.HeadSHA(), opt,
		)
		if err != nil {
			return errors.Wrap(err, "failed to compare commits for merge group")
		}
		allFiles = append(allFiles, cmp.Files...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	if len(allFiles) > MaxPullRequestFiles {
		return errors.Errorf("number of changed files (%d) exceeds limit (%d)", len(allFiles), MaxPullRequestFiles)
	}

	mgc.files = make([]*commit.File, 0, len(allFiles))
	for _, f := range allFiles {
		status := commit.FileModified
		switch f.GetStatus() {
		case "added":
			status = commit.FileAdded
		case "removed":
			status = commit.FileDeleted
		case "renamed":
			// Break renames into components: the new file is added and we
			// generate an extra entry for the old file that is deleted.
			// Attribute all modifications to the new file to avoid double
			// counting.
			status = commit.FileAdded
			mgc.files = append(mgc.files, &commit.File{
				Filename:  f.GetPreviousFilename(),
				Status:    commit.FileDeleted,
				Additions: 0,
				Deletions: 0,
			})
		}

		mgc.files = append(mgc.files, &commit.File{
			Filename:  f.GetFilename(),
			Status:    status,
			Additions: f.GetAdditions(),
			Deletions: f.GetDeletions(),
		})
	}
	return nil
}

// loadCommits walks the commit history starting from headSHA back through
// ancestors until baseSHA is reached. The base commit itself is excluded.
// GraphQL is used so that each commit carries the same author, committer,
// and signature data as commits loaded for a pull request.
func (mgc *GitHubMergeGroupContext) loadCommits() error {
	var q struct {
		Repository struct {
			Object struct {
				Commit struct {
					History struct {
						PageInfo v4PageInfo
						Nodes    []*v4Commit
					} `graphql:"history(first: 100, after: $cursor)"`
				} `graphql:"... on Commit"`
			} `graphql:"object(oid: $head)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	qvars := map[string]any{
		"owner":  githubv4.String(mgc.owner),
		"name":   githubv4.String(mgc.repo),
		"head":   githubv4.GitObjectID(mgc.HeadSHA()),
		"cursor": (*githubv4.String)(nil),
	}

	var commits []*commit.Commit
	for {
		if err := mgc.v4client.Query(mgc.ctx, &q, qvars); err != nil {
			return errors.Wrap(err, "failed to load merge group commits")
		}
		for _, raw := range q.Repository.Object.Commit.History.Nodes {
			if raw.OID == mgc.baseSHA {
				mgc.commits = commits
				return nil
			}
			commits = append(commits, raw.ToCommit())
			if len(commits) > MaxPullRequestCommits {
				return errors.Errorf("too many commits in merge group, maximum is %d", MaxPullRequestCommits)
			}
		}
		if !q.Repository.Object.Commit.History.PageInfo.UpdateCursor(qvars, "cursor") {
			break
		}
	}

	// The history was exhausted without reaching baseSHA. This should not
	// happen for a well-formed merge group, but treat it as an error rather
	// than silently returning ancestors of the base.
	return errors.Errorf("merge group base commit %.10s is not an ancestor of head commit %.10s", mgc.baseSHA, mgc.HeadSHA())
}
