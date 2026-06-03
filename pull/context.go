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

package pull

import (
	"time"

	"github.com/palantir/policy-bot/commit"
)

// Context is the context for a pull request. It extends commit.Context with
// methods that return information that only exists in the context of a pull
// request (the PR's identity, author, comments, reviews, requested
// reviewers, labels, and lifecycle state).
//
// A new Context should be created for each request, so implementations are not
// required to be thread-safe.
type Context interface {
	commit.Context

	// Number returns the number of the pull request.
	Number() int

	// Title returns the title of the pull request
	Title() string

	// Body returns a struct that includes LastEditedAt for the pull request body
	Body() (*Body, error)

	// Author returns the username of the user who opened the pull request.
	Author() string

	// CreatedAt returns the time when the pull request was created.
	CreatedAt() time.Time

	// IsOpen returns true when the state of the pull request is "open"
	IsOpen() bool

	// IsClosed returns true when the state of the pull request is "closed"
	IsClosed() bool

	// Comments lists all comments on a Pull Request. The comment order is
	// implementation dependent.
	Comments() ([]*Comment, error)

	// Reviews lists all reviews on a Pull Request. The review order is
	// implementation dependent.
	Reviews() ([]*Review, error)

	// IsDraft returns the draft status of the Pull Request.
	IsDraft() bool

	// RequestedReviewers returns any current and dismissed review requests on
	// the pull request.
	RequestedReviewers() ([]*Reviewer, error)

	// Labels returns a list of labels applied on the Pull Request
	Labels() ([]string, error)
}

type Comment struct {
	CreatedAt    time.Time
	LastEditedAt time.Time
	Author       string
	Body         string
}

type ReviewState string

const (
	ReviewApproved         ReviewState = "approved"
	ReviewChangesRequested ReviewState = "changes_requested"
	ReviewCommented        ReviewState = "commented"
	ReviewDismissed        ReviewState = "dismissed"
	ReviewPending          ReviewState = "pending"
)

type Review struct {
	ID           string
	CreatedAt    time.Time
	LastEditedAt time.Time
	Author       string
	State        ReviewState
	Body         string
	SHA          string

	Teams []string
}

type ReviewerType string

const (
	ReviewerUser ReviewerType = "user"
	ReviewerTeam ReviewerType = "team"
)

type Reviewer struct {
	Type    ReviewerType
	Name    string
	Removed bool
}

type Body struct {
	Body         string
	CreatedAt    time.Time
	Author       string
	LastEditedAt time.Time
}
