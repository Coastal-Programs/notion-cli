import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import { UpdatePageParameters, PageObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import { getPageTitle, outputRawJson } from '../../helper'
import { resolveNotionId } from '../../utils/notion-resolver'
import { AutomationFlags } from '../../base-flags'
import {
  NotionCLIError,
  wrapNotionError
} from '../../errors'
import { expandSimpleProperties } from '../../utils/property-expander'

export default class PageUpdate extends Command {
  static description = 'Update a page'

  static aliases: string[] = ['page:u']

  static examples = [
    {
      description: 'Update a page and output table',
      command: `$ notion-cli page update PAGE_ID`,
    },
    {
      description: 'Update a page via URL',
      command: `$ notion-cli page update https://notion.so/PAGE_ID -a`,
    },
    {
      description: 'Update page properties with simple format (recommended for AI agents)',
      command: `$ notion-cli page update PAGE_ID -S --properties '{"Status": "Done", "Priority": "High"}'`,
    },
    {
      description: 'Update page properties with relative date',
      command: `$ notion-cli page update PAGE_ID -S --properties '{"Due Date": "tomorrow", "Status": "In Progress"}'`,
    },
    {
      description: 'Update page with multi-select tags',
      command: `$ notion-cli page update PAGE_ID -S --properties '{"Tags": ["urgent", "bug"], "Status": "Done"}'`,
    },
    {
      description: 'Update a page and output raw json',
      command: `$ notion-cli page update PAGE_ID -r`,
    },
    {
      description: 'Update a page and archive',
      command: `$ notion-cli page update PAGE_ID -a`,
    },
    {
      description: 'Update a page and unarchive',
      command: `$ notion-cli page update PAGE_ID -u`,
    },
    {
      description: 'Update a page and archive and output raw json',
      command: `$ notion-cli page update PAGE_ID -a -r`,
    },
    {
      description: 'Update a page and unarchive and output raw json',
      command: `$ notion-cli page update PAGE_ID -u -r`,
    },
    {
      description: 'Update a page and output JSON for automation',
      command: `$ notion-cli page update PAGE_ID -a --json`,
    },
  ]

  static args = {
    page_id: Args.string({
      required: true,
      description: 'Page ID or full Notion URL (e.g., https://notion.so/...)',
    }),
  }

  static flags = {
    archived: Flags.boolean({ char: 'a', description: 'Archive the page' }),
    unarchive: Flags.boolean({ char: 'u', description: 'Unarchive the page' }),
    properties: Flags.string({
      description: 'Page properties to update as JSON string',
    }),
    'simple-properties': Flags.boolean({
      char: 'S',
      description: 'Use simplified property format (flat key-value pairs, recommended for AI agents)',
      default: false,
    }),
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...ux.table.flags(),
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(PageUpdate)

    try {
      // Resolve ID from URL, direct ID, or name (future)
      const pageId = await resolveNotionId(args.page_id, 'page')

      const pageProps: UpdatePageParameters = {
        page_id: pageId,
      }

      // Handle archived flags
      if (flags.archived) {
        pageProps.archived = true
      }
      if (flags.unarchive) {
        pageProps.archived = false
      }

      // Handle properties update
      if (flags.properties) {
        try {
          const parsedProps = JSON.parse(flags.properties)

          if (flags['simple-properties']) {
            // User provided simple format - expand to Notion format
            // Need to get the page first to find its parent database
            const page = await notion.retrievePage({ page_id: pageId })

            // Check if page is in a database
            if (!('parent' in page) || !('data_source_id' in page.parent)) {
              throw new Error(
                'The --simple-properties flag can only be used with pages in a database. ' +
                'This page does not have a parent database.'
              )
            }

            // Get the database schema
            const parentDataSourceId = page.parent.data_source_id
            const dbSchema = await notion.retrieveDataSource(parentDataSourceId)

            // Expand simple properties to Notion format
            pageProps.properties = await expandSimpleProperties(parsedProps, dbSchema.properties)
          } else {
            // Use raw Notion format
            pageProps.properties = parsedProps
          }
        } catch (error: any) {
          if (error.message.includes('Unexpected token') || error.message.includes('JSON')) {
            throw new Error(
              `Invalid JSON in --properties flag: ${error.message}\n` +
              `Example: --properties '{"Status": "Done", "Priority": "High"}'`
            )
          }
          throw error
        }
      }

      const res = await notion.updatePageProps(pageProps)

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
          get: (row: PageObjectResponse) => {
            return getPageTitle(row)
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
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            resourceType: 'page',
            attemptedId: args.page_id,
            endpoint: 'pages.update'
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
