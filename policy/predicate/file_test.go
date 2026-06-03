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
	"regexp"
	"testing"

	"github.com/palantir/policy-bot/commit"
	"github.com/palantir/policy-bot/policy/common"
	"github.com/palantir/policy-bot/pull/pulltest"
	"github.com/stretchr/testify/assert"
)

func TestChangedFiles(t *testing.T) {
	p := &ChangedFiles{
		Paths: []common.Regexp{
			common.NewCompiledRegexp(regexp.MustCompile("app/.*\\.go")),
			common.NewCompiledRegexp(regexp.MustCompile("server/.*\\.go")),
		},
		IgnorePaths: []common.Regexp{
			common.NewCompiledRegexp(regexp.MustCompile(".*/special\\.go")),
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"empty",
			[]*commit.File{},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/.*\\.go", "server/.*\\.go"},
					"while ignoring": {".*/special\\.go"},
				},
			},
		},
		{
			"onlyMatches",
			[]*commit.File{
				{
					Filename: "app/client.go",
					Status:   commit.FileAdded,
				},
				{
					Filename: "server/server.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"app/client.go"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/.*\\.go", "server/.*\\.go"},
					"while ignoring": {".*/special\\.go"},
				},
			},
		},
		{
			"someMatches",
			[]*commit.File{
				{
					Filename: "app/client.go",
					Status:   commit.FileAdded,
				},
				{
					Filename: "model/user.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"app/client.go"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/.*\\.go", "server/.*\\.go"},
					"while ignoring": {".*/special\\.go"},
				},
			},
		},
		{
			"noMatches",
			[]*commit.File{
				{
					Filename: "model/order.go",
					Status:   commit.FileDeleted,
				},
				{
					Filename: "model/user.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"model/order.go", "model/user.go"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/.*\\.go", "server/.*\\.go"},
					"while ignoring": {".*/special\\.go"},
				},
			},
		},
		{
			"ignoreAll",
			[]*commit.File{
				{
					Filename: "app/special.go",
					Status:   commit.FileDeleted,
				},
				{
					Filename: "server/special.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"app/special.go", "server/special.go"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/.*\\.go", "server/.*\\.go"},
					"while ignoring": {".*/special\\.go"},
				},
			},
		},
		{
			"ignoreSome",
			[]*commit.File{
				{
					Filename: "app/normal.go",
					Status:   commit.FileDeleted,
				},
				{
					Filename: "server/special.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"app/normal.go"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/.*\\.go", "server/.*\\.go"},
					"while ignoring": {".*/special\\.go"},
				},
			},
		},
	})
}

func TestNoChangedFiles(t *testing.T) {
	p := &NoChangedFiles{
		Paths: []common.Regexp{
			common.NewCompiledRegexp(regexp.MustCompile("app/.*\\.go")),
			common.NewCompiledRegexp(regexp.MustCompile("server/.*\\.go")),
		},
		IgnorePaths: []common.Regexp{
			common.NewCompiledRegexp(regexp.MustCompile(".*/special\\.go")),
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"empty",
			[]*commit.File{},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/.*\\.go", "server/.*\\.go"},
					"while ignoring": {".*/special\\.go"},
				},
			},
		},
		{
			"onlyMatches",
			[]*commit.File{
				{
					Filename: "app/client.go",
					Status:   commit.FileAdded,
				},
				{
					Filename: "server/server.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"app/client.go"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/.*\\.go", "server/.*\\.go"},
					"while ignoring": {".*/special\\.go"},
				},
			},
		},
		{
			"someMatches",
			[]*commit.File{
				{
					Filename: "app/client.go",
					Status:   commit.FileAdded,
				},
				{
					Filename: "model/user.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"app/client.go"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/.*\\.go", "server/.*\\.go"},
					"while ignoring": {".*/special\\.go"},
				},
			},
		},
		{
			"noMatches",
			[]*commit.File{
				{
					Filename: "model/order.go",
					Status:   commit.FileDeleted,
				},
				{
					Filename: "model/user.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"model/order.go", "model/user.go"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/.*\\.go", "server/.*\\.go"},
					"while ignoring": {".*/special\\.go"},
				},
			},
		},
		{
			"ignoreAll",
			[]*commit.File{
				{
					Filename: "app/special.go",
					Status:   commit.FileDeleted,
				},
				{
					Filename: "server/special.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"app/special.go", "server/special.go"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/.*\\.go", "server/.*\\.go"},
					"while ignoring": {".*/special\\.go"},
				},
			},
		},
		{
			"ignoreSome",
			[]*commit.File{
				{
					Filename: "app/normal.go",
					Status:   commit.FileDeleted,
				},
				{
					Filename: "server/special.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"app/normal.go"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/.*\\.go", "server/.*\\.go"},
					"while ignoring": {".*/special\\.go"},
				},
			},
		},
	})
}

