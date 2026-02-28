import { Args, Command, Flags } from "@oclif/core";
import { tableFlags, formatTable } from "../../utils/table-formatter";
import * as notion from "../../notion";
import { BlockObjectResponse } from "@notionhq/client/build/src/api-endpoints";
import { getBlockPlainText, outputRawJson, stripMetadata } from "../../helper";
import { AutomationFlags } from "../../base-flags";
import { NotionCLIError, wrapNotionError } from "../../errors";
import { resolveNotionId } from "../../utils/notion-resolver";

export default class BlockRetrieve extends Command {
  static description = "Retrieve a block";

  static aliases: string[] = ["block:r"];

  static examples = [
    {
      description: "Retrieve a block",
      command: `$ notion-cli block retrieve BLOCK_ID`,
    },
    {
      description: "Retrieve a block and output raw json",
      command: `$ notion-cli block retrieve BLOCK_ID -r`,
    },
    {
      description: "Retrieve a block and output JSON for automation",
      command: `$ notion-cli block retrieve BLOCK_ID --json`,
    },
  ];

  static args = {
    block_id: Args.string({ required: true, description: "Block ID or URL" }),
  };

  static flags = {
    raw: Flags.boolean({
      char: "r",
      description: "output raw json",
    }),
    ...tableFlags,
    ...AutomationFlags,
  };

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(BlockRetrieve);

    try {
      const blockId = await resolveNotionId(args.block_id, "page");
      let res = await notion.retrieveBlock(blockId);

      // Apply minimal flag to strip metadata
      if (flags.minimal) {
        res = stripMetadata(res);
      }

      // Handle JSON output for automation
      if (flags.json) {
        this.log(
          JSON.stringify(
            {
              success: true,
              data: res,
              timestamp: new Date().toISOString(),
            },
            null,
            2,
          ),
        );
        process.exit(0);
        return;
      }

      // Handle raw JSON output (legacy)
      if (flags.raw) {
        outputRawJson(res);
        process.exit(0);
        return;
      }

      // Handle table output
      const columns = {
        object: {},
        id: {},
        type: {},
        parent: {},
        content: {
          get: (row: BlockObjectResponse) => {
            return getBlockPlainText(row);
          },
        },
      };
      const options = {
        printLine: this.log.bind(this),
        ...flags,
      };
      formatTable([res], columns, options);
      process.exit(0);
    } catch (error) {
      const cliError =
        error instanceof NotionCLIError
          ? error
          : wrapNotionError(error, {
              resourceType: "block",
              attemptedId: args.block_id,
              endpoint: "blocks.retrieve",
            });

      if (flags.json) {
        this.log(JSON.stringify(cliError.toJSON(), null, 2));
      } else {
        this.error(cliError.toHumanString());
      }
      process.exit(1);
    }
  }
}
