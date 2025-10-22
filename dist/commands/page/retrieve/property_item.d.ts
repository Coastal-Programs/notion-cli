import { Command } from '@oclif/core';
export default class PageRetrievePropertyItem extends Command {
    static description: string;
    static aliases: string[];
    static examples: {
        description: string;
        command: string;
    }[];
    static args: {
        page_id: import("@oclif/core/lib/interfaces/parser").Arg<string, Record<string, unknown>>;
        property_id: import("@oclif/core/lib/interfaces/parser").Arg<string, Record<string, unknown>>;
    };
    run(): Promise<void>;
}