func TestOnlyChangedFiles(t *testing.T) {
	p := &OnlyChangedFiles{
		Paths: []common.Regexp{
			common.NewCompiledRegexp(regexp.MustCompile("app/.*\\.go")),
			common.NewCompiledRegexp(regexp.MustCompile("server/.*\\.go")),
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"empty",
			[]*commit.File{},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{},
				ConditionValues: []string{"app/.*\\.go", "server/.*\\.go"},
			},
		},
		{
			"onlyMatches",
			[]*commit.File{
				{
					Filename: "app/client.go",
					Status:   commit.FileAdded,
				},
				{
					Filename: "server/server.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied:       true,
				Values:          []string{"app/client.go", "server/server.go"},
				ConditionValues: []string{"app/.*\\.go", "server/.*\\.go"},
			},
		},
		{
			"someMatches",
			[]*commit.File{
				{
					Filename: "app/client.go",
					Status:   commit.FileAdded,
				},
				{
					Filename: "model/user.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{"model/user.go"},
				ConditionValues: []string{"app/.*\\.go", "server/.*\\.go"},
			},
		},
		{
			"noMatches",
			[]*commit.File{
				{
					Filename: "model/order.go",
					Status:   commit.FileDeleted,
				},
				{
					Filename: "model/user.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{"model/order.go"},
				ConditionValues: []string{"app/.*\\.go", "server/.*\\.go"},
			},
		},
	})
}

func TestFileNotDeleted(t *testing.T) {
	p := &FileNotDeleted{
		Paths: []common.Regexp{
			common.NewCompiledRegexp(regexp.MustCompile("workflows/.*\\.yaml")),
			common.NewCompiledRegexp(regexp.MustCompile("actions/.*\\.yaml")),
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"file not changed",
			[]*commit.File{},
			&common.PredicateResult{
				Satisfied:       true,
				Values:          []string{},
				ConditionValues: []string{"workflows/.*\\.yaml", "actions/.*\\.yaml"},
			},
		},
		{
			"files exist and modified",
			[]*commit.File{
				{
					Filename: "workflows/workflow.yaml",
					Status:   commit.FileModified,
				},
				{
					Filename: "actions/action.yaml",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied:       true,
				Values:          []string{},
				ConditionValues: []string{"workflows/.*\\.yaml", "actions/.*\\.yaml"},
			},
		},
		{
			"one file deleted",
			[]*commit.File{
				{
					Filename: "workflows/workflow.yaml",
					Status:   commit.FileDeleted,
				},
				{
					Filename: "actions/action.yaml",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{"workflows/workflow.yaml"},
				ConditionValues: []string{"workflows/.*\\.yaml", "actions/.*\\.yaml"},
			},
		},
		{
			"multiple files deleted",
			[]*commit.File{
				{
					Filename: "workflows/workflow.yaml",
					Status:   commit.FileDeleted,
				},
				{
					Filename: "actions/action.yaml",
					Status:   commit.FileDeleted,
				},
			},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{"workflows/workflow.yaml"},
				ConditionValues: []string{"workflows/.*\\.yaml", "actions/.*\\.yaml"},
			},
		},
		{
			"file deleted not matching",
			[]*commit.File{
				{
					Filename: "some/otherfile.yaml",
					Status:   commit.FileDeleted,
				},
			},
			&common.PredicateResult{
				Satisfied:       true,
				Values:          []string{"some/otherfile.yaml"},
				ConditionValues: []string{"workflows/.*\\.yaml", "actions/.*\\.yaml"},
			},
		},
	})
}

func TestFileAdded(t *testing.T) {
	p := &FileAdded{
		Paths: []common.Regexp{
			common.NewCompiledRegexp(regexp.MustCompile("app/.*\\.go")),
			common.NewCompiledRegexp(regexp.MustCompile("server/.*\\.go")),
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"no files",
			[]*commit.File{},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{},
				ConditionValues: []string{"app/.*\\.go", "server/.*\\.go"},
			},
		},
		{
			"matching file added",
			[]*commit.File{
				{
					Filename: "app/client.go",
					Status:   commit.FileAdded,
				},
				{
					Filename: "other/file.go",
					Status:   commit.FileAdded,
				},
			},
			&common.PredicateResult{
				Satisfied:       true,
				Values:          []string{"app/client.go"},
				ConditionValues: []string{"app/.*\\.go", "server/.*\\.go"},
			},
		},
		{
			"no matching files added",
			[]*commit.File{
				{
					Filename: "other/file1.go",
					Status:   commit.FileAdded,
				},
				{
					Filename: "other/file2.go",
					Status:   commit.FileAdded,
				},
			},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{"other/file1.go", "other/file2.go"},
				ConditionValues: []string{"app/.*\\.go", "server/.*\\.go"},
			},
		},
		{
			"files exist but modified",
			[]*commit.File{
				{
					Filename: "app/client.go",
					Status:   commit.FileModified,
				},
				{
					Filename: "server/server.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{},
				ConditionValues: []string{"app/.*\\.go", "server/.*\\.go"},
			},
		},
	})
}

