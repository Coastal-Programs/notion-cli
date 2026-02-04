import { Args, Command, Flags } from '@oclif/core'
import { tableFlags, formatTable } from '../../utils/table-formatter'
import {
  UpdateDataSourceParameters,
  DataSourceObjectResponse,
} from '@notionhq/client/build/src/api-endpoints'
import * as notion from '../../notion'
import { outputRawJson, getDataSourceTitle } from '../../helper'
import { AutomationFlags } from '../../base-flags'
import { NotionCLIError, wrapNotionError } from '../../errors'
import { resolveNotionId } from '../../utils/notion-resolver'

export default class DbUpdate extends Command {
  static description = 'Update a data source (table) title and properties'

  static aliases: string[] = ['db:u', 'ds:update', 'ds:u']

  static examples = [
    {
      description: 'Update a data source with a specific data_source_id and title',
      command: `$ notion-cli db update DATA_SOURCE_ID -t 'My Data Source'`,
    },
    {
      description: 'Update a data source via URL',
      command: `$ notion-cli db update https://notion.so/DATABASE_ID -t 'My Data Source'`,
    },
    {
      description: 'Update a data source with a specific data_source_id and output raw json',
      command: `$ notion-cli db update DATA_SOURCE_ID -t 'My Table' -r`,
    },
  ]

  static args = {
    database_id: Args.string({
      required: true,
      description: 'Data source ID or URL (the ID of the table you want to update)',
    }),
  }

  static flags = {
    title: Flags.string({
      char: 't',
      description: 'New database title',
      required: true,
    }),
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...tableFlags,
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(DbUpdate)

    try {
      // Resolve ID from URL, direct ID, or name (future)
      const dataSourceId = await resolveNotionId(args.database_id, 'database')
      const dsTitle = flags.title

      // TODO: support other properties (description, properties schema, etc.)
      const dsProps: UpdateDataSourceParameters = {
        data_source_id: dataSourceId,
        title: [
          {
            type: 'text',
            text: {
              content: dsTitle,
            },
          },
        ],
      }

      const res = await notion.updateDataSource(dsProps)

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
        title: {
          get: (row: DataSourceObjectResponse) => {
            return getDataSourceTitle(row)
          },
        },
        object: {},
        id: {},
        url: {},
      }
      const options = {
        printLine: this.log.bind(this),
        ...flags,
      }
      formatTable([res], columns, options)
      process.exit(0)
    } catch (error) {
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            resourceType: 'database',
            attemptedId: args.database_id,
            endpoint: 'dataSources.update'
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
