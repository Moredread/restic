# Data Corruption Prevention - Test Improvements

Generated: 2025-11-14

## Summary

Added comprehensive test coverage for critical data integrity scenarios in packages that are **not backend-dependent** but could lead to data corruption if bugs exist. Focus areas: cryptography and file filtering.

## Coverage Improvements

### Internal/Crypto Package
- **Before**: 56.8% coverage
- **After**: 61.9% coverage
- **Improvement**: +5.1 percentage points
- **File**: `internal/crypto/crypto_corruption_test.go`

### Internal/Filter Package
- **Before**: 64.0% coverage
- **After**: 74.2% coverage
- **Improvement**: +10.2 percentage points
- **File**: `internal/filter/filter_corruption_test.go`

## Critical Scenarios Tested

### Cryptography (`internal/crypto`)

The crypto package implements AES-256 + Poly1305 authenticated encryption. Bugs here could lead to:
- **Silent data corruption** (failed authentication not detected)
- **Data loss** (decryption with wrong keys)
- **Security vulnerabilities** (nonce reuse, weak keys)

#### New Tests Added:

1. **TestZeroLengthPlaintext**
   - **Risk**: Edge case handling of empty data
   - **Validates**: Encryption/decryption of zero-length data works correctly
   - **Prevents**: Buffer handling errors with empty files

2. **TestInvalidKeySeal** & **TestInvalidKeyOpen**
   - **Risk**: Using uninitialized keys could corrupt data silently
   - **Validates**: Invalid keys are rejected (panic for Seal, error for Open)
   - **Prevents**: Silent corruption from zero/invalid keys

3. **TestZeroNonceOpen**
   - **Risk**: Zero nonces are a critical security vulnerability in CTR mode
   - **Validates**: Zero nonces are rejected
   - **Prevents**: Cryptographic weakness from nonce reuse

4. **TestCiphertextTooShort**
   - **Risk**: Truncated data could indicate storage corruption
   - **Validates**: Ciphertext shorter than MAC size is rejected
   - **Prevents**: Reading corrupted/incomplete data
   - **Tests**: Empty, 1 byte, half MAC, one byte short

5. **TestCiphertextCorruptionAllBytes**
   - **Risk**: Corruption in ANY byte must be detected
   - **Validates**: MAC verification detects single-bit flips in any byte
   - **Prevents**: Undetected data corruption
   - **Coverage**: Tests corruption of all 256+ bytes in ciphertext

6. **TestMACCorruptionAllBytes**
   - **Risk**: MAC integrity is critical for authentication
   - **Validates**: Corruption of any bit in the 16-byte MAC is detected
   - **Prevents**: Authentication bypass
   - **Coverage**: Tests all 128 bits (16 bytes × 8 bits)

7. **TestNonceCorruption**
   - **Risk**: Wrong nonce produces garbage plaintext
   - **Validates**: Nonce mismatch is detected via MAC verification
   - **Prevents**: Decryption with wrong nonce

8. **TestKeyCorruption**
   - **Risk**: Wrong key produces garbage plaintext
   - **Validates**: Key mismatch is detected via MAC verification
   - **Prevents**: Decryption with wrong key

9. **TestCiphertextTruncation**
   - **Risk**: Incomplete writes during storage
   - **Validates**: Truncation is detected at various lengths
   - **Prevents**: Accepting partially written data
   - **Tests**: Full length - 1, full - 5, half, just past MAC, exactly MAC

10. **TestCiphertextExtension**
    - **Risk**: Extra data could indicate corruption or format errors
    - **Validates**: Extra bytes cause MAC verification failure
    - **Prevents**: Accepting malformed ciphertext

11. **TestPartiallyInvalidKey**
    - **Risk**: Partially initialized keys
    - **Validates**: Both EncryptionKey and MACKey must be non-zero
    - **Prevents**: Using partially initialized cryptographic material

12. **TestNonceReuse**
    - **Risk**: CTR mode with reused nonces leaks plaintext
    - **Validates**: Documents the nonce reuse vulnerability
    - **Prevents**: Understanding of why unique nonces are critical
    - **Note**: This test demonstrates the vulnerability, not a fix

### File Filtering (`internal/filter`)