func TestFileDeleted(t *testing.T) {
	p := &FileDeleted{
		Paths: []common.Regexp{
			common.NewCompiledRegexp(regexp.MustCompile("app/.*\\.go")),
			common.NewCompiledRegexp(regexp.MustCompile("server/.*\\.go")),
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"no files",
			[]*commit.File{},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{},
				ConditionValues: []string{"app/.*\\.go", "server/.*\\.go"},
			},
		},
		{
			"matching file deleted",
			[]*commit.File{
				{
					Filename: "app/client.go",
					Status:   commit.FileDeleted,
				},
				{
					Filename: "other/file.go",
					Status:   commit.FileDeleted,
				},
			},
			&common.PredicateResult{
				Satisfied:       true,
				Values:          []string{"app/client.go"},
				ConditionValues: []string{"app/.*\\.go", "server/.*\\.go"},
			},
		},
		{
			"no matching files deleted",
			[]*commit.File{
				{
					Filename: "other/file1.go",
					Status:   commit.FileDeleted,
				},
				{
					Filename: "other/file2.go",
					Status:   commit.FileDeleted,
				},
			},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{"other/file1.go", "other/file2.go"},
				ConditionValues: []string{"app/.*\\.go", "server/.*\\.go"},
			},
		},
		{
			"files exist but modified",
			[]*commit.File{
				{
					Filename: "app/client.go",
					Status:   commit.FileModified,
				},
				{
					Filename: "server/server.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{},
				ConditionValues: []string{"app/.*\\.go", "server/.*\\.go"},
			},
		},
	})
}

func TestFileNotAdded(t *testing.T) {
	p := &FileNotAdded{
		Paths: []common.Regexp{
			common.NewCompiledRegexp(regexp.MustCompile("workflows/.*\\.yaml")),
			common.NewCompiledRegexp(regexp.MustCompile("actions/.*\\.yaml")),
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"file not changed",
			[]*commit.File{},
			&common.PredicateResult{
				Satisfied:       true,
				Values:          []string{},
				ConditionValues: []string{"workflows/.*\\.yaml", "actions/.*\\.yaml"},
			},
		},
		{
			"files exist and modified",
			[]*commit.File{
				{
					Filename: "workflows/workflow.yaml",
					Status:   commit.FileModified,
				},
				{
					Filename: "actions/action.yaml",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied:       true,
				Values:          []string{},
				ConditionValues: []string{"workflows/.*\\.yaml", "actions/.*\\.yaml"},
			},
		},
		{
			"one file added",
			[]*commit.File{
				{
					Filename: "workflows/workflow.yaml",
					Status:   commit.FileAdded,
				},
				{
					Filename: "actions/action.yaml",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{"workflows/workflow.yaml"},
				ConditionValues: []string{"workflows/.*\\.yaml", "actions/.*\\.yaml"},
			},
		},
		{
			"multiple files added",
			[]*commit.File{
				{
					Filename: "workflows/workflow.yaml",
					Status:   commit.FileAdded,
				},
				{
					Filename: "actions/action.yaml",
					Status:   commit.FileAdded,
				},
			},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{"workflows/workflow.yaml"},
				ConditionValues: []string{"workflows/.*\\.yaml", "actions/.*\\.yaml"},
			},
		},
		{
			"file added not matching",
			[]*commit.File{
				{
					Filename: "some/otherfile.yaml",
					Status:   commit.FileAdded,
				},
			},
			&common.PredicateResult{
				Satisfied:       true,
				Values:          []string{"some/otherfile.yaml"},
				ConditionValues: []string{"workflows/.*\\.yaml", "actions/.*\\.yaml"},
			},
		},
	})
}

