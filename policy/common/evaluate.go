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

package common

import (
	"context"

	"github.com/palantir/policy-bot/commit"
	"github.com/palantir/policy-bot/pull"
)

// PullRequestEvaluator evaluates a policy against a pull.Context. It is
// implemented by evaluators that require pull request data (for example,
// reviews, comments, labels, or approval candidates).
type PullRequestEvaluator interface {
	Triggered

	EvaluatePullRequest(ctx context.Context, prctx pull.Context) Result
}

// CommitEvaluator evaluates a policy against a commit.Context, without any
// pull request data. It is implemented by evaluators that can run on the
// commit-scoped subset of a policy (for example, for merge group events).
// Evaluators that require pull request data should set Error on the returned
// Result rather than implementing this interface in a way that silently
// produces misleading output.
type CommitEvaluator interface {
	Triggered

	EvaluateCommit(ctx context.Context, cctx commit.Context) Result
}
