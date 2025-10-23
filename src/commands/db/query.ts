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
  showRawFlagHint,
  stripMetadata,
} from '../../helper'
import { client } from '../../notion'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'
import {
  NotionCLIError,
  NotionCLIErrorFactory,
  wrapNotionError
} from '../../errors'
import { resolveNotionId } from '../../utils/notion-resolver'

export default class DbQuery extends Command {
  static description = 'Query a database'

  static aliases: string[] = ['db:q']

  static examples = [
    {
      description: 'Query a database with full data (recommended for AI assistants)',
      command: `$ notion-cli db query DATABASE_ID --raw`,
    },
    {
      description: 'Query all records as JSON',
      command: `$ notion-cli db query DATABASE_ID --json`,
    },
    {
      description: 'Filter with JSON object (recommended for AI agents)',
      command: `$ notion-cli db query DATABASE_ID --filter '{"property": "Status", "select": {"equals": "Done"}}' --json`,
    },
    {
      description: 'Simple text search across properties',
      command: `$ notion-cli db query DATABASE_ID --search "urgent" --json`,
    },
    {
      description: 'Load complex filter from file',
      command: `$ notion-cli db query DATABASE_ID --file-filter ./filter.json --json`,
    },
    {
      description: 'Query with AND filter',
      command: `$ notion-cli db query DATABASE_ID --filter '{"and": [{"property": "Status", "select": {"equals": "Done"}}, {"property": "Priority", "number": {"greater_than": 5}}]}' --json`,
    },
    {
      description: 'Query using database URL',
      command: `$ notion-cli db query https://notion.so/DATABASE_ID --json`,
    },
    {
      description: 'Query with sorting',
      command: `$ notion-cli db query DATABASE_ID --sort-property Name --sort-direction desc`,
    },
    {
      description: 'Query with pagination',
      command: `$ notion-cli db query DATABASE_ID --page-size 50`,
    },
    {
      description: 'Get all pages (bypass pagination)',
      command: `$ notion-cli db query DATABASE_ID --page-all`,
    },
    {
      description: 'Output as CSV',
      command: `$ notion-cli db query DATABASE_ID --csv`,
    },
    {
      description: 'Output as markdown table',
      command: `$ notion-cli db query DATABASE_ID --markdown`,
    },
    {
      description: 'Output as compact JSON',
      command: `$ notion-cli db query DATABASE_ID --compact-json`,
    },
    {
      description: 'Output as pretty table',
      command: `$ notion-cli db query DATABASE_ID --pretty`,
    },
  ]

  static args = {
    database_id: Args.string({
      required: true,
      description: 'Database or data source ID or URL (required for automation)',
    }),
  }

  static flags = {
    'page-size': Flags.integer({
      char: 'p',
      description: 'The number of results to return (1-100)',
      min: 1,
      max: 100,
      default: 10,
    }),
    'page-all': Flags.boolean({
      char: 'A',
      description: 'Get all pages (bypass pagination)',
      default: false,
    }),
    'sort-property': Flags.string({
      description: 'The property to sort results by',
    }),
    'sort-direction': Flags.string({
      options: ['asc', 'desc'],
      description: 'The direction to sort results',
      default: 'asc',
    }),
    raw: Flags.boolean({
      char: 'r',
      description: 'Output raw JSON (recommended for AI assistants - returns all page data)',
      default: false,
    }),
    ...ux.table.flags(),
    ...AutomationFlags,
    ...OutputFormatFlags,

    // New simplified filter interface (placed AFTER table flags to override)
    filter: Flags.string({
      char: 'f',
      description: 'Filter as JSON object (Notion filter API format)',
      exclusive: ['search', 'file-filter', 'rawFilter', 'fileFilter'],
    }),

    'file-filter': Flags.string({
      char: 'F',
      description: 'Load filter from JSON file',
      exclusive: ['filter', 'search', 'rawFilter', 'fileFilter'],
    }),

    search: Flags.string({
      char: 's',
      description: 'Simple text search (searches across title and common text properties)',
      exclusive: ['filter', 'file-filter', 'rawFilter', 'fileFilter'],
    }),

    // DEPRECATED: Keep for backward compatibility
    rawFilter: Flags.string({
      char: 'a',
      description: 'DEPRECATED: Use --filter instead. JSON stringified filter string',
      hidden: true,
      exclusive: ['filter', 'search', 'file-filter', 'fileFilter'],
    }),
    fileFilter: Flags.string({
      description: 'DEPRECATED: Use --file-filter instead. JSON filter file path',
      hidden: true,
      exclusive: ['filter', 'search', 'file-filter', 'rawFilter'],
    }),
  }

