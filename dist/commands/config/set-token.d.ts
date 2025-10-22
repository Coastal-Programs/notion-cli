import { Command } from '@oclif/core';
export default class ConfigSetToken extends Command {
    static description: string;
    static aliases: string[];
    static examples: {
        description: string;
        command: string;
    }[];
    static args: {
        token: import("@oclif/core/lib/interfaces/parser").Arg<string, Record<string, unknown>>;
    };
    static flags: {
        json: import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
        'page-size': import("@oclif/core/lib/interfaces").OptionFlag<number, import("@oclif/core/lib/interfaces/parser").CustomOptions>;
        retry: import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
        timeout: import("@oclif/core/lib/interfaces").OptionFlag<number, import("@oclif/core/lib/interfaces/parser").CustomOptions>;
        'no-cache': import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
    };
    run(): Promise<void>;
    /**
     * Detect the current shell
     */
    private detectShell;
    /**
     * Get the rc file path for the detected shell
     */
    private getRcFilePath;
}
