import { Command } from '@oclif/core';
/**
 * Interactive first-time setup wizard for Notion CLI
 *
 * Guides new users through:
 * 1. Token configuration
 * 2. Connection testing
 * 3. Workspace synchronization
 *
 * Designed to provide a welcoming, educational experience that sets users up for success.
 */
export default class Init extends Command {
    static description: string;
    static examples: {
        description: string;
        command: string;
    }[];
    static flags: {
        json: import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
        'page-size': import("@oclif/core/lib/interfaces").OptionFlag<number, import("@oclif/core/lib/interfaces").CustomOptions>;
        retry: import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
        timeout: import("@oclif/core/lib/interfaces").OptionFlag<number, import("@oclif/core/lib/interfaces").CustomOptions>;
        'no-cache': import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
        verbose: import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
        minimal: import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
    };
    private isJsonMode;
    run(): Promise<void>;
    /**
     * Check if user already has a configured token
     */
    private checkExistingSetup;
    /**
     * Prompt user if they want to reconfigure
     */
    private promptReconfigure;
    /**
     * Show welcome message
     */
    private showWelcome;
    /**
     * Step 1: Setup token
     */
    private setupToken;
    /**
     * Step 2: Test connection
     */
    private testConnection;
    /**
     * Step 3: Sync workspace
     */
    private syncWorkspace;
    /**
     * Show success summary
     */
    private showSuccess;
}