  public async run(): Promise<void> {
    const { flags, args } = await this.parse(DbQuery)

    try {
      // Handle deprecation warnings (output to stderr to not pollute stdout)
      if (flags.rawFilter) {
        console.error('⚠️  Warning: --rawFilter is deprecated and will be removed in v6.0.0')
        console.error('   Use --filter instead: notion-cli db query DS_ID --filter \'...\'')
        console.error('')
      }
      if (flags.fileFilter) {
        console.error('⚠️  Warning: --fileFilter is deprecated and will be removed in v6.0.0')
        console.error('   Use --file-filter instead: notion-cli db query DS_ID --file-filter ./filter.json')
        console.error('')
      }

      // Resolve ID from URL, direct ID, or name (future)
      const databaseId = await resolveNotionId(args.database_id, 'database')

      let queryParams: QueryDataSourceParameters

      // Build filter
      let filter: any = undefined

      try {
        if (flags.filter || flags.rawFilter) {
          // JSON filter object (new flag or deprecated rawFilter)
          const filterStr = flags.filter || flags.rawFilter
          try {
            filter = JSON.parse(filterStr!)
          } catch (error: any) {
            throw NotionCLIErrorFactory.invalidJson(filterStr!, error)
          }
        } else if (flags['file-filter'] || flags.fileFilter) {
          // Load from file (new flag or deprecated fileFilter)
          const filterFile = flags['file-filter'] || flags.fileFilter
          const fp = path.join('./', filterFile!)
          let fj: string
          try {
            fj = fs.readFileSync(fp, { encoding: 'utf-8' })
            filter = JSON.parse(fj)
          } catch (error: any) {
            if (error.code === 'ENOENT') {
              throw NotionCLIErrorFactory.invalidJson(
                filterFile!,
                new Error(`File not found: ${filterFile}`)
              )
            }
            throw NotionCLIErrorFactory.invalidJson(fj, error)
          }
        } else if (flags.search) {
          // Simple text search - convert to Notion filter
          // Search across common text properties using OR
          // Note: This searches properties named "Name", "Title", and "Description"
          // For more complex searches, use --filter with explicit property names
          filter = {
            or: [
              { property: 'Name', title: { contains: flags.search } },
              { property: 'Title', title: { contains: flags.search } },
              { property: 'Description', rich_text: { contains: flags.search } },
              { property: 'Name', rich_text: { contains: flags.search } },
            ]
          }
        }

        // Build sorts
        let sorts: QueryDataSourceParameters['sorts'] = []
        const direction = flags['sort-direction'] == 'desc' ? 'descending' : 'ascending'
        if (flags['sort-property']) {
          sorts.push({
            property: flags['sort-property'],
            direction: direction,
          })
        }

        // Build query parameters
        queryParams = {
          data_source_id: databaseId,
          filter: filter as QueryDataSourceParameters['filter'],
          sorts: sorts.length > 0 ? sorts : undefined,
          page_size: flags['page-size'],
        }
      } catch (e: any) {
        // Re-throw NotionCLIError, wrap others
        if (e instanceof NotionCLIError) {
          throw e
        }
        throw wrapNotionError(e, {
          resourceType: 'database',
          userInput: args.database_id
        })
      }

      // Fetch pages from database
      let pages = []
      if (flags['page-all']) {
        pages = await notion.fetchAllPagesInDS(databaseId, queryParams.filter)
      } else {
        const res = await client.dataSources.query(queryParams)
        pages.push(...res.results)
      }

      // Apply minimal flag to strip metadata
      if (flags.minimal) {
        pages = stripMetadata(pages)
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
        // Show hint after table output (use first page as sample)
        if (pages.length > 0) {
          showRawFlagHint(pages.length, pages[0])
        }
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

      // Show hint after table output to make -r flag discoverable
      // Use first page as sample to count fields
      if (pages.length > 0) {
        showRawFlagHint(pages.length, pages[0])
      }
      process.exit(0)
    } catch (error) {
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            resourceType: 'database',
            attemptedId: args.database_id,
            endpoint: 'dataSources.query'
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
