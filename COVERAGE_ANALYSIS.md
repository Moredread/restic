# Test Coverage Analysis for Restic

Generated: 2025-11-14

## Executive Summary

The restic codebase has varying test coverage across different packages. While some core utility packages have excellent coverage (>90%), there are significant gaps in backend integrations and certain command packages that could not be tested due to network issues in the current environment.

## Overall Statistics

- **Packages Successfully Tested**: 36
- **Packages with High Coverage (>80%)**: 13 packages (36%)
- **Packages with Medium Coverage (50-80%)**: 12 packages (33%)
- **Packages with Low Coverage (<50%)**: 11 packages (31%)
- **Test Failures**: 1 package (internal/fs - xattr tests)
- **Build Failures**: 14 packages (due to network issues)

## Detailed Coverage by Category

### High Coverage Packages (>80%)

These packages have excellent test coverage:

| Package | Coverage | Notes |
|---------|----------|-------|
| internal/feature | 92.6% | Feature flag management |
| internal/bloblru | 92.3% | Blob LRU cache |
| internal/repository/hashing | 91.7% | Repository hashing |
| internal/walker | 91.2% | File system walker |
| internal/backend/retry | 91.0% | Retry logic |
| internal/ui/table | 90.8% | Table UI |
| internal/textfile | 90.0% | Text file handling |
| internal/backend/mem | 89.9% | In-memory backend |
| internal/ui/restore | 89.9% | Restore UI |
| internal/options | 87.8% | Options parsing |
| internal/backend/sema | 87.2% | Semaphore implementation |
| internal/errors | 82.4% | Error handling |
| internal/ui/termstatus | 82.2% | Terminal status display |

### Medium Coverage Packages (50-80%)

These packages have reasonable coverage but could be improved:

| Package | Coverage | Recommendations |
|---------|----------|----------------|
| internal/backend/limiter | 80.0% | Add edge case tests |
| internal/repository/pack | 77.3% | Test pack file operations |
| internal/backend/local | 76.6% | Add error path tests |
| internal/ui/progress | 72.9% | Test progress edge cases |
| internal/backend/dryrun | 72.2% | Test dry-run scenarios |
| internal/backend/location | 70.7% | Add location parsing tests |
| internal/backend/layout | 68.6% | Test layout edge cases |
| internal/backend/cache | 66.2% | Add cache eviction tests |
| internal/filter | 64.0% | Add filter pattern tests |
| internal/backend | 59.7% | Add backend interface tests |
| internal/backend/rclone | 57.8% | Add rclone integration tests |
| internal/crypto | 56.8% | **IMPORTANT**: Add encryption tests |

### Low Coverage Packages (<50%)

These packages need significant testing improvements:

| Package | Coverage | Priority | Recommendations |
|---------|----------|----------|----------------|
| internal/ui | 47.1% | Medium | Add UI interaction tests |
| internal/ui/backup | 46.8% | High | Test backup UI flows |
| internal/backend/rest | 37.8% | High | Add REST API tests |
| internal/backend/util | 30.8% | Medium | Test utility functions |
| internal/selfupdate | 15.4% | Low | Test update mechanisms |
| internal/debug | 15.3% | Low | Add debug output tests |
| internal/terminal | 14.4% | Low | Test terminal interactions |
| internal/backend/sftp | 17.5% | High | **CRITICAL**: Add SFTP tests |
| internal/backend/swift | 12.6% | High | **CRITICAL**: Add Swift tests |
| internal/backend/b2 | 13.7% | High | **CRITICAL**: Add B2 tests |
| internal/backend/azure | 5.4% | **CRITICAL** | **CRITICAL**: Add Azure tests |

### Packages That Could Not Be Tested

Due to network connectivity issues in the test environment, the following packages could not be built/tested:

- **cmd/restic** - Main command package
- **internal/archiver** - Archive creation logic
- **internal/backend/all** - All backends integration
- **internal/backend/gs** - Google Cloud Storage backend
- **internal/backend/s3** - S3 backend
- **internal/checker** - Repository checker
- **internal/data** - Data management
- **internal/dump** - Dump functionality
- **internal/fuse** - FUSE filesystem
- **internal/global** - Global utilities
- **internal/migrations** - Repository migrations
- **internal/repository** - Core repository logic
- **internal/repository/index** - Repository index
- **internal/restic** - Core restic types
- **internal/restorer** - Restore functionality

**Note**: These packages need to be tested in an environment with proper network connectivity.

## Critical Findings

### 1. Cloud Backend Coverage is Very Low

The cloud backend packages (Azure, B2, Swift, GS, S3) have critically low test coverage:
- Azure: 5.4%
- B2: 13.7%
- Swift: 12.6%
- GS: Could not test
- S3: Could not test

**Recommendation**: Implement comprehensive integration tests for cloud backends. The CI/CD workflow (`.github/workflows/tests.yml`) shows that cloud backend tests are configured but require secrets.

### 2. Core Packages Could Not Be Tested

Critical packages like `cmd/restic`, `internal/repository`, and `internal/archiver` could not be tested due to network issues.

**Recommendation**: Run tests in a properly configured environment with network access or use a Go module proxy.

### 3. Test Failures in internal/fs

The `internal/fs` package has failing tests related to xattr (extended attributes):
- TestNodeRestoreAt failures for xattr handling
- TestOverwriteXattr failures

**Recommendation**: Investigate and fix xattr-related test failures. These may be environment-specific issues.

## Recommendations

### Immediate Actions

1. **Fix Network Issues**: Configure the test environment to access Go module proxy
2. **Fix xattr Tests**: Investigate and resolve the xattr test failures in internal/fs
3. **Add Cloud Backend Tests**: Prioritize adding tests for Azure (5.4% coverage)

### Short-term Improvements

1. **Improve Crypto Coverage**: The crypto package (56.8%) is critical for security and should have >90% coverage
2. **Add Backend Tests**: Focus on REST (37.8%), SFTP (17.5%), Swift (12.6%), and B2 (13.7%)
3. **Test UI Components**: Improve coverage for ui/backup (46.8%) and ui (47.1%)

### Long-term Goals

1. **Target 80% Coverage**: Aim for >80% coverage across all packages
2. **Integration Tests**: Add comprehensive integration tests for all backends
3. **CI/CD Enhancement**: Ensure all tests run in CI/CD pipeline with proper secrets

## Test Execution Notes

According to `.github/workflows/tests.yml` line 155, the standard test command is:
```bash
go test -cover ./...
```

For cloud backend tests (line 192):
```bash
go test -cover -parallel 5 -timeout 15m ./internal/backend/...
```

## How to View Coverage

To generate an HTML coverage report locally:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

To get coverage by function:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Conclusion

The restic project has good test coverage in core utility packages (>90% in many cases), but needs improvement in:
1. Cloud backend integrations (Azure, B2, Swift, GS, S3)
2. Core command and repository packages (currently untested due to network issues)
3. Cryptography package (should be >90% for security-critical code)

The test infrastructure is well-configured in CI/CD, but the current environment lacks network connectivity to fully test the codebase.
