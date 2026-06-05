// Copyright 2021 Palantir Technologies, Inc.
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
	"github.com/palantir/policy-bot/policy/common"
)

type Predicates struct {
	ChangedFiles     *ChangedFiles     `yaml:"changed_files,omitempty"`
	NoChangedFiles   *NoChangedFiles   `yaml:"no_changed_files,omitempty"`
	OnlyChangedFiles *OnlyChangedFiles `yaml:"only_changed_files,omitempty"`
	FileAdded        *FileAdded        `yaml:"file_added,omitempty"`
	FileNotAdded     *FileNotAdded     `yaml:"file_not_added,omitempty"`
	FileDeleted      *FileDeleted      `yaml:"file_deleted,omitempty"`
	FileNotDeleted   *FileNotDeleted   `yaml:"file_not_deleted,omitempty"`

	HasAuthorIn             *HasAuthorIn             `yaml:"has_author_in,omitempty"`
	HasContributorIn        *HasContributorIn        `yaml:"has_contributor_in,omitempty"`
	OnlyHasContributorsIn   *OnlyHasContributorsIn   `yaml:"only_has_contributors_in,omitempty"`
	AuthorIsOnlyContributor *AuthorIsOnlyContributor `yaml:"author_is_only_contributor,omitempty"`

	TargetsBranch *TargetsBranch `yaml:"targets_branch,omitempty"`
	FromBranch    *FromBranch    `yaml:"from_branch,omitempty"`

	ModifiedLines *ModifiedLines `yaml:"modified_lines,omitempty"`

	HasStatus *HasStatus `yaml:"has_status,omitempty"`
	// `has_successful_status` is a deprecated field that is kept for backwards
	// compatibility.  `has_status` replaces it, and can accept any conclusion
	// rather than just "success".
	HasSuccessfulStatus *HasSuccessfulStatus `yaml:"has_successful_status,omitempty"`

	HasWorkflowResult *HasWorkflowResult `yaml:"has_workflow_result,omitempty"`

	HasLabels *HasLabels `yaml:"has_labels,omitempty"`

	Repository *Repository `yaml:"repository,omitempty"`
	Title      *Title      `yaml:"title,omitempty"`

	HasValidSignatures       *HasValidSignatures       `yaml:"has_valid_signatures,omitempty"`
	HasValidSignaturesBy     *HasValidSignaturesBy     `yaml:"has_valid_signatures_by,omitempty"`
	HasValidSignaturesByKeys *HasValidSignaturesByKeys `yaml:"has_valid_signatures_by_keys,omitempty"`

	CustomPropertyIsNull        *CustomPropertyIsNull        `yaml:"custom_property_is_null,omitempty"`
	CustomPropertyIsNotNull     *CustomPropertyIsNotNull     `yaml:"custom_property_is_not_null,omitempty"`
	CustomPropertyMatchesAnyOf  *CustomPropertyMatchesAnyOf  `yaml:"custom_property_matches_any_of,omitempty"`
	CustomPropertyMatchesNoneOf *CustomPropertyMatchesNoneOf `yaml:"custom_property_matches_none_of,omitempty"`
}

func (p Predicates) IsZero() bool {
	return p == Predicates{}
}

// Predicates returns all non-nil predicates as a slice of common.Triggered.
// The concrete elements implement either CommitPredicate (commit-scoped) or
// PullRequestPredicate (pull-request-scoped); use predicate.EvaluatePullRequest to
// dispatch evaluation correctly.
func (p *Predicates) Predicates() []common.Triggered {
	var ps []common.Triggered

	if p.ChangedFiles != nil {
		ps = append(ps, p.ChangedFiles)
	}
	if p.NoChangedFiles != nil {
		ps = append(ps, p.NoChangedFiles)
	}
	if p.OnlyChangedFiles != nil {
		ps = append(ps, p.OnlyChangedFiles)
	}
	if p.FileAdded != nil {
		ps = append(ps, p.FileAdded)
	}
	if p.FileNotAdded != nil {
		ps = append(ps, p.FileNotAdded)
	}
	if p.FileDeleted != nil {
		ps = append(ps, p.FileDeleted)
	}
	if p.FileNotDeleted != nil {
		ps = append(ps, p.FileNotDeleted)
	}

	if p.HasAuthorIn != nil {
		ps = append(ps, p.HasAuthorIn)
	}
	if p.HasContributorIn != nil {
		ps = append(ps, p.HasContributorIn)
	}
	if p.OnlyHasContributorsIn != nil {
		ps = append(ps, p.OnlyHasContributorsIn)
	}
	if p.AuthorIsOnlyContributor != nil {
		ps = append(ps, p.AuthorIsOnlyContributor)
	}

	if p.TargetsBranch != nil {
		ps = append(ps, p.TargetsBranch)
	}
	if p.FromBranch != nil {
		ps = append(ps, p.FromBranch)
	}

	if p.ModifiedLines != nil {
		ps = append(ps, p.ModifiedLines)
	}

	if p.HasStatus != nil {
		ps = append(ps, p.HasStatus)
	}

	if p.HasSuccessfulStatus != nil {
		ps = append(ps, p.HasSuccessfulStatus)
	}

	if p.HasWorkflowResult != nil {
		ps = append(ps, p.HasWorkflowResult)
	}

	if p.HasLabels != nil {
		ps = append(ps, p.HasLabels)
	}

	if p.Repository != nil {
		ps = append(ps, p.Repository)
	}

	if p.Title != nil {
		ps = append(ps, p.Title)
	}

	if p.HasValidSignatures != nil {
		ps = append(ps, p.HasValidSignatures)
	}

	if p.HasValidSignaturesBy != nil {
		ps = append(ps, p.HasValidSignaturesBy)
	}

	if p.HasValidSignaturesByKeys != nil {
		ps = append(ps, p.HasValidSignaturesByKeys)
	}

	if p.CustomPropertyIsNotNull != nil {
		ps = append(ps, p.CustomPropertyIsNotNull)
	}
	if p.CustomPropertyIsNull != nil {
		ps = append(ps, p.CustomPropertyIsNull)
	}
	if p.CustomPropertyMatchesAnyOf != nil {
		ps = append(ps, p.CustomPropertyMatchesAnyOf)
	}
	if p.CustomPropertyMatchesNoneOf != nil {
		ps = append(ps, p.CustomPropertyMatchesNoneOf)
	}

	return ps
}
