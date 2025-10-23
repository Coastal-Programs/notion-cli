import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../../notion'
import { BlockObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import { getBlockPlainText, outputRawJson, stripMetadata } from '../../../helper'
import { AutomationFlags } from '../../../base-flags'
import {
  NotionCLIError,
  wrapNotionError
} from '../../../errors'

export default class BlockRetrieveChildren extends Command {
  static description = 'Retrieve block children'

  static aliases: string[] = ['block:r:c']

  static examples = [
    {
      description: 'Retrieve block children',
      command: `$ notion-cli block retrieve:children BLOCK_ID`,
    },
    {
      description: 'Retrieve block children and output raw json',
      command: `$ notion-cli block retrieve:children BLOCK_ID -r`,
    },
    {
      description: 'Retrieve block children and output JSON for automation',
      command: `$ notion-cli block retrieve:children BLOCK_ID --json`,
    },
  ]

  static args = {
    block_id: Args.string({
      description: 'block_id or page_id',
      required: true,
    }),
  }

  static flags = {
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...ux.table.flags(),
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(BlockRetrieveChildren)

    try {
      // TODO: Add support start_cursor, page_size
      let res = await notion.retrieveBlockChildren(args.block_id)

      // Apply minimal flag to strip metadata
      if (flags.minimal) {
        res = stripMetadata(res)
      }

      // Handle JSON output for automation
      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          data: res,
          timestamp: new Date().toISOString()
        }, null, 2))
        process.exit(0)
        return
      }

      // Handle raw JSON output (legacy)
      if (flags.raw) {
        outputRawJson(res)
        process.exit(0)
        return
      }

      // Handle table output
      const columns = {
        object: {},
        id: {},
        type: {},
        content: {
          get: (row: BlockObjectResponse) => {
            return getBlockPlainText(row)
          },
        },
      }
      const options = {
        printLine: this.log.bind(this),
        ...flags,
      }
      ux.table(res.results, columns, options)
      process.exit(0)
    } catch (error) {
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            resourceType: 'block',
            attemptedId: args.block_id,
            endpoint: 'blocks.children.list'
          })

      if (flags.json) {
        this.log(JSON.stringify(cliError.toJSON(), null, 2))
      } else {
        this.error(cliError.toHumanString())
      }
      process.exit(1)
    }
  }
}
