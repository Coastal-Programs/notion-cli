import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import {
  PageObjectResponse,
  DatabaseObjectResponse,
  DataSourceObjectResponse,
  QueryDataSourceResponse,
  QueryDataSourceParameters,
} from '@notionhq/client/build/src/api-endpoints'
import { isFullDataSource, isFullPage } from '@notionhq/client'
import * as fs from 'fs'
import * as path from 'path'
import {
  buildDatabaseQueryFilter,
  getFilterFields,
  outputRawJson,
  outputCompactJson,
  outputMarkdownTable,
  outputPrettyTable,
  getDbTitle,
  getDataSourceTitle,
  getPageTitle,
} from '../../helper'
import { client } from '../../notion'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'
import { NotionCLIError, wrapNotionError } from '../../errors'

export default class DbQuery extends Command {
  static description = 'Query a database'

  static aliases: string[] = ['db:q']

  static examples = [
    {
      description: 'Query a db with a specific database_id',
      command: `$ notion-cli db query DATABASE_ID`,
    },
    {
      description: 'Query a db with a specific database_id and raw filter string',
      command: `$ notion-cli db query -a '{"and": ...}' DATABASE_ID`,
    },
    {
      description: 'Query a db with a specific database_id and filter file',
      command: `$ notion-cli db query -f ./path/to/filter.json DATABASE_ID`,
    },
    {
      description: 'Query a db with a specific database_id and output CSV',
      command: `$ notion-cli db query --csv DATABASE_ID`,
    },
    {
      description: 'Query a db with a specific database_id and output raw json',
      command: `$ notion-cli db query --raw DATABASE_ID`,
    },
    {
      description: 'Query a db with a specific database_id and output markdown table',
      command: `$ notion-cli db query --markdown DATABASE_ID`,
    },
    {
      description: 'Query a db with a specific database_id and output compact json',
      command: `$ notion-cli db query --compact-json DATABASE_ID`,
    },
    {
      description: 'Query a db with a specific database_id and output pretty table',
      command: `$ notion-cli db query --pretty DATABASE_ID`,
    },
    {
      description: 'Query a db with a specific database_id and page size',
      command: `$ notion-cli db query -p 10 DATABASE_ID`,
    },
    {
      description: 'Query a db with a specific database_id and get all pages',
      command: `$ notion-cli db query -A DATABASE_ID`,
    },
    {
      description: 'Query a db with a specific database_id and sort property and sort direction',
      command: `$ notion-cli db query -s Name -d desc DATABASE_ID`,
    },
  ]

  static args = {
    database_id: Args.string({
      required: true,
      description: 'Database or data source ID (required for automation)',
    }),
  }

  static flags = {
    rawFilter: Flags.string({
      char: 'a',
      description: 'JSON stringified filter string',
    }),
    fileFilter: Flags.string({
      char: 'f',
      description: 'JSON filter file path',
    }),
    pageSize: Flags.integer({
      char: 'p',
      description: 'The number of results to return(1-100). ',
      min: 1,
      max: 100,
      default: 10,
    }),
    pageAll: Flags.boolean({
      char: 'A',
      description: 'get all pages',
      default: false,
    }),
    sortProperty: Flags.string({
      char: 's',
      description: 'The property to sort results by',
    }),
    sortDirection: Flags.string({
      char: 'd',
      options: ['asc', 'desc'],
      description: 'The direction to sort results',
      default: 'asc',
    }),
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
      default: false,
    }),
    ...ux.table.flags(),
    ...AutomationFlags,
    ...OutputFormatFlags,
  }

  public async run(): Promise<void> {
    const { flags, args } = await this.parse(DbQuery)

    let databaseId = args.database_id
    let queryParams: QueryDataSourceParameters

    try {
      // Build query parameters
      try {
        if (flags.rawFilter != undefined) {
          const filter = JSON.parse(flags.rawFilter)
          queryParams = {
            data_source_id: databaseId,
            filter: filter as QueryDataSourceParameters['filter'],
            page_size: flags.pageSize,
          }
        } else if (flags.fileFilter != undefined) {
          const fp = path.join('./', flags.fileFilter)
          const fj = fs.readFileSync(fp, { encoding: 'utf-8' })
          const filter = JSON.parse(fj)
          queryParams = {
            data_source_id: databaseId,
            filter: filter as QueryDataSourceParameters['filter'],
            page_size: flags.pageSize,
          }
        } else {
          let sorts: QueryDataSourceParameters['sorts'] = []
          const direction = flags.sortDirection == 'desc' ? 'descending' : 'ascending'
          if (flags.sortProperty != undefined) {
            sorts.push({
              property: flags.sortProperty,
              direction: direction,
            })
          }
          queryParams = {
            data_source_id: databaseId,
            sorts: sorts,
            page_size: flags.pageSize,
          }
        }
      } catch (e) {
        throw new NotionCLIError(
          'VALIDATION_ERROR' as any,
          `Failed to parse filter: ${e.message}`,
          { error: e }
        )
      }

      // Fetch pages from database
      let pages = []
      if (flags.pageAll) {
        pages = await notion.fetchAllPagesInDS(databaseId, queryParams.filter)
      } else {
        const res = await client.dataSources.query(queryParams)
        pages.push(...res.results)
      }

      // Define columns for table output
      const columns = {
        title: {
          get: (row: any) => {
            if (row.object == 'data_source' && isFullDataSource(row)) {
              return getDataSourceTitle(row)
            }
            if (row.object == 'page' && isFullPage(row)) {
              return getPageTitle(row)
            }
            return 'Untitled'
          },
        },
        object: {},
        id: {},
        url: {},
      }

      // Handle compact JSON output
      if (flags['compact-json']) {
        outputCompactJson(pages)
        process.exit(0)
        return
      }

      // Handle markdown table output
      if (flags.markdown) {
        outputMarkdownTable(pages, columns)
        process.exit(0)
        return
      }

      // Handle pretty table output
      if (flags.pretty) {
        outputPrettyTable(pages, columns)
        process.exit(0)
        return
      }

      // Handle JSON output for automation
      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          data: pages,
          count: pages.length,
          timestamp: new Date().toISOString()
        }, null, 2))
        process.exit(0)
        return
      }

      // Handle raw JSON output (legacy)
      if (flags.raw) {
        outputRawJson(pages)
        process.exit(0)
        return
      }

      // Handle table output (default)
      const options = {
        printLine: this.log.bind(this),
        ...flags,
      }
      ux.table(pages, columns, options)
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
