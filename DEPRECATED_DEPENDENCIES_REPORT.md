# Deprecated Dependencies Report

**Date:** 2025-10-25
**Project:** @coastal-programs/notion-cli v5.5.0
**Audit By:** backend-architect agent

---

## Executive Summary

- **Total Vulnerabilities:** 11 (9 moderate, 2 high)
- **Deprecated Direct Dependencies:** 2
- **Deprecated Transitive Dependencies:** Multiple (via oclif@3)
- **Blocking Issues:** Most vulnerabilities require oclif v4 upgrade (currently excluded)

---

## Direct Dependencies Analysis

### Production Dependencies (5) - No Issues
- `@notionhq/client@2.2.15` - PINNED (do not upgrade per constraints)
- `@oclif/core@^2` - OK (v4 excluded)
- `@oclif/plugin-help@^5` - OK (v6 excluded)
- `dayjs@^1.11.13` - OK
- `notion-to-md@^3.1.6` - OK

### Dev Dependencies - Issues Identified

#### DEPRECATED - Direct Dependencies
1. **eslint@8.57.1** → v9.38.0 available
   - Status: DEPRECATED (v9 released)
   - Action: UPGRADE to v9 with flat config
   - Priority: HIGH
   - Breaking: Yes (requires flat config migration)

2. **eslint-config-oclif-typescript@1.0.3** → v3.1.14 available
   - Status: DEPRECATED
   - Action: REMOVE (replaced by eslint-config-oclif@^5)
   - Priority: HIGH
   - Breaking: No (covered by eslint-config-oclif)

#### OUTDATED - Non-Breaking Updates Available
3. **eslint-config-prettier@8.10.2** → v10.1.8 available
   - Status: Outdated
   - Action: UPGRADE after ESLint v9
   - Priority: MEDIUM

4. **eslint-plugin-unicorn@46.0.1** → v61.0.2 available
   - Status: Outdated
   - Action: UPGRADE after ESLint v9
   - Priority: MEDIUM

5. **prettier@2.8.8** → v3.6.2 available
   - Status: Outdated
   - Action: UPGRADE
   - Priority: MEDIUM

6. **typescript@4.9.5** → v5.9.3 available
   - Status: Outdated
   - Action: Consider v5 (check for breaking changes)
   - Priority: MEDIUM

7. **shx@0.3.4** → v0.4.0 available
   - Status: Outdated
   - Action: UPGRADE
   - Priority: LOW

#### TRANSITIVE - Via oclif@3
8. **fancy-test@2.0.42** (via @oclif/test@2.5.6)
   - Status: DEPRECATED
   - Action: DEFER (only transitive, not directly used)
   - Priority: LOW
   - Note: Would require @oclif/test@4 (requires oclif v4)

---

## Security Vulnerabilities Analysis

### High Severity (2)

1. **lodash.template** - Command Injection (GHSA-35jh-r3h4-6jhm)
   - Path: oclif@3 → @oclif/plugin-warn-if-update-available → lodash.template
   - CVE: CVE-2019-10744
   - Fix: Requires oclif@4.22.32+ (BLOCKED - oclif v4 excluded)

### Moderate Severity (9)

All moderate vulnerabilities are in @octokit/* packages used by oclif@3:
- @octokit/request-error (ReDoS)
- @octokit/request (ReDoS)
- @octokit/graphql (ReDoS)
- @octokit/core (transitive)
- @octokit/plugin-paginate-rest (ReDoS)
- @octokit/rest (transitive)
- github-username (transitive)
- yeoman-generator (transitive)
- oclif@3 (root cause)

**Fix Available:** oclif@4.22.32 (BLOCKED by project constraints)

---

## Replacement Strategy

### Phase 1: ESLint v9 Migration (Tasks 3.2-3.3) - 3-4 hours
1. Install eslint@^9
2. Migrate .eslintrc.json → eslint.config.js (flat config)
3. Remove eslint-config-oclif-typescript
4. Update to eslint-config-oclif@^5
5. Update @typescript-eslint plugins to v9-compatible versions
6. Update eslint-config-prettier@^10
7. Update eslint-plugin-unicorn@^61
8. Test: `npm run lint`
9. Fix any new lint errors

### Phase 2: Update Other Dev Dependencies (Task 3.6) - 2-3 hours
1. Update prettier@^3
2. Update shx@^0.4
3. Update @types/* packages
4. Consider TypeScript@^5 (analyze breaking changes first)
5. Test after each update

### Phase 3: Security Issues (Task 3.5) - 1-2 hours
1. Run `npm audit fix` (non-breaking)
2. Document remaining vulnerabilities blocked by oclif@3
3. Verify <5 moderate vulnerabilities achievable
   - If not: Document as BLOCKED by oclif v4 constraint

### Phase 4: fancy-test (Task 3.4) - DEFERRED
- Action: Document only (transitive dependency)
- Rationale: Requires oclif v4 upgrade
- Risk: Low (only used in dev, not in published package)

---

## Risk Assessment

### Low Risk Updates
- eslint-config-prettier
- prettier
- shx
- @types/* packages

### Medium Risk Updates
- eslint@9 (flat config is breaking change)
- eslint-plugin-unicorn (major version bump)
- TypeScript@5 (major version bump)

### High Risk / Blocked Updates
- oclif@4 (EXCLUDED per constraints)
- @oclif/test@4 (requires oclif v4)
- fancy-test removal (requires @oclif/test v4)

---

## Expected Outcomes

### Achievable in This Sprint
- ESLint v9 with flat config: YES
- Remove eslint-config-oclif-typescript: YES
- <5 moderate vulnerabilities: UNLIKELY (blocked by oclif v4)
- All dev dependencies updated: YES (except oclif-related)
- Zero deprecation warnings: NO (fancy-test will remain)

### Blocked by Constraints
- Cannot fix 9/11 vulnerabilities (require oclif v4)
- Cannot remove fancy-test (transitive via @oclif/test@2)
- Best case: Document these as technical debt

---

## Recommendations

1. **Proceed with ESLint v9 migration** (high value, independent of oclif)
2. **Update non-oclif dev dependencies** (low risk)
3. **Document oclif-related vulnerabilities** as accepted technical debt
4. **Consider oclif v4 upgrade** in future sprint (separate effort)
5. **Create test/setup.ts** to fix broken test suite

---

## Success Criteria Adjustment

Original: <5 moderate vulnerabilities
**Revised:** <5 **fixable** moderate vulnerabilities (excluding oclif-blocked)

The 9 moderate vulnerabilities are all in oclif@3 transitive dependencies and cannot be fixed without upgrading to oclif v4, which is explicitly excluded per constraints.

**Recommended Goal:** Reduce fixable vulnerabilities to 0, document the 9 oclif-related vulnerabilities as technical debt.