func TestModifiedLines(t *testing.T) {
	p := &ModifiedLines{
		Additions: ComparisonExpr{Op: OpGreaterThan, Value: 100},
		Deletions: ComparisonExpr{Op: OpGreaterThan, Value: 10},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"empty",
			[]*commit.File{},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"+0", "-0"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"added lines > 100", "deleted lines > 10"},
				},
			},
		},
		{
			"additions",
			[]*commit.File{
				{Additions: 55},
				{Additions: 10},
				{Additions: 45},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"+110"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"added lines > 100"},
				},
			},
		},
		{
			"deletions",
			[]*commit.File{
				{Additions: 5},
				{Additions: 10, Deletions: 10},
				{Additions: 5},
				{Deletions: 10},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"-20"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"deleted lines > 10"},
				},
			},
		},
	})

	p = &ModifiedLines{
		Total: ComparisonExpr{Op: OpGreaterThan, Value: 100},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"total",
			[]*commit.File{
				{Additions: 20, Deletions: 20},
				{Additions: 20},
				{Deletions: 20},
				{Additions: 20, Deletions: 20},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"total 120"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"total modifications > 100"},
				},
			},
		},
	})

	p = &ModifiedLines{
		Total: ComparisonExpr{Op: OpEquals, Value: 100},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"total",
			[]*commit.File{
				{Additions: 20, Deletions: 20},
				{Additions: 20},
				{Additions: 20, Deletions: 20},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"total 100"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"total modifications = 100"},
				},
			},
		},
	})

	p = &ModifiedLines{
		Additions: ComparisonExpr{Op: OpEquals, Value: 100},
		Deletions: ComparisonExpr{Op: OpEquals, Value: 25},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"empty",
			[]*commit.File{},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"+0", "-0"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"added lines = 100", "deleted lines = 25"},
				},
			},
		},
		{
			"additions",
			[]*commit.File{
				{Additions: 55},
				{Additions: 45},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"+100"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"added lines = 100"},
				},
			},
		},
		{
			"deletions",
			[]*commit.File{
				{Additions: 5, Deletions: 5},
				{Deletions: 10},
				{Additions: 5},
				{Deletions: 10},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"-25"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"deleted lines = 25"},
				},
			},
		},
	})

}

