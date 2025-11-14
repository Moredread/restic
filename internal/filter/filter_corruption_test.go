package filter

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadPatternsFromFilesEdgeCases tests critical edge cases that could lead to
// incorrect filtering and data loss
func TestReadPatternsFromFilesEdgeCases(t *testing.T) {
	tmpdir := t.TempDir()

	tests := []struct {
		name            string
		fileContent     string
		expectedCount   int
		expectedPattern string
		description     string
	}{
		{
			name:            "empty_file",
			fileContent:     "",
			expectedCount:   0,
			expectedPattern: "",
			description:     "Empty exclude file should result in no patterns, not error",
		},
		{
			name:            "only_comments",
			fileContent:     "# comment 1\n# comment 2\n",
			expectedCount:   0,
			expectedPattern: "",
			description:     "File with only comments should result in no patterns",
		},
		{
			name:            "only_whitespace",
			fileContent:     "   \n\t\n  \t  \n",
			expectedCount:   0,
			expectedPattern: "",
			description:     "File with only whitespace should result in no patterns",
		},
		{
			name:            "pattern_with_leading_whitespace",
			fileContent:     "   /important/data\n",
			expectedCount:   1,
			expectedPattern: "/important/data",
			description:     "Leading whitespace should be trimmed from patterns",
		},
		{
			name:            "pattern_with_trailing_whitespace",
			fileContent:     "/important/data   \n",
			expectedCount:   1,
			expectedPattern: "/important/data",
			description:     "Trailing whitespace should be trimmed from patterns",
		},
		{
			name:            "comment_after_pattern",
			fileContent:     "/data # this is a comment\n",
			expectedCount:   1,
			expectedPattern: "/data # this is a comment",
			description:     "Comments after patterns are NOT stripped (by design), full line is used",
		},
		{
			name:            "dollar_escape",
			fileContent:     "/path/with/$$dollar\n",
			expectedCount:   1,
			expectedPattern: "/path/with/$dollar",
			description:     "$$ should be replaced with single $ to allow literal dollar signs",
		},
		{
			name:            "single_dollar_not_expanded",
			fileContent:     "/path/with/$\n",
			expectedCount:   1,
			expectedPattern: "/path/with/$",
			description:     "Single $ at end should remain as $",
		},
		{
			name:            "env_var_expansion",
			fileContent:     "/path/to/$USER/data\n",
			expectedCount:   1,
			expectedPattern: "",
			description:     "Environment variables should be expanded",
		},
		{
			name:            "wildcard_pattern",
			fileContent:     "*.log\n",
			expectedCount:   1,
			expectedPattern: "*.log",
			description:     "Wildcard patterns should be preserved",
		},
		{
			name:            "double_wildcard_pattern",
			fileContent:     "**/temp/**\n",
			expectedCount:   1,
			expectedPattern: "**/temp/**",
			description:     "Double wildcard patterns should be preserved",
		},
		{
			name:            "negation_pattern",
			fileContent:     "!important.txt\n",
			expectedCount:   1,
			expectedPattern: "!important.txt",
			description:     "Negation patterns should be preserved",
		},
		{
			name:            "multiple_patterns",
			fileContent:     "/data\n/logs\n/cache\n",
			expectedCount:   3,
			expectedPattern: "/data",
			description:     "Multiple patterns should all be read",
		},
		{
			name:            "mixed_content",
			fileContent:     "# Header comment\n\n/data\n  # inline comment\n/logs  \n\n# Footer\n",
			expectedCount:   2,
			expectedPattern: "/data",
			description:     "Mixed content with comments, whitespace, and patterns",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create test file
			filename := filepath.Join(tmpdir, tc.name+".txt")
			if err := os.WriteFile(filename, []byte(tc.fileContent), 0600); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			patterns, err := readPatternsFromFiles([]string{filename})
			if err != nil {
				t.Fatalf("readPatternsFromFiles failed: %v", err)
			}

			if len(patterns) != tc.expectedCount {
				t.Errorf("%s: expected %d patterns, got %d: %v",
					tc.description, tc.expectedCount, len(patterns), patterns)
			}

			if tc.expectedCount > 0 && tc.expectedPattern != "" {
				if tc.name == "env_var_expansion" {
					// Special case: check that USER was expanded
					if patterns[0] == "/path/to/$USER/data" {
						t.Errorf("%s: environment variable was not expanded", tc.description)
					}
				} else if patterns[0] != tc.expectedPattern {
					t.Errorf("%s: expected first pattern %q, got %q",
						tc.description, tc.expectedPattern, patterns[0])
				}
			}
		})
	}
}

