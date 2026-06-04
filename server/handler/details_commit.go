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

package handler

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/alexedwards/scs"
	"github.com/bluekeyes/hatpear"
	"github.com/bluekeyes/templatetree"
	"github.com/google/go-github/v85/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/palantir/policy-bot/policy/common"
	"github.com/palantir/policy-bot/pull"
	"github.com/pkg/errors"
	"goji.io/pat"
)

// DetailsCommit renders the policy evaluation details page for a commit
// identified by SHA. The current implementation supports commits that are the
// head of a GitHub merge queue branch; other commits are reported as not
// found.
type DetailsCommit struct {
	Base
	Sessions  *scs.Manager
	Templates templatetree.Tree[*template.Template]
}

func (h *DetailsCommit) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	owner := pat.Param(r, "owner")
	repo := pat.Param(r, "repo")
	sha := pat.Param(r, "sha")
	if owner == "" || repo == "" || sha == "" {
		http.Error(w, "Invalid commit", http.StatusBadRequest)
		return nil
	}

	installation, err := h.Installations.GetByOwner(ctx, owner)
	if err != nil {
		if _, notFound := err.(githubapp.InstallationNotFound); notFound {
			h.renderCommit404(w, owner, repo, sha)
		} else {
			hatpear.Store(r, err)
		}
		return nil
	}

	client, err := h.ClientCreator.NewInstallationClient(installation.ID)
	if err != nil {
		hatpear.Store(r, errors.Wrap(err, "failed to create github client"))
		return nil
	}

	user, hasPermission, err := checkUserPermissions(h.Sessions, r, client, owner, repo)
	if err != nil {
		hatpear.Store(r, err)
		return nil
	}
	if !hasPermission {
		h.renderCommit404(w, owner, repo, sha)
		return nil
	}

	mqRef, err := findMergeQueueRef(ctx, client, owner, repo, sha)
	if err != nil {
		if isNotFound(err) {
			h.renderCommit404(w, owner, repo, sha)
			return nil
		}
		hatpear.Store(r, errors.Wrap(err, "failed to look up branches for commit"))
		return nil
	}
	if mqRef == nil {
		h.renderCommit404(w, owner, repo, sha)
		return nil
	}

	v4client, err := h.ClientCreator.NewInstallationV4Client(installation.ID)
	if err != nil {
		hatpear.Store(r, errors.Wrap(err, "failed to create github v4 client"))
		return nil
	}

	ghRepo, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		hatpear.Store(r, errors.Wrap(err, "failed to load repository"))
		return nil
	}

	ctx, logger := githubapp.PrepareRepoContext(ctx, installation.ID, ghRepo)
	logger = logger.With().Str(LogKeyGitHubSHA, sha).Logger()
	ctx = logger.WithContext(ctx)

	mbrCtx := NewCrossOrgMembershipContext(ctx, client, owner, h.Installations, h.ClientCreator)
	cctx := pull.NewGitHubMergeGroupContext(
		ctx, mbrCtx, h.GlobalCache, client, v4client,
		owner, repo, ghRepo.GetID(),
		mqRef.BaseSHA, sha,
		mqRef.BaseBranch, mqRef.HeadBranch,
	)

	fetchedConfig := h.ConfigFetcher.ConfigForRepositoryBranch(ctx, client, owner, repo, mqRef.BaseBranch)

	evalCtx := &CommitEvalContext{
		Client:        client,
		V4Client:      v4client,
		Options:       h.PullOpts,
		PublicURL:     h.BaseConfig.PublicURL,
		CommitContext: cctx,
		Config:        fetchedConfig,
	}

	data := detailsPageData{
		BasePath:       getBasePath(h.BaseConfig.PublicURL),
		User:           user,
		PolicyURL:      getPolicyURL(ghRepo.GetHTMLURL(), fetchedConfig),
		PageTitle:      fmt.Sprintf("%s@%s", ghRepo.GetFullName(), shortSHA(sha)),
		RepoFullName:   ghRepo.GetFullName(),
		BaseRef:        mqRef.BaseBranch,
		HeadingHref:    fmt.Sprintf("%s/commit/%s", ghRepo.GetHTMLURL(), sha),
		HeadingText:    shortSHA(sha),
		HeadingTooltip: "View the commit on GitHub",
		Title:          fmt.Sprintf("Merge group for #%d", mqRef.PRNumber),
	}

	evaluator, err := evalCtx.ParseConfig(ctx, common.TriggerAll)
	if err != nil {
		data.Error = err
		return h.render(w, data)
	}
	if evaluator == nil {
		data.Error = errors.Errorf("Invalid policy at %s: %s", evalCtx.Config.Source, evalCtx.Config.Path)
		return h.render(w, data)
	}

	result, err := evalCtx.EvaluatePolicy(ctx, evaluator)
	data.Result = &result

	if err != nil {
		if _, ok := errors.Cause(err).(*pull.TemporaryError); ok {
			data.IsTemporaryError = true
		}
		data.Error = err
	}

	return h.render(w, data)
}

func (h *DetailsCommit) render(w http.ResponseWriter, data detailsPageData) error {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	return h.Templates.ExecuteTemplate(w, "details.html.tmpl", data)
}

func (h *DetailsCommit) renderCommit404(w http.ResponseWriter, owner, repo, sha string) {
	msg := fmt.Sprintf(
		"Not Found: %s/%s@%s\n\nThe repository or commit does not exist, you do not have permission, or policy-bot is not installed.",
		owner, repo, sha,
	)
	http.Error(w, msg, http.StatusNotFound)
}

// findMergeQueueRef returns the merge queue ref that points at sha as its
// head, if any exists. A nil result with a nil error means the commit is not
// the head of any merge queue branch.
func findMergeQueueRef(ctx context.Context, client *github.Client, owner, repo, sha string) (*pull.MergeQueueRef, error) {
	branches, _, err := client.Repositories.ListBranchesHeadCommit(ctx, owner, repo, sha)
	if err != nil {
		return nil, err
	}
	for _, b := range branches {
		if ref, ok := pull.ParseMergeQueueRef(b.GetName()); ok {
			return &ref, nil
		}
	}
	return nil, nil
}

// shortSHA returns the conventional short form of a commit SHA. If sha is
// already short, it is returned unchanged.
func shortSHA(sha string) string {
	const shortLen = 7
	if len(sha) <= shortLen {
		return sha
	}
	return sha[:shortLen]
}