func TestModifiedLinesFiles(t *testing.T) {
	// Test with include patterns only
	p := &ModifiedLines{
		Additions: ComparisonExpr{Op: OpGreaterThan, Value: 50},
		Files: ModifiedLinesFileFilter{
			Include: []common.Regexp{
				common.NewCompiledRegexp(regexp.MustCompile(`.*\.go`)),
				common.NewCompiledRegexp(regexp.MustCompile(`.*\.js`)),
			},
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"filtered files meet criteria",
			[]*commit.File{
				{Filename: "app.go", Additions: 30},
				{Filename: "server.go", Additions: 25},
				{Filename: "readme.md", Additions: 100}, // Should be ignored
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"+55"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"added lines > 50"},
					"in files matching":           {`.*\.go`, `.*\.js`},
				},
			},
		},
		{
			"filtered files don't meet criteria",
			[]*commit.File{
				{Filename: "app.go", Additions: 20},
				{Filename: "server.go", Additions: 15},
				{Filename: "readme.md", Additions: 100}, // Should be ignored
			},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"+35"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"added lines > 50"},
					"in files matching":           {`.*\.go`, `.*\.js`},
				},
			},
		},
		{
			"no matching files",
			[]*commit.File{
				{Filename: "readme.md", Additions: 100},
				{Filename: "config.xml", Additions: 50},
			},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"+0"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"added lines > 50"},
					"in files matching":           {`.*\.go`, `.*\.js`},
				},
			},
		},
	})

	// Test with total modifications and include patterns
	p = &ModifiedLines{
		Total: ComparisonExpr{Op: OpGreaterThan, Value: 75},
		Files: ModifiedLinesFileFilter{
			Include: []common.Regexp{
				common.NewCompiledRegexp(regexp.MustCompile(`src/.*\.ts`)),
			},
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"typescript files with total modifications",
			[]*commit.File{
				{Filename: "src/app.ts", Additions: 30, Deletions: 20},
				{Filename: "src/utils.ts", Additions: 15, Deletions: 15},
				{Filename: "docs/readme.md", Additions: 100, Deletions: 50}, // Should be ignored
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"total 80"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"total modifications > 75"},
					"in files matching":           {`src/.*\.ts`},
				},
			},
		},
	})

	// Test with exclude patterns only
	p = &ModifiedLines{
		Additions: ComparisonExpr{Op: OpGreaterThan, Value: 50},
		Files: ModifiedLinesFileFilter{
			Exclude: []common.Regexp{
				common.NewCompiledRegexp(regexp.MustCompile(`.*\.md`)),
				common.NewCompiledRegexp(regexp.MustCompile(`.*\.txt`)),
			},
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"exclude patterns filter out unwanted files",
			[]*commit.File{
				{Filename: "app.go", Additions: 30},
				{Filename: "server.go", Additions: 25},
				{Filename: "readme.md", Additions: 100}, // Should be excluded
				{Filename: "notes.txt", Additions: 50},  // Should be excluded
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"+55"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"added lines > 50"},
					"excluding files matching":    {`.*\.md`, `.*\.txt`},
				},
			},
		},
	})

	// Test with both include and exclude patterns (exclude takes precedence)
	p = &ModifiedLines{
		Total: ComparisonExpr{Op: OpGreaterThan, Value: 50},
		Files: ModifiedLinesFileFilter{
			Include: []common.Regexp{
				common.NewCompiledRegexp(regexp.MustCompile(`src/.*`)),
			},
			Exclude: []common.Regexp{
				common.NewCompiledRegexp(regexp.MustCompile(`.*\.test\..*`)),
				common.NewCompiledRegexp(regexp.MustCompile(`.*_test\..*`)),
			},
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"include and exclude patterns work together",
			[]*commit.File{
				{Filename: "src/app.go", Additions: 30, Deletions: 10},
				{Filename: "src/app.test.go", Additions: 20, Deletions: 5},   // Should be excluded
				{Filename: "src/utils_test.go", Additions: 15, Deletions: 5}, // Should be excluded
				{Filename: "src/server.go", Additions: 25, Deletions: 5},
				{Filename: "docs/readme.md", Additions: 100, Deletions: 50}, // Not in src/, should be excluded
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"total 70"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"total modifications > 50"},
					"in files matching":           {`src/.*`},
					"excluding files matching":    {`.*\.test\..*`, `.*_test\..*`},
				},
			},
		},
		{
			"conflicting patterns - exclude takes precedence",
			[]*commit.File{
				{Filename: "src/app.test.go", Additions: 100, Deletions: 50},  // Matches include but also exclude
				{Filename: "src/utils_test.go", Additions: 75, Deletions: 25}, // Matches include but also exclude
			},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"total 0"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"total modifications > 50"},
					"in files matching":           {`src/.*`},
					"excluding files matching":    {`.*\.test\..*`, `.*_test\..*`},
				},
			},
		},
	})

	// Test with no Files config (all files should be counted)
	p = &ModifiedLines{
		Additions: ComparisonExpr{Op: OpGreaterThan, Value: 100},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"no file config counts all files",
			[]*commit.File{
				{Filename: "app.go", Additions: 30},
				{Filename: "readme.md", Additions: 50},
				{Filename: "test.txt", Additions: 25},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"+105"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"added lines > 100"},
				},
			},
		},
	})
}

