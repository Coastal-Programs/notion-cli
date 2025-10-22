import { Args, Command, Flags, ux } from '@oclif/core'
import {
  CreateDatabaseParameters,
  DatabaseObjectResponse,
} from '@notionhq/client/build/src/api-endpoints'
import * as notion from '../../notion'
import { outputRawJson, getDbTitle } from '../../helper'
import { AutomationFlags } from '../../base-flags'
import { NotionCLIError, wrapNotionError } from '../../errors'
import { resolveNotionId } from '../../utils/notion-resolver'

export default class DbCreate extends Command {
  static description = 'Create a database with an initial data source (table)'

  static aliases: string[] = ['db:c']

  static examples = [
    {
      description: 'Create a database with an initial data source',
      command: `$ notion-cli db create PAGE_ID -t 'My Database'`,
    },
    {
      description: 'Create a database using page URL',
      command: `$ notion-cli db create https://notion.so/PAGE_ID -t 'My Database'`,
    },
    {
      description: 'Create a database with an initial data source and output raw json',
      command: `$ notion-cli db create PAGE_ID -t 'My Database' -r`,
    },
  ]

  static args = {
    page_id: Args.string({ required: true, description: 'Parent page ID or URL where the database will be created' }),
  }

  static flags = {
    title: Flags.string({
      char: 't',
      description: 'Title for the database (and initial data source)',
      required: true,
    }),
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...ux.table.flags(),
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(DbCreate)

    try {
      // Resolve ID from URL, direct ID, or name (future)
      const pageId = await resolveNotionId(args.page_id, 'page')
      console.log(`Creating a database in page ${pageId}`)

      const dbTitle = flags.title

      // TODO: support other properties
      const dbProps: CreateDatabaseParameters = {
        parent: {
          type: 'page_id',
          page_id: pageId,
        },
        title: [
          {
            type: 'text',
            text: {
              content: dbTitle,
            },
          },
        ],
        initial_data_source: {
          properties: {
            Name: {
              title: {},
            },
          },
        },
      }

      const res = await notion.createDb(dbProps)

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
          get: (row: DatabaseObjectResponse) => {
            return getDbTitle(row)
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
      ux.table([res], columns, options)
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
