import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../../notion'
import { BlockObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import { getBlockPlainText, outputRawJson, enrichChildDatabaseBlock, getChildDatabasesWithIds } from '../../../helper'
import { AutomationFlags } from '../../../base-flags'
import { wrapNotionError } from '../../../errors'

export default class BlockRetrieveChildren extends Command {
  static description = 'Retrieve block children (use --show-databases to discover queryable databases)'

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
      description: 'List child databases with queryable IDs',
      command: `$ notion-cli block retrieve:children PAGE_ID --show-databases`,
    },
    {
      description: 'Get child databases as JSON (for piping to db query)',
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
      description: 'list child_database blocks with their queryable data_source_id',
      default: false,
    }),
    ...ux.table.flags(),
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(BlockRetrieveChildren)

    try {
      // TODO: Add support start_cursor, page_size
      const res = await notion.retrieveBlockChildren(args.block_id)

      // Handle --show-databases flag
      if (flags['show-databases']) {
        const databases = await getChildDatabasesWithIds(res.results as BlockObjectResponse[])

        // Handle JSON output for automation
        if (flags.json) {
          this.log(JSON.stringify({
            success: true,
            data: databases,
            count: databases.length,
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

        // Handle table output
        const columns = {
          block_id: {
            header: 'Block ID',
          },
          title: {
            header: 'Database Title',
          },
          data_source_id: {
            header: 'Data Source ID (for querying)',
          },
          database_id: {
            header: 'Database ID',
          },
        }
        const options = {
          printLine: this.log.bind(this),
          ...flags,
        }
        ux.table(databases, columns, options)

        // Show helpful hint if databases were found
        if (databases.length > 0) {
          const queryableCount = databases.filter(db => db.data_source_id).length
          if (queryableCount > 0) {
            this.log(`\nTip: Use "notion-cli db query <data_source_id>" to query these databases`)
          }
          if (queryableCount < databases.length) {
            this.log(`\nNote: ${databases.length - queryableCount} database(s) could not be resolved to a queryable ID`)
          }
        } else {
          this.log('No child_database blocks found on this page')
        }

        process.exit(0)
        return
      }

      // Enrich child_database blocks with data_source_id when in raw/json mode
      if (flags.json || flags.raw) {
        // Enrich all child_database blocks in the results
        const enrichPromises = res.results
          .filter((block: any) => block.type === 'child_database')
          .map((block: any) => enrichChildDatabaseBlock(block))

        await Promise.all(enrichPromises)
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
      const cliError = wrapNotionError(error)
      if (flags.json) {
        this.log(JSON.stringify(cliError.toJSON(), null, 2))
      } else {
        this.error(cliError.message)
      }
      process.exit(1)
    }
  }
}