func TestComparisonExpr(t *testing.T) {
	tests := map[string]struct {
		Expr   ComparisonExpr
		Value  int64
		Output bool
	}{
		"greaterThanTrue": {
			Expr:   ComparisonExpr{Op: OpGreaterThan, Value: 100},
			Value:  200,
			Output: true,
		},
		"greaterThanFalse": {
			Expr:   ComparisonExpr{Op: OpGreaterThan, Value: 100},
			Value:  50,
			Output: false,
		},
		"lessThanTrue": {
			Expr:   ComparisonExpr{Op: OpLessThan, Value: 100},
			Value:  50,
			Output: true,
		},
		"lessThanFalse": {
			Expr:   ComparisonExpr{Op: OpLessThan, Value: 100},
			Value:  200,
			Output: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ok := test.Expr.Evaluate(test.Value)
			assert.Equal(t, test.Output, ok, "evaluation was not correct")
		})
	}

	t.Run("isEmpty", func(t *testing.T) {
		assert.True(t, ComparisonExpr{}.IsEmpty(), "expression was not empty")
		assert.False(t, ComparisonExpr{Op: OpGreaterThan, Value: 100}.IsEmpty(), "expression was empty")
	})

	parseTests := map[string]struct {
		Input  string
		Output ComparisonExpr
		Err    bool
	}{
		"lessThan": {
			Input:  "<100",
			Output: ComparisonExpr{Op: OpLessThan, Value: 100},
		},
		"greaterThan": {
			Input:  ">100",
			Output: ComparisonExpr{Op: OpGreaterThan, Value: 100},
		},
		"equals": {
			Input:  "=100",
			Output: ComparisonExpr{Op: OpEquals, Value: 100},
		},
		"innerSpaces": {
			Input:  "<   35",
			Output: ComparisonExpr{Op: OpLessThan, Value: 35},
		},
		"leadingSpaces": {
			Input:  "   < 35",
			Output: ComparisonExpr{Op: OpLessThan, Value: 35},
		},
		"trailngSpaces": {
			Input:  "< 35   ",
			Output: ComparisonExpr{Op: OpLessThan, Value: 35},
		},
		"invalidOp": {
			Input: "~10",
			Err:   true,
		},
		"invalidValue": {
			Input: "< 10ab",
			Err:   true,
		},
	}

	for name, test := range parseTests {
		t.Run(name, func(t *testing.T) {
			var expr ComparisonExpr
			err := expr.UnmarshalText([]byte(test.Input))
			if test.Err {
				assert.Error(t, err, "expected error parsing expression, but got nil")
				return
			}
			if assert.NoError(t, err, "unexpected error parsing expression") {
				assert.Equal(t, test.Output, expr, "parsed expression was not correct")
			}
		})
	}
}

type FileTestCase struct {
	Name                    string
	Files                   []*commit.File
	ExpectedPredicateResult *common.PredicateResult
}

func runFileTests(t *testing.T, p Predicate, cases []FileTestCase) {
	ctx := context.Background()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			prctx := &pulltest.Context{
				ChangedFilesValue: tc.Files,
			}

			predicateResult, err := p.Evaluate(ctx, prctx)
			if assert.NoError(t, err, "evaluation failed") {
				assertPredicateResult(t, tc.ExpectedPredicateResult, predicateResult)
			}
		})
	}
}

func TestChangedFilesGlob(t *testing.T) {
	p := &ChangedFiles{
		Globs: []common.Glob{
			common.Glob("app/**/*.go"),
			common.Glob("server/*.go"),
		},
		IgnoreGlobs: []common.Glob{
			common.Glob("**/special.go"),
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"empty",
			[]*commit.File{},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/**/*.go", "server/*.go"},
					"while ignoring": {"**/special.go"},
				},
			},
		},
		{
			"recursiveMatch",
			[]*commit.File{
				{
					Filename: "app/utils/helper.go",
					Status:   commit.FileAdded,
				},
				{
					Filename: "server/api.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"app/utils/helper.go"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/**/*.go", "server/*.go"},
					"while ignoring": {"**/special.go"},
				},
			},
		},
		{
			"wildcardMatch",
			[]*commit.File{
				{
					Filename: "server/handler.go",
					Status:   commit.FileAdded,
				},
				{
					Filename: "model/user.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"server/handler.go"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/**/*.go", "server/*.go"},
					"while ignoring": {"**/special.go"},
				},
			},
		},
		{
			"ignoreGlob",
			[]*commit.File{
				{
					Filename: "app/utils/special.go",
					Status:   commit.FileAdded,
				},
				{
					Filename: "server/special.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"app/utils/special.go", "server/special.go"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/**/*.go", "server/*.go"},
					"while ignoring": {"**/special.go"},
				},
			},
		},
		{
			"noMatch",
			[]*commit.File{
				{
					Filename: "model/order.go",
					Status:   commit.FileDeleted,
				},
				{
					Filename: "docs/readme.md",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"model/order.go", "docs/readme.md"},
				ConditionsMap: map[string][]string{
					"path patterns":  {"app/**/*.go", "server/*.go"},
					"while ignoring": {"**/special.go"},
				},
			},
		},
	})
}