The filter package determines which files are backed up or excluded. Bugs here could lead to:
- **Data loss** (excluding critical files unintentionally)
- **Privacy breach** (including sensitive files that should be excluded)
- **Incomplete backups** (wrong pattern matching)

#### New Tests Added:

1. **TestReadPatternsFromFilesEdgeCases**
   - **Risk**: Malformed exclude files could cause wrong filtering
   - **Validates**: 14 edge cases in pattern file parsing
   - **Prevents**: Silent failures in exclude pattern handling

   **Specific cases tested:**
   - Empty files → no patterns (not error)
   - Only comments → no patterns
   - Only whitespace → no patterns
   - Leading/trailing whitespace → trimmed correctly
   - `$$` escape → replaced with `$`
   - Environment variable expansion → works correctly
   - Wildcard patterns → preserved
   - Double wildcard (`**`) → preserved
   - Negation patterns (`!`) → preserved
   - Multiple patterns → all read
   - Mixed content → parsed correctly

2. **TestReadPatternsFromFilesMultipleFiles**
   - **Risk**: Multiple exclude files not combined correctly
   - **Validates**: Patterns from all files are concatenated
   - **Prevents**: Missing excludes from secondary files
   - **Tests**: 3 files with 5 total patterns

3. **TestReadPatternsFromFilesMissingFile**
   - **Risk**: Missing exclude file could go unnoticed
   - **Validates**: Error returned for missing files
   - **Prevents**: Silent failure when exclude file doesn't exist

4. **TestExcludePatternOptionsEmpty**
   - **Risk**: Incorrectly detecting if excludes are configured
   - **Validates**: Empty() returns correct value for all combinations
   - **Prevents**: Logic errors in exclude handling
   - **Tests**: 6 combinations of empty/non-empty fields

5. **TestFilterEdgeCases**
   - **Risk**: Wrong pattern matching leads to wrong files backed up
   - **Validates**: 10 critical pattern matching scenarios
   - **Prevents**: Unexpected file inclusion/exclusion

   **Specific cases tested:**
   - Empty pattern matches everything
   - Root path matching
   - `**` wildcard depth
   - `/data/**` matching `/data` itself
   - Prefix matching (`/data` vs `/data-backup`)
   - Wildcard extensions (`*.log` vs `*.log.gz`)
   - Non-recursive single wildcard behavior

6. **TestFilterNegationEdgeCases**
   - **Risk**: Negation (`!pattern`) not detected correctly
   - **Validates**: Negation flag is set correctly
   - **Prevents**: Negation patterns being treated as normal patterns

7. **TestChildMatchEdgeCases**
   - **Risk**: Directory traversal decisions could skip important files
   - **Validates**: ChildMatch determines if subdirectories should be traversed
   - **Prevents**: Premature directory exclusion
   - **Tests**: 5 scenarios including relative patterns, absolute patterns, and `**`

## Potential Data Corruption Issues Prevented

### High Severity

1. **Silent Data Corruption in Crypto**
   - **Tests**: TestCiphertextCorruptionAllBytes, TestMACCorruptionAllBytes
   - **Impact**: Any single-bit corruption is now verified to be detected
   - **Without tests**: Corruption could go unnoticed, leading to data loss

2. **Invalid Key Usage**
   - **Tests**: TestInvalidKeySeal, TestInvalidKeyOpen, TestPartiallyInvalidKey
   - **Impact**: Prevents encryption/decryption with uninitialized keys
   - **Without tests**: Could produce garbage output or crash

3. **Truncated Ciphertext**
   - **Tests**: TestCiphertextTooShort, TestCiphertextTruncation
   - **Impact**: Detects incomplete writes to storage
   - **Without tests**: Could attempt to decrypt partial data

4. **Wrong File Exclusion**
   - **Tests**: TestFilterEdgeCases, TestReadPatternsFromFilesEdgeCases
   - **Impact**: Ensures exclude patterns work as expected
   - **Without tests**: Critical files might not be backed up

### Medium Severity

5. **Nonce/Key Mismatches**
   - **Tests**: TestNonceCorruption, TestKeyCorruption, TestZeroNonceOpen
   - **Impact**: Ensures wrong keys/nonces are detected
   - **Without tests**: Could produce garbage plaintext silently

