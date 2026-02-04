import { Args, Command, Flags } from '@oclif/core'
import * as notion from '../../../notion'
import { BlockObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import { getBlockPlainText, outputRawJson, stripMetadata, enrichChildDatabaseBlock, getChildDatabasesWithIds } from '../../../helper'
import { AutomationFlags } from '../../../base-flags'
import {
  NotionCLIError,
  wrapNotionError
} from '../../../errors'
import { tableFlags, formatTable } from '../../../utils/table-formatter'

export default class BlockRetrieveChildren extends Command {
  static description = 'Retrieve block children (supports database discovery via --show-databases)'

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
    {
      description: 'Discover databases on a page with queryable IDs',
      command: `$ notion-cli block retrieve:children PAGE_ID --show-databases`,
    },
    {
      description: 'Get databases as JSON for automation',
      command: `$ notion-cli block retrieve:children PAGE_ID --show-databases --json`,
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
    'show-databases': Flags.boolean({
      char: 'd',
      description: 'show only child databases with their queryable IDs (data_source_id)',
      default: false,
    }),
    ...tableFlags,
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(BlockRetrieveChildren)

    try {
      // TODO: Add support start_cursor, page_size
      let res = await notion.retrieveBlockChildren(args.block_id)

      // Handle --show-databases flag: filter and enrich child_database blocks
      if (flags['show-databases']) {
        const databases = await getChildDatabasesWithIds(res.results as BlockObjectResponse[])

        // Handle JSON output for automation
        if (flags.json) {
          this.log(JSON.stringify({
            success: true,
            data: databases,
            timestamp: new Date().toISOString()
          }, null, 2))
          process.exit(0)
          return
        }

        // Handle raw JSON output
        if (flags.raw) {
          outputRawJson(databases)
          process.exit(0)
          return
        }

        // Display databases in table format
        const columns = {
          block_id: {
            header: 'Block ID',
          },
          title: {
            header: 'Title',
          },
          data_source_id: {
            header: 'Data Source ID',
          },
          database_id: {
            header: 'Database ID',
          },
        }

        const options = {
          printLine: this.log.bind(this),
          ...flags,
        }

        formatTable(databases, columns, options)

        // Show helpful tip
        if (databases.length > 0) {
          this.log('\nTip: Use the data_source_id to query databases:')
          this.log(`  notion-cli db query <data_source_id>`)
        } else {
          this.log('\nNo child databases found on this page.')
        }

        process.exit(0)
        return
      }

      // Auto-enrich child_database blocks for JSON/raw output
      if (flags.json || flags.raw) {
        const enrichedResults = await Promise.all(
          (res.results as BlockObjectResponse[]).map(async (block) => {
            if (block.type === 'child_database') {
              return await enrichChildDatabaseBlock(block)
            }
            return block
          })
        )
        res.results = enrichedResults
      }

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
      formatTable(res.results, columns, options)
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
