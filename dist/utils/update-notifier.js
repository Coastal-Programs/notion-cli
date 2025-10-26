"use strict";
/**
 * Update notifier utility
 * Checks for new versions of notion-cli and notifies users non-intrusively
 *
 * Runs asynchronously in background, doesn't block CLI execution
 * Caches results for 1 day to avoid unnecessary npm registry checks
 * Respects NO_UPDATE_NOTIFIER environment variable and CI environments
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.checkForUpdates = void 0;
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
function checkForUpdates() {
    try {
        // Load dependencies dynamically to avoid rootDir issues
        const updateNotifier = require('update-notifier').default || require('update-notifier');
        const packageJson = require('../../package.json');
        // Initialize update notifier with package info
        const notifier = updateNotifier({
            pkg: packageJson,
            updateCheckInterval: 1000 * 60 * 60 * 24, // Check once per day
        });
        // Show notification if update is available
        // This displays a yellow-bordered box with update info
        notifier.notify({
            defer: true,
            isGlobal: true, // This is a global CLI tool
        });
    }
    catch (error) {
        // Silently fail - don't break CLI if update check fails
        // This could happen if npm registry is unreachable, network issues, etc.
        // Debug mode: Show error details for troubleshooting
        if (process.env.DEBUG) {
            console.error('Update check failed:', error instanceof Error ? error.message : error);
        }
    }
}
exports.checkForUpdates = checkForUpdates;
