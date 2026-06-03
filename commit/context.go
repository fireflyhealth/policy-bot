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

package commit

import (
	"time"
)

// Context is the context for a commit on a repository. It defines methods to
// get information about a commit (its files, parents, signature, CI status)
// and about the repository that contains it (collaborators, custom
// properties, team permissions), independent of any pull request that may
// contain the commit.
//
// A new Context should be created for each request, so implementations are
// not required to be thread-safe.
type Context interface {
	MembershipContext

	// EvaluationTimestamp returns the time at the start of the evaluation,
	// usually the creation time of the context. All calls on the same context
	// should return the same value.
	EvaluationTimestamp() time.Time

	// RepositoryOwner returns the owner of the repo that contains the commit.
	RepositoryOwner() string

	// RepositoryName returns the name of the repo that contains the commit.
	RepositoryName() string

	// RepositoryCustomProperties returns the custom properties of the repo
	// that contains the commit. For an unset property, the key is _not_
	// present in the map.
	RepositoryCustomProperties() (map[string]CustomProperty, error)

	// HeadSHA returns the SHA of the head commit being evaluated.
	HeadSHA() string

	// Branches returns the base (also known as target) and head branch names
	// associated with the commit. Branches in this repository have no prefix,
	// while branches in forks are prefixed with the owner of the fork and a
	// colon. The base branch will always be unprefixed.
	Branches() (base string, head string)

	// ChangedFiles returns the files that were changed by the commit (or by
	// the set of commits, for contexts that group multiple commits).
	ChangedFiles() ([]*File, error)

	// Commits returns the commits being evaluated. The commit order is
	// implementation dependent.
	Commits() ([]*Commit, error)

	// PushedAt returns the time at which the commit with sha was pushed. The
	// returned time may be after the actual push time, but must not be
	// before.
	PushedAt(sha string) (time.Time, error)

	// RepositoryCollaborators returns the repository collaborators. Filters
	// to collaborators with at least the specified permission level.
	// Filtering by permission can significantly improve performance.
	RepositoryCollaborators(minPermission Permission) ([]*Collaborator, error)

	// CollaboratorPermission returns the permission level of user on the
	// repository.
	CollaboratorPermission(user string) (Permission, error)

	// Teams lists the set of team collaborators, along with their respective
	// permission on the repo.
	Teams() (map[string]Permission, error)

	// LatestStatuses returns a map of status check names to the latest
	// result for the head commit.
	LatestStatuses() (map[string]string, error)

	// LatestWorkflowRuns returns the latest GitHub Actions workflow runs for
	// the head commit. The keys of the map are paths to the workflow files
	// and the values are the conclusions of the latest runs, one per event
	// type.
	LatestWorkflowRuns() (map[string][]string, error)
}
