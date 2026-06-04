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

package predicate

import (
	"context"
	"fmt"

	"github.com/palantir/policy-bot/commit"
	"github.com/palantir/policy-bot/policy/common"
	"github.com/palantir/policy-bot/pull"
)

// CommitPredicate is a commit-scoped predicate. It is evaluated against a
// commit.Context and does not require any pull request data.
type CommitPredicate interface {
	common.Triggered

	// EvaluateCommit determines if the predicate is satisfied.
	EvaluateCommit(ctx context.Context, cctx commit.Context) (*common.PredicateResult, error)
}

// PullRequestPredicate is a pull-request-scoped predicate. It requires data
// that only exists in the context of a pull request (for example, the
// author, title, body, comments, reviews, or labels) and is therefore
// evaluated against a pull.Context.
type PullRequestPredicate interface {
	common.Triggered

	// EvaluatePullRequest determines if the predicate is satisfied.
	EvaluatePullRequest(ctx context.Context, prctx pull.Context) (*common.PredicateResult, error)
}

// EvaluatePullRequest dispatches to the correct method based on the
// underlying type of p. It is intended for use by callers that have a
// pull.Context and want to evaluate any predicate without caring which
// interface it implements.
func EvaluatePullRequest(ctx context.Context, p common.Triggered, prctx pull.Context) (*common.PredicateResult, error) {
	switch p := p.(type) {
	case CommitPredicate:
		return p.EvaluateCommit(ctx, prctx)
	case PullRequestPredicate:
		return p.EvaluatePullRequest(ctx, prctx)
	default:
		return nil, fmt.Errorf("unknown predicate type %T", p)
	}
}
