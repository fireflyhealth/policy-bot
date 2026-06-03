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
	"net/http"
	"net/url"
	"testing"

	"github.com/google/go-github/v85/github"
	"github.com/palantir/policy-bot/commit"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testMergeGroupBaseSHA = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	testMergeGroupHeadSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testMergeGroupMidSHA  = "cccccccccccccccccccccccccccccccccccccccc"
)

func TestMergeGroupBranches(t *testing.T) {
	ctx := makeMergeGroupContext(t, &ResponsePlayer{})

	base, head := ctx.Branches()
	assert.Equal(t, "main", base)
	assert.Equal(t, "gh-readonly-queue/main/pr-1-abc", head)
	assert.Equal(t, testMergeGroupHeadSHA, ctx.HeadSHA())
}

func TestMergeGroupChangedFiles(t *testing.T) {
	rp := &ResponsePlayer{}
	filesRule := rp.AddRule(
		ExactPathMatcher("/repos/testorg/testrepo/compare/"+testMergeGroupBaseSHA+"..."+testMergeGroupHeadSHA),
		"testdata/responses/merge_group_compare.yml",
	)

	ctx := makeMergeGroupContext(t, rp)

	files, err := ctx.ChangedFiles()
	require.NoError(t, err)

	require.Len(t, files, 5, "rename should expand to a delete + add pair")
	assert.Equal(t, 1, filesRule.Count, "no http request was made")

	assert.Equal(t, "path/foo.txt", files[0].Filename)
	assert.Equal(t, commit.FileAdded, files[0].Status)

	assert.Equal(t, "path/bar.txt", files[1].Filename)
	assert.Equal(t, commit.FileDeleted, files[1].Status)

	assert.Equal(t, "README.md", files[2].Filename)
	assert.Equal(t, commit.FileModified, files[2].Status)

	assert.Equal(t, "path/old.txt", files[3].Filename)
	assert.Equal(t, commit.FileDeleted, files[3].Status)
	assert.Equal(t, 0, files[3].Additions)
	assert.Equal(t, 0, files[3].Deletions)

	assert.Equal(t, "path/new.txt", files[4].Filename)
	assert.Equal(t, commit.FileAdded, files[4].Status)
	assert.Equal(t, 2, files[4].Additions)
	assert.Equal(t, 4, files[4].Deletions)

	files, err = ctx.ChangedFiles()
	require.NoError(t, err)
	require.Len(t, files, 5)
	assert.Equal(t, 1, filesRule.Count, "cached files were not used")
}

func TestMergeGroupCommits(t *testing.T) {
	rp := &ResponsePlayer{}
	historyRule := rp.AddRule(
		GraphQLNodePrefixMatcher("repository.object.Commit.history"),
		"testdata/responses/merge_group_history.yml",
	)

	ctx := makeMergeGroupContext(t, rp)

	commits, err := ctx.Commits()
	require.NoError(t, err)

	require.Len(t, commits, 2, "base commit should be excluded from the result")
	assert.Equal(t, 1, historyRule.Count, "no http request was made")

	assert.Equal(t, testMergeGroupHeadSHA, commits[0].SHA)
	assert.Equal(t, "mhaypenny", commits[0].Author)
	assert.Equal(t, "mhaypenny", commits[0].Committer)
	assert.Nil(t, commits[0].Signature)

	assert.Equal(t, testMergeGroupMidSHA, commits[1].SHA)
	assert.Equal(t, "ttest", commits[1].Author)
	assert.Equal(t, "mhaypenny", commits[1].Committer)
	require.NotNil(t, commits[1].Signature)
	assert.Equal(t, "3AA5C34371567BD2", commits[1].Signature.KeyID)
	assert.Equal(t, "ttest", commits[1].Signature.Signer)
	assert.True(t, commits[1].Signature.IsValid)

	commits, err = ctx.Commits()
	require.NoError(t, err)
	require.Len(t, commits, 2)
	assert.Equal(t, 1, historyRule.Count, "cached commits were not used")
}

func TestMergeGroupCommitsBaseNotReached(t *testing.T) {
	rp := &ResponsePlayer{}
	rp.AddRule(
		GraphQLNodePrefixMatcher("repository.object.Commit.history"),
		"testdata/responses/merge_group_history_no_base.yml",
	)

	ctx := makeMergeGroupContext(t, rp)

	_, err := ctx.Commits()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not an ancestor")
}

func makeMergeGroupContext(t *testing.T, rp *ResponsePlayer) commit.Context {
	ctx := context.Background()
	client := github.NewClient(&http.Client{Transport: rp})
	v4client := githubv4.NewClient(&http.Client{Transport: rp})

	base, _ := url.Parse("http://github.localhost/")
	client.BaseURL = base

	mbrCtx := NewGitHubMembershipContext(ctx, client)
	return NewGitHubMergeGroupContext(
		ctx, mbrCtx, NewMockGlobalCache(), client, v4client,
		"testorg", "testrepo", 1234,
		testMergeGroupBaseSHA, testMergeGroupHeadSHA,
		"main", "gh-readonly-queue/main/pr-1-abc",
	)
}
