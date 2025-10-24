import { Command } from '@oclif/core';
export default class Doctor extends Command {
    static description: string;
    static aliases: string[];
    static examples: {
        description: string;
        command: string;
    }[];
    static flags: {
        json: import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
    };
    run(): Promise<void>;
    /**
     * Check Node.js version meets requirement (>=18.0.0)
     */
    private checkNodeVersion;
    /**
     * Check if NOTION_TOKEN environment variable is set
     */
    private checkTokenSet;
    /**
     * Check if token format is valid
     * Accepts both "secret_" prefix (internal integrations) and "ntn_" prefix (OAuth tokens)
     */
    private checkTokenFormat;
    /**
     * Check network connectivity to api.notion.com
     */
    private checkNetworkConnectivity;
    /**
     * Check if can connect to Notion API (whoami check)
     */
    private checkApiConnection;
    /**
     * Check if workspace cache exists
     */
    private checkCacheExists;
    /**
     * Check if cache is fresh (< 24 hours old) or needs sync
     */
    private checkCacheFreshness;
    /**
     * Test HTTPS connection to a host
     */
    private checkHttpsConnection;
    /**
     * Print human-readable output
     */
    private printHumanReadable;
}
