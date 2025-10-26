/**
 * Update notifier utility
 * Checks for new versions of notion-cli and notifies users non-intrusively
 *
 * Runs asynchronously in background, doesn't block CLI execution
 * Caches results for 1 day to avoid unnecessary npm registry checks
 * Respects NO_UPDATE_NOTIFIER environment variable and CI environments
 */
/**
 * Check for updates and notify user if a new version is available
 *
 * This runs asynchronously and won't block CLI execution.
 * Checks are cached for 1 day by default.
 *
 * Set DEBUG=1 environment variable to see error messages if update check fails.
 *
 * @example
 * ```bash
 * # Silent mode (default)
 * notion-cli --version
 *
 * # Debug mode
 * DEBUG=1 notion-cli --version
 * ```
 */
export declare function checkForUpdates(): void;
