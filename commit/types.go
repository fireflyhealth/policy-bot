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

// Package commit defines data types and interfaces describing properties of
// a commit on a repository, independent of any pull request that may contain
// that commit. These types are shared by pull request contexts and other
// commit-scoped contexts (for example, GitHub merge groups).
package commit

// MembershipContext defines methods to get information
// about about user membership in Github organizations and teams.
type MembershipContext interface {
	// IsTeamMember returns true if the user is a member of the given team.
	// Teams are specified as "org-name/team-name".
	IsTeamMember(team, user string) (bool, error)

	// IsOrgMember returns true if the user is a member of the given organzation.
	IsOrgMember(org, user string) (bool, error)

	// TeamMembers returns the list of usernames in the given organization's team.
	TeamMembers(team string) ([]string, error)

	// OrganizationMembers returns the list of org member usernames in the given organization.
	OrganizationMembers(org string) ([]string, error)
}

type FileStatus int

const (
	FileModified FileStatus = iota
	FileAdded
	FileDeleted
)

type File struct {
	Filename  string
	Status    FileStatus
	Additions int
	Deletions int
}

type Commit struct {
	SHA             string
	Parents         []string
	CommittedViaWeb bool

	// Author is the login name of the author. It is empty if the author is not
	// a real user.
	Author string

	// Commiter is the login name of the committer. It is empty if the
	// committer is not a real user.
	Committer string

	// Signature is the signature and details that was extracted from the commit.
	// It is nil if the commit has no signature
	Signature *Signature
}

// Users returns the login names of the users associated with this commit.
func (c *Commit) Users() []string {
	var users []string
	if c.Author != "" {
		users = append(users, c.Author)
	}
	if c.Committer != "" {
		users = append(users, c.Committer)
	}
	return users
}

type SignatureType string

const (
	SignatureGpg   SignatureType = "GpgSignature"
	SignatureSmime SignatureType = "SmimeSignature"
	SignatureSSH   SignatureType = "SshSignature"
)

type Signature struct {
	Type           SignatureType
	IsValid        bool
	KeyID          string
	KeyFingerprint string
	Signer         string
	State          string
}

type Collaborator struct {
	Name        string
	Permissions []CollaboratorPermission
}

type CollaboratorPermission struct {
	Permission Permission

	// True if Permission is granted by a direct or team association with the
	// repository. If false, the permission is granted by the organization.
	ViaRepo bool
}

type CustomProperty struct {
	String *string
	Array  []string
}
