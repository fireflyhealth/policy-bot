// Copyright 2018 Palantir Technologies, Inc.
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

package handler

import (
	"context"

	"github.com/google/go-github/v85/github"
	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/palantir/policy-bot/policy/common"
	"github.com/palantir/policy-bot/pull"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	LogKeyGitHubSHA = "github_sha"
)

type Base struct {
	githubapp.ClientCreator

	Installations githubapp.InstallationsService
	GlobalCache   pull.GlobalCache
	ConfigFetcher *ConfigFetcher
	BaseConfig    *baseapp.HTTPConfig
	PullOpts      *PullEvaluationOptions

	AppName string
}

// PostStatus posts a GitHub commit status with consistent logging.
func PostStatus(ctx context.Context, client *github.Client, owner, repo, ref string, status github.RepoStatus) error {
	zerolog.Ctx(ctx).Info().Msgf("Setting %q status on %s to %s: %s", status.GetContext(), ref, status.GetState(), status.GetDescription())
	_, _, err := client.Repositories.CreateStatus(ctx, owner, repo, ref, status)
	return errors.WithStack(err)
}

func (b *Base) PreparePRContext(ctx context.Context, installationID int64, pr *github.PullRequest) (context.Context, zerolog.Logger) {
	ctx, logger := githubapp.PreparePRContext(ctx, installationID, pr.GetBase().GetRepo(), pr.GetNumber())

	logger = logger.With().Str(LogKeyGitHubSHA, pr.GetHead().GetSHA()).Logger()
	ctx = logger.WithContext(ctx)

	return ctx, logger
}

func (b *Base) NewEvalContext(ctx context.Context, installationID int64, loc pull.Locator) (*EvalContext, error) {
	client, err := b.NewInstallationClient(installationID)
	if err != nil {
		return nil, err
	}

	v4client, err := b.NewInstallationV4Client(installationID)
	if err != nil {
		return nil, err
	}

	mbrCtx := NewCrossOrgMembershipContext(ctx, client, loc.Owner, b.Installations, b.ClientCreator)
	prctx, err := pull.NewGitHubPullRequestContext(ctx, mbrCtx, b.GlobalCache, client, v4client, loc)
	if err != nil {
		return nil, err
	}

	baseBranch, _ := prctx.Branches()
	owner := prctx.RepositoryOwner()
	repository := prctx.RepositoryName()

	fetchedConfig := b.ConfigFetcher.ConfigForRepositoryBranch(ctx, client, owner, repository, baseBranch)

	return &EvalContext{
		Client:   client,
		V4Client: v4client,

		Options:   b.PullOpts,
		PublicURL: b.BaseConfig.PublicURL,

		PullContext: prctx,
		Config:      fetchedConfig,
	}, nil
}

func (b *Base) Evaluate(ctx context.Context, installationID int64, trigger common.Trigger, loc pull.Locator) error {
	evalCtx, err := b.NewEvalContext(ctx, installationID, loc)
	if err != nil {
		return errors.Wrap(err, "failed to create evaluation context")
	}
	return evalCtx.Evaluate(ctx, trigger)
}

// EvaluateMergeGroupCommit checks whether sha is the head of a merge queue
// branch in repo and, if so, evaluates the policy for that commit. It is a
// no-op when the commit is not associated with a merge group. Callers should
// invoke this from event handlers that observe commit-level changes (e.g.
// status, check_run) so merge group evaluations stay current.
func (b *Base) EvaluateMergeGroupCommit(ctx context.Context, installationID int64, repo *github.Repository, sha string, trigger common.Trigger) error {
	client, err := b.NewInstallationClient(installationID)
	if err != nil {
		return err
	}

	owner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()

	mqRef, err := findMergeQueueRef(ctx, client, owner, repoName, sha)
	if err != nil {
		return errors.Wrap(err, "failed to look up branches for commit")
	}
	if mqRef == nil {
		return nil
	}

	v4client, err := b.NewInstallationV4Client(installationID)
	if err != nil {
		return err
	}

	mbrCtx := NewCrossOrgMembershipContext(ctx, client, owner, b.Installations, b.ClientCreator)
	cctx := pull.NewGitHubMergeGroupContext(
		ctx, mbrCtx, b.GlobalCache, client, v4client,
		owner, repoName, repo.GetID(),
		mqRef.BaseSHA, sha,
		mqRef.BaseBranch, mqRef.HeadBranch,
	)

	fetchedConfig := b.ConfigFetcher.ConfigForRepositoryBranch(ctx, client, owner, repoName, mqRef.BaseBranch)

	evalCtx := &CommitEvalContext{
		Client:   client,
		V4Client: v4client,

		Options:   b.PullOpts,
		PublicURL: b.BaseConfig.PublicURL,

		CommitContext: cctx,
		Config:        fetchedConfig,
	}

	return evalCtx.Evaluate(ctx, trigger)
}
