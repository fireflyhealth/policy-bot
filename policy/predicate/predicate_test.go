// Copyright 2019 Palantir Technologies, Inc.
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
	"testing"

	"github.com/palantir/policy-bot/commit"
	"github.com/palantir/policy-bot/policy/common"
	"github.com/palantir/policy-bot/pull"
	"github.com/palantir/policy-bot/pull/pulltest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertPredicateResult(t *testing.T, expected, actual *common.PredicateResult) {
	assert.Equal(t, expected.Satisfied, actual.Satisfied, "predicate was not correct")
	assert.Equal(t, expected.Values, actual.Values, "values were not correct")
	assert.Equal(t, expected.ConditionsMap, actual.ConditionsMap, "conditions were not correct")
	assert.Equal(t, expected.ConditionValues, actual.ConditionValues, "conditions were not correct")
}

type stubCommitPredicate struct {
	satisfied bool
	called    bool
}

func (s *stubCommitPredicate) Trigger() common.Trigger { return common.TriggerStatic }

func (s *stubCommitPredicate) EvaluateCommit(ctx context.Context, cctx commit.Context) (*common.PredicateResult, error) {
	s.called = true
	return &common.PredicateResult{Satisfied: s.satisfied}, nil
}

type stubPRPredicate struct {
	satisfied bool
	called    bool
}

func (s *stubPRPredicate) Trigger() common.Trigger { return common.TriggerStatic }

func (s *stubPRPredicate) EvaluatePullRequest(ctx context.Context, prctx pull.Context) (*common.PredicateResult, error) {
	s.called = true
	return &common.PredicateResult{Satisfied: s.satisfied}, nil
}

type stubUnknown struct{}

func (s stubUnknown) Trigger() common.Trigger { return common.TriggerStatic }

func TestEvaluatePullRequest(t *testing.T) {
	ctx := context.Background()
	prctx := &pulltest.Context{}

	t.Run("commitPredicate", func(t *testing.T) {
		p := &stubCommitPredicate{satisfied: true}
		res, err := EvaluatePullRequest(ctx, p, prctx)
		require.NoError(t, err)
		assert.True(t, p.called)
		assert.True(t, res.Satisfied)
	})

	t.Run("pullRequestPredicate", func(t *testing.T) {
		p := &stubPRPredicate{satisfied: true}
		res, err := EvaluatePullRequest(ctx, p, prctx)
		require.NoError(t, err)
		assert.True(t, p.called)
		assert.True(t, res.Satisfied)
	})

	t.Run("unknown", func(t *testing.T) {
		_, err := EvaluatePullRequest(ctx, stubUnknown{}, prctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown predicate type")
	})
}

func TestEvaluateCommit(t *testing.T) {
	ctx := context.Background()
	cctx := &pulltest.Context{}

	t.Run("commitPredicate", func(t *testing.T) {
		p := &stubCommitPredicate{satisfied: true}
		res, err := EvaluateCommit(ctx, p, cctx)
		require.NoError(t, err)
		assert.True(t, p.called)
		assert.True(t, res.Satisfied)
	})

	t.Run("pullRequestPredicateErrors", func(t *testing.T) {
		p := &stubPRPredicate{satisfied: true}
		res, err := EvaluateCommit(ctx, p, cctx)
		require.Error(t, err)
		assert.Nil(t, res)
		assert.False(t, p.called)
		assert.Contains(t, err.Error(), "requires pull request data")
	})

	t.Run("unknown", func(t *testing.T) {
		_, err := EvaluateCommit(ctx, stubUnknown{}, cctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown predicate type")
	})
}