func TestOnlyChangedFilesGlob(t *testing.T) {
	p := &OnlyChangedFiles{
		Globs: []common.Glob{
			common.Glob("src/**"),
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"allMatch",
			[]*commit.File{
				{
					Filename: "src/main.go",
					Status:   commit.FileAdded,
				},
				{
					Filename: "src/utils/helper.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied:       true,
				Values:          []string{"src/main.go", "src/utils/helper.go"},
				ConditionValues: []string{"src/**"},
			},
		},
		{
			"someMatch",
			[]*commit.File{
				{
					Filename: "src/main.go",
					Status:   commit.FileAdded,
				},
				{
					Filename: "docs/readme.md",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{"docs/readme.md"},
				ConditionValues: []string{"src/**"},
			},
		},
	})
}

func TestFileAddedGlob(t *testing.T) {
	p := &FileAdded{
		Globs: []common.Glob{
			common.Glob("app/**/*.go"),
			common.Glob("server/*.go"),
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"matchRecursive",
			[]*commit.File{
				{
					Filename: "app/utils/helper.go",
					Status:   commit.FileAdded,
				},
			},
			&common.PredicateResult{
				Satisfied:       true,
				Values:          []string{"app/utils/helper.go"},
				ConditionValues: []string{"app/**/*.go", "server/*.go"},
			},
		},
		{
			"matchWildcard",
			[]*commit.File{
				{
					Filename: "server/api.go",
					Status:   commit.FileAdded,
				},
			},
			&common.PredicateResult{
				Satisfied:       true,
				Values:          []string{"server/api.go"},
				ConditionValues: []string{"app/**/*.go", "server/*.go"},
			},
		},
		{
			"noMatch",
			[]*commit.File{
				{
					Filename: "model/user.go",
					Status:   commit.FileAdded,
				},
			},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{"model/user.go"},
				ConditionValues: []string{"app/**/*.go", "server/*.go"},
			},
		},
		{
			"modifiedNotCounted",
			[]*commit.File{
				{
					Filename: "app/client.go",
					Status:   commit.FileModified,
				},
			},
			&common.PredicateResult{
				Satisfied:       false,
				Values:          []string{},
				ConditionValues: []string{"app/**/*.go", "server/*.go"},
			},
		},
	})
}

func TestModifiedLinesGlob(t *testing.T) {
	p := &ModifiedLines{
		Additions: ComparisonExpr{Op: OpGreaterThan, Value: 10},
		Files: ModifiedLinesFileFilter{
			IncludeGlobs: []common.Glob{
				common.Glob("**/*.go"),
			},
			ExcludeGlobs: []common.Glob{
				common.Glob("vendor/**"),
			},
		},
	}

	runFileTests(t, p, []FileTestCase{
		{
			"matchIncluded",
			[]*commit.File{
				{Filename: "app/main.go", Additions: 15},
				{Filename: "server/handler.go", Additions: 5},
			},
			&common.PredicateResult{
				Satisfied: true,
				Values:    []string{"+20"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"added lines > 10"},
					"in files matching":           {"**/*.go"},
					"excluding files matching":    {"vendor/**"},
				},
			},
		},
		{
			"excludeVendor",
			[]*commit.File{
				{Filename: "vendor/dep.go", Additions: 100},
				{Filename: "app/main.go", Additions: 5},
			},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"+5"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"added lines > 10"},
					"in files matching":           {"**/*.go"},
					"excluding files matching":    {"vendor/**"},
				},
			},
		},
		{
			"excludeNonGoFiles",
			[]*commit.File{
				{Filename: "readme.md", Additions: 100},
				{Filename: "app/main.go", Additions: 5},
			},
			&common.PredicateResult{
				Satisfied: false,
				Values:    []string{"+5"},
				ConditionsMap: map[string][]string{
					"the modification conditions": {"added lines > 10"},
					"in files matching":           {"**/*.go"},
					"excluding files matching":    {"vendor/**"},
				},
			},
		},
	})
}