6. **Pattern File Parsing**
   - **Tests**: TestReadPatternsFromFilesEdgeCases (14 cases)
   - **Impact**: Ensures exclude files are parsed correctly
   - **Without tests**: Malformed files could cause wrong behavior

7. **Empty Data Handling**
   - **Tests**: TestZeroLengthPlaintext
   - **Impact**: Ensures empty files are handled correctly
   - **Without tests**: Edge case might cause errors

### Low Severity

8. **Directory Traversal**
   - **Tests**: TestChildMatchEdgeCases
   - **Impact**: Ensures correct directory traversal decisions
   - **Without tests**: Might traverse unnecessary directories or skip important ones

9. **Negation Pattern Handling**
   - **Tests**: TestFilterNegationEdgeCases
   - **Impact**: Ensures `!pattern` is detected
   - **Without tests**: Negation patterns might not work

## Test Quality Metrics

### Crypto Tests
- **Total new tests**: 12 comprehensive test functions
- **Line coverage improvement**: +5.1%
- **Scenarios covered**:
  - Edge cases: 4
  - Corruption detection: 5
  - Invalid input handling: 3
- **All tests**: ✅ PASSING

### Filter Tests
- **Total new tests**: 7 comprehensive test functions
- **Sub-tests**: 37 individual scenarios
- **Line coverage improvement**: +10.2%
- **Scenarios covered**:
  - Pattern file parsing: 14 edge cases
  - Pattern matching: 10 scenarios
  - Negation: 2 scenarios
  - ChildMatch: 5 scenarios
  - Multi-file handling: 1 scenario
  - Error handling: 1 scenario
  - Empty detection: 6 scenarios
- **All tests**: ✅ PASSING

## Verification

All tests have been run and pass successfully:

```bash
$ go test -v ./internal/crypto -run "Corruption|InvalidKey|ZeroNonce|ZeroLength"
PASS
ok  	github.com/restic/restic/internal/crypto	0.010s

$ go test ./internal/filter -run "Corruption|EdgeCases|Missing"
PASS
ok  	github.com/restic/restic/internal/filter	0.025s
```

Coverage verification:
```bash
$ go test -cover ./internal/crypto ./internal/filter
ok  	github.com/restic/restic/internal/crypto	coverage: 61.9% of statements
ok  	github.com/restic/restic/internal/filter	coverage: 74.2% of statements
```

## Recommendations

### Immediate Actions

1. **Run these tests in CI/CD**: Ensure all new tests run on every commit
2. **Monitor coverage**: Track that coverage doesn't decrease
3. **Review crypto code**: While tests pass, consider security audit of crypto implementation

### Short-term Improvements

1. **Add benchmark tests**: Verify performance impact of edge case handling
2. **Add fuzz testing**: Use Go's fuzzing for crypto and filter packages
3. **Test with real corruption**: Inject actual filesystem corruption to verify detection

### Long-term Goals

1. **Increase crypto coverage to >90%**: Security-critical code should have maximum coverage
2. **Add integration tests**: Test crypto + storage + retrieval end-to-end
3. **Add corruption simulation**: Systematically test all corruption scenarios

## Files Modified

1. `/home/user/restic/internal/crypto/crypto_corruption_test.go` - NEW
2. `/home/user/restic/internal/filter/filter_corruption_test.go` - NEW
3. `.gitignore` - Updated to exclude test artifacts

## Related Documentation

- Original coverage analysis: `COVERAGE_ANALYSIS.md`
- Coverage summary: `coverage_summary.txt`
- Visual summary: `coverage_visual_summary.txt`

## Conclusion

These tests significantly improve the reliability of restic's critical data integrity paths:

1. **Cryptography**: Now have comprehensive tests for corruption detection, ensuring the MAC properly protects data integrity
2. **File Filtering**: Now have comprehensive tests for pattern matching edge cases, ensuring backups include/exclude the right files

Both areas are **critical for preventing data loss** and these tests provide strong guarantees that the code handles edge cases correctly.

The tests are designed to catch:
- Implementation bugs
- Regression from future changes
- Edge cases that might not be obvious
- Data corruption scenarios

**All tests pass**, indicating the current implementation correctly handles these scenarios.
