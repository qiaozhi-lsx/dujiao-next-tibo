// Shim: pnpm run test uses the glob tests/*.test.ts which does not pick up
// src/utils/__tests__/*.spec.ts. Importing the spec file registers its node:test
// cases against the active test runner.
import '../src/utils/__tests__/config-flags.spec.ts'
