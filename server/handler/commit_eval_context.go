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
	"strings"

	"github.com/google/go-github/v85/github"
	"github.com/palantir/policy-bot/commit"
	"github.com/palantir/policy-bot/policy"
	"github.com/palantir/policy-bot/policy/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/shurcooL/githubv4"
)

// CommitEvalContext is the commit-scoped twin of EvalContext. It is used by
// handlers that evaluate policy against a commit.Context (for example, merge
// group webhooks) rather than a pull.Context.
type CommitEvalContext struct {
	Client   *github.Client
	V4Client *githubv4.Client

	Options   *PullEvaluationOptions
	PublicURL string

	CommitContext commit.Context
	Config        FetchedConfig

	// If true, store statuses in the Status field instead of posting them to
	// GitHub. Only the last status is saved, so when this option is enabled,
	// callers should check for a non-nil status after each method call.
	SkipPostStatus bool
	Status         *github.RepoStatus
}

// Evaluate runs the full process for evaluating a commit-scoped event.
func (ec *CommitEvalContext) Evaluate(ctx context.Context, trigger common.Trigger) error {
	evaluator, err := ec.ParseConfig(ctx, trigger)
	if err != nil {
		return err
	}
	if evaluator == nil {
		return nil
	}

	_, err = ec.EvaluatePolicy(ctx, evaluator)
	return err
}

// ParseConfig checks and validates the configuration in the CommitEvalContext
// and returns a non-nil CommitEvaluator if the policy exists, is valid, and
// requires evaluation for the trigger.
func (ec *CommitEvalContext) ParseConfig(ctx context.Context, trigger common.Trigger) (common.CommitEvaluator, error) {
	logger := zerolog.Ctx(ctx)

	fc := ec.Config
	switch {
	case fc.LoadError != nil:
		msg := fmt.Sprintf("Error loading policy from %s", fc.Source)
		logger.Warn().Err(fc.LoadError).Bool("seen_policy", fc.SeenPolicy).Msg(msg)

		if fc.SeenPolicy {
			ec.PostStatus(ctx, "error", msg)
		}
		return nil, errors.Wrapf(fc.LoadError, "failed to load policy: %s: %s", fc.Source, fc.Path)

	case fc.ParseError != nil:
		msg := fmt.Sprintf("Invalid policy in %s: %s", fc.Source, fc.Path)
		logger.Warn().Err(fc.ParseError).Msg(msg)

		ec.PostStatus(ctx, "error", msg)
		return nil, errors.Wrapf(fc.ParseError, "failed to parse policy: %s: %s", fc.Source, fc.Path)

	case fc.Config == nil:
		logger.Debug().Msg("No policy defined for repository")
		return nil, nil
	}

	opts := &policy.GlobalOptions{
		IgnoreEditedComments: ec.Options.IgnoreEditedComments,
		ApprovalDefaults:     ec.Options.ApprovalDefaults,
	}

	evaluator, err := policy.ParsePolicy(fc.Config, opts)
	if err != nil {
		msg := fmt.Sprintf("Invalid policy in %s: %s", fc.Source, fc.Path)
		logger.Warn().Err(err).Msg(msg)

		ec.PostStatus(ctx, "error", msg)
		return nil, errors.Wrapf(err, "failed to create evaluator: %s: %s", fc.Source, fc.Path)
	}

	commitEvaluator, ok := evaluator.(common.CommitEvaluator)
	if !ok {
		msg := fmt.Sprintf("Policy in %s: %s does not support commit-only evaluation", fc.Source, fc.Path)
		ec.PostStatus(ctx, "error", msg)
		return nil, errors.New(msg)
	}

	policyTrigger := commitEvaluator.Trigger()
	if !trigger.Matches(policyTrigger) {
		logger.Debug().
			Str("event_trigger", trigger.String()).
			Str("policy_trigger", policyTrigger.String()).
			Msg("No evaluation necessary for this trigger, skipping")
		return nil, nil
	}

	return commitEvaluator, nil
}

// EvaluatePolicy evaluates the policy for a commit and generates a result.
// The evaluator must be non-nil, meaning callers should check the output of
// ParseConfig before calling this method.
func (ec *CommitEvalContext) EvaluatePolicy(ctx context.Context, evaluator common.CommitEvaluator) (common.Result, error) {
	logger := zerolog.Ctx(ctx)

	result := evaluator.EvaluateCommit(ctx, ec.CommitContext)
	if result.Error != nil {
		msg := fmt.Sprintf("Error evaluating policy in %s: %s", ec.Config.Source, ec.Config.Path)
		logger.Warn().Err(result.Error).Msg(msg)

		ec.PostStatus(ctx, "error", msg)
		return result, result.Error
	}

	statusDescription := result.StatusDescription

	var statusState string
	switch result.Status {
	case common.StatusApproved:
		statusState = "success"
	case common.StatusDisapproved:
		statusState = "failure"
	case common.StatusPending:
		statusState = "pending"
	case common.StatusSkipped:
		statusState = "error"
		statusDescription = "All rules were skipped. At least one rule must match."
	default:
		err := errors.Errorf("Evaluation resulted in unexpected status: %s", result.Status)
		return result, err
	}

	ec.PostStatus(ctx, statusState, statusDescription)
	return result, nil
}

// PostStatus posts a status for the evaluated commit. Unlike the pull request
// version, there is no IsOpen check (merge groups have no equivalent state).
func (ec *CommitEvalContext) PostStatus(ctx context.Context, state, message string) {
	logger := zerolog.Ctx(ctx)

	owner := ec.CommitContext.RepositoryOwner()
	repo := ec.CommitContext.RepositoryName()
	sha := ec.CommitContext.HeadSHA()
	base, _ := ec.CommitContext.Branches()

	publicURL := strings.TrimSuffix(ec.PublicURL, "/")
	detailsURL := fmt.Sprintf("%s/details/%s/%s/commit/%s", publicURL, owner, repo, sha)

	status := github.RepoStatus{
		State:       &state,
		Context:     new(fmt.Sprintf("%s: %s", ec.Options.StatusCheckContext, base)),
		Description: &message,
		TargetURL:   &detailsURL,
	}

	if ec.SkipPostStatus {
		ec.Status = &status
		return
	}

	if err := PostStatus(ctx, ec.Client, owner, repo, sha, status); err != nil {
		logger.Err(err).Msg("Failed to post repo status")
	}
	if ec.Options.PostInsecureStatusChecks {
		status.Context = new(ec.Options.StatusCheckContext)
		if err := PostStatus(ctx, ec.Client, owner, repo, sha, status); err != nil {
			logger.Err(err).Msg("Failed to post insecure repo status")
		}
	}
}
