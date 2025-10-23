import { Args, Command, Flags, ux } from '@oclif/core'
import { GetDataSourceResponse } from '@notionhq/client/build/src/api-endpoints'
import * as notion from '../../notion'
import {
  outputRawJson,
  outputCompactJson,
  outputMarkdownTable,
  outputPrettyTable,
  getDataSourceTitle,
  showRawFlagHint
} from '../../helper'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'
import { handleCliError } from '../../errors'
import { resolveNotionId } from '../../utils/notion-resolver'

export default class DbRetrieve extends Command {
  static description = 'Retrieve a data source (table) schema and properties'

  static aliases: string[] = ['db:r', 'ds:retrieve', 'ds:r']

  static examples = [
    {
      description: 'Retrieve a data source with full schema (recommended for AI assistants)',
      command: 'notion-cli db retrieve DATA_SOURCE_ID -r',
    },
    {
      description: 'Retrieve a data source schema via data_source_id',
      command: 'notion-cli db retrieve DATA_SOURCE_ID',
    },
    {
      description: 'Retrieve a data source via URL',
      command: 'notion-cli db retrieve https://notion.so/DATABASE_ID',
    },
    {
      description: 'Retrieve a data source and output as markdown table',
      command: 'notion-cli db retrieve DATA_SOURCE_ID --markdown',
    },
    {
      description: 'Retrieve a data source and output as compact JSON',
      command: 'notion-cli db retrieve DATA_SOURCE_ID --compact-json',
    },
  ]

  static args = {
    database_id: Args.string({
      required: true,
      description: 'Data source ID or URL (the ID of the table whose schema you want to retrieve)',
    }),
  }

  static flags = {
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json (recommended for AI assistants - returns full schema)',
    }),
    ...ux.table.flags(),
    ...AutomationFlags,
    ...OutputFormatFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(DbRetrieve)

    try {
      // Resolve ID from URL, direct ID, or name (future)
      const dataSourceId = await resolveNotionId(args.database_id, 'database')

      const res = await notion.retrieveDataSource(dataSourceId)

      // Define columns for table output
      const columns = {
        title: {
          get: (row: GetDataSourceResponse) => {
            return getDataSourceTitle(row)
          },
        },
        object: {},
        id: {},
        url: {},
      }

      // Handle compact JSON output
      if (flags['compact-json']) {
        outputCompactJson(res)
        process.exit(0)
        return
      }

      // Handle markdown table output
      if (flags.markdown) {
        outputMarkdownTable([res], columns)
        process.exit(0)
        return
      }

      // Handle pretty table output
      if (flags.pretty) {
        outputPrettyTable([res], columns)
        // Show hint after table output
        showRawFlagHint(1, res)
        process.exit(0)
        return
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

      // Handle table output (default)
      const options = {
        printLine: this.log.bind(this),
        ...flags,
      }
      ux.table([res], columns, options)

      // Show hint after table output to make -r flag discoverable
      showRawFlagHint(1, res)
      process.exit(0)
    } catch (error) {
      handleCliError(error, flags.json, {
        resourceType: 'database',
        attemptedId: args.database_id,
        endpoint: 'dataSources.retrieve'
      })
    }
  }
}
