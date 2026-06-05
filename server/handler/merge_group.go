// Copyright 2023 Palantir Technologies, Inc.
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
	"encoding/json"
	"strings"

	"github.com/google/go-github/v85/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/palantir/policy-bot/policy/common"
	"github.com/palantir/policy-bot/pull"
	"github.com/pkg/errors"
)

type MergeGroup struct {
	Base
}

func (h *MergeGroup) Handles() []string { return []string{"merge_group"} }

// Handle merge_group
// https://docs.github.com/webhooks-and-events/webhooks/webhook-events-and-payloads#merge_group
func (h *MergeGroup) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.MergeGroupEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse merge group event payload")
	}

	if event.GetAction() != "checks_requested" {
		return nil
	}

	mg := event.GetMergeGroup()
	headSHA := mg.GetHeadSHA()

	installationID := githubapp.GetInstallationIDFromEvent(&event)
	ctx, logger := githubapp.PrepareRepoContext(ctx, installationID, event.GetRepo())
	logger = logger.With().Str(LogKeyGitHubSHA, headSHA).Logger()
	ctx = logger.WithContext(ctx)

	client, err := h.NewInstallationClient(installationID)
	if err != nil {
		return err
	}
	v4client, err := h.NewInstallationV4Client(installationID)
	if err != nil {
		return err
	}

	repo := event.GetRepo()
	owner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()
	repoID := repo.GetID()
	baseBranch := strings.TrimPrefix(mg.GetBaseRef(), "refs/heads/")
	headBranch := strings.TrimPrefix(mg.GetHeadRef(), "refs/heads/")

	mbrCtx := NewCrossOrgMembershipContext(ctx, client, owner, h.Installations, h.ClientCreator)
	cctx := pull.NewGitHubMergeGroupContext(
		ctx, mbrCtx, h.GlobalCache, client, v4client,
		owner, repoName, repoID,
		mg.GetBaseSHA(), headSHA,
		baseBranch, headBranch,
	)

	fetchedConfig := h.ConfigFetcher.ConfigForRepositoryBranch(ctx, client, owner, repoName, baseBranch)

	evalCtx := &CommitEvalContext{
		Client:   client,
		V4Client: v4client,

		Options:   h.PullOpts,
		PublicURL: h.BaseConfig.PublicURL,

		CommitContext: cctx,
		Config:        fetchedConfig,
	}

	return evalCtx.Evaluate(ctx, common.TriggerCommit)
}