// TestReadPatternsFromFilesMultipleFiles tests reading from multiple exclude files
// This is critical because misconfiguration could lead to missing excludes
func TestReadPatternsFromFilesMultipleFiles(t *testing.T) {
	tmpdir := t.TempDir()

	// Create multiple exclude files
	file1 := filepath.Join(tmpdir, "exclude1.txt")
	file2 := filepath.Join(tmpdir, "exclude2.txt")
	file3 := filepath.Join(tmpdir, "exclude3.txt")

	if err := os.WriteFile(file1, []byte("/data1\n/data2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("/logs\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file3, []byte("/cache\n/temp\n"), 0600); err != nil {
		t.Fatal(err)
	}

	patterns, err := readPatternsFromFiles([]string{file1, file2, file3})
	if err != nil {
		t.Fatalf("readPatternsFromFiles failed: %v", err)
	}

	expected := []string{"/data1", "/data2", "/logs", "/cache", "/temp"}
	if len(patterns) != len(expected) {
		t.Errorf("expected %d patterns, got %d", len(expected), len(patterns))
	}

	for i, exp := range expected {
		if i >= len(patterns) {
			t.Errorf("missing pattern at index %d: %q", i, exp)
			continue
		}
		if patterns[i] != exp {
			t.Errorf("pattern at index %d: expected %q, got %q", i, exp, patterns[i])
		}
	}
}

// TestReadPatternsFromFilesMissingFile tests handling of missing exclude files
// This should return an error to alert the user
func TestReadPatternsFromFilesMissingFile(t *testing.T) {
	_, err := readPatternsFromFiles([]string{"/nonexistent/file.txt"})
	if err == nil {
		t.Error("readPatternsFromFiles should fail with missing file")
	}
}

// TestExcludePatternOptionsEmpty tests the Empty() function
// This is important for determining if any excludes are configured
func TestExcludePatternOptionsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		opts     ExcludePatternOptions
		expected bool
	}{
		{
			name:     "completely_empty",
			opts:     ExcludePatternOptions{},
			expected: true,
		},
		{
			name: "has_exclude",
			opts: ExcludePatternOptions{
				Excludes: []string{"/data"},
			},
			expected: false,
		},
		{
			name: "has_insensitive_exclude",
			opts: ExcludePatternOptions{
				InsensitiveExcludes: []string{"/data"},
			},
			expected: false,
		},
		{
			name: "has_exclude_file",
			opts: ExcludePatternOptions{
				ExcludeFiles: []string{"/path/to/file"},
			},
			expected: false,
		},
		{
			name: "has_insensitive_exclude_file",
			opts: ExcludePatternOptions{
				InsensitiveExcludeFiles: []string{"/path/to/file"},
			},
			expected: false,
		},
		{
			name: "multiple_empty_slices",
			opts: ExcludePatternOptions{
				Excludes:                []string{},
				InsensitiveExcludes:     []string{},
				ExcludeFiles:            []string{},
				InsensitiveExcludeFiles: []string{},
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.opts.Empty()
			if result != tc.expected {
				t.Errorf("Empty() = %v, want %v", result, tc.expected)
			}
		})
	}
}

