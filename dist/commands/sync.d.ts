import { Command } from '@oclif/core';
export default class Sync extends Command {
    static description: string;
    static aliases: string[];
    static examples: {
        description: string;
        command: string;
    }[];
    static flags: {
        json: import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
        'page-size': import("@oclif/core/lib/interfaces").OptionFlag<number, import("@oclif/core/lib/interfaces/parser").CustomOptions>;
        retry: import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
        timeout: import("@oclif/core/lib/interfaces").OptionFlag<number, import("@oclif/core/lib/interfaces/parser").CustomOptions>;
        'no-cache': import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
        force: import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
    };
    run(): Promise<void>;
    /**
     * Fetch all databases from Notion API with pagination
     */
    private fetchAllDatabases;
}