// TestFilterEdgeCases tests critical edge cases in pattern matching
// that could lead to unintended file inclusion/exclusion
func TestFilterEdgeCases(t *testing.T) {
	tests := []struct {
		pattern       string
		path          string
		shouldMatch   bool
		description   string
	}{
		{
			pattern:       "",
			path:          "/any/path",
			shouldMatch:   true,
			description:   "empty pattern matches everything",
		},
		{
			pattern:       "/",
			path:          "/",
			shouldMatch:   true,
			description:   "root matches root",
		},
		{
			pattern:       "**",
			path:          "/any/deep/path",
			shouldMatch:   true,
			description:   "** matches any depth",
		},
		{
			pattern:       "/data/**",
			path:          "/data",
			shouldMatch:   true,
			description:   "/data/** should match /data itself",
		},
		{
			pattern:       "/data/**",
			path:          "/data/file.txt",
			shouldMatch:   true,
			description:   "/data/** should match files in /data",
		},
		{
			pattern:       "/data/**",
			path:          "/data/subdir/file.txt",
			shouldMatch:   true,
			description:   "/data/** should match nested files",
		},
		{
			pattern:       "/data",
			path:          "/data-backup",
			shouldMatch:   false,
			description:   "/data should not match /data-backup",
		},
		{
			pattern:       "*.log",
			path:          "/path/to/file.log",
			shouldMatch:   true,
			description:   "*.log should match .log files",
		},
		{
			pattern:       "*.log",
			path:          "/path/to/file.log.gz",
			shouldMatch:   false,
			description:   "*.log should not match .log.gz",
		},
		{
			pattern:       "/data/*.log",
			path:          "/data/subdir/file.log",
			shouldMatch:   false,
			description:   "/data/*.log should not match nested files (single wildcard is not recursive)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			matched, err := Match(tc.pattern, tc.path)
			if err != nil {
				t.Fatalf("Match failed: %v", err)
			}

			if matched != tc.shouldMatch {
				t.Errorf("%s: pattern=%q path=%q: expected match=%v, got %v",
					tc.description, tc.pattern, tc.path, tc.shouldMatch, matched)
			}
		})
	}
}

// TestFilterNegationEdgeCases tests negation pattern edge cases
// Incorrect negation handling could cause important files to be excluded
func TestFilterNegationEdgeCases(t *testing.T) {
	tests := []struct {
		pattern       string
		path          string
		isNegated     bool
		description   string
	}{
		{
			pattern:       "!important.txt",
			path:          "/important.txt",
			isNegated:     true,
			description:   "negation pattern should be detected",
		},
		{
			pattern:       "/data",
			path:          "/data",
			isNegated:     false,
			description:   "non-negated pattern",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			pattern := preparePattern(tc.pattern)
			if pattern.isNegated != tc.isNegated {
				t.Errorf("%s: expected isNegated=%v, got %v",
					tc.description, tc.isNegated, pattern.isNegated)
			}
		})
	}
}

// TestChildMatchEdgeCases tests ChildMatch edge cases
// ChildMatch determines if children of a path could match, affecting directory traversal
func TestChildMatchEdgeCases(t *testing.T) {
	tests := []struct {
		pattern       string
		path          string
		shouldMatch   bool
		description   string
	}{
		{
			pattern:       "",
			path:          "/any/path",
			shouldMatch:   true,
			description:   "empty pattern: children can match",
		},
		{
			pattern:       "relative/path",
			path:          "/absolute/path",
			shouldMatch:   true,
			description:   "relative pattern can always be nested",
		},
		{
			pattern:       "/data/specific/file.txt",
			path:          "/data",
			shouldMatch:   true,
			description:   "children of /data can match /data/specific/file.txt",
		},
		{
			pattern:       "/data/specific/file.txt",
			path:          "/other",
			shouldMatch:   false,
			description:   "children of /other cannot match /data/specific/file.txt",
		},
		{
			pattern:       "/data/**/file.txt",
			path:          "/data",
			shouldMatch:   true,
			description:   "** in pattern: children can match",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			matched, err := ChildMatch(tc.pattern, tc.path)
			if err != nil {
				t.Fatalf("ChildMatch failed: %v", err)
			}

			if matched != tc.shouldMatch {
				t.Errorf("%s: pattern=%q path=%q: expected match=%v, got %v",
					tc.description, tc.pattern, tc.path, tc.shouldMatch, matched)
			}
		})
	}
}
