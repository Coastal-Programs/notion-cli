import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import { UpdatePageParameters, PageObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import { getPageTitle, outputRawJson } from '../../helper'
import { resolveNotionId } from '../../utils/notion-resolver'
import { AutomationFlags } from '../../base-flags'
import { wrapNotionError } from '../../errors'

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
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...ux.table.flags(),
    ...AutomationFlags,
  }

  // NOTE: Support only archived or un archive property for now
  // TODO: Add support for updating a page properties, icon, cover
  public async run(): Promise<void> {
    const { args, flags } = await this.parse(PageUpdate)

    try {
      // Resolve ID from URL, direct ID, or name (future)
      const pageId = await resolveNotionId(args.page_id, 'page')

      const pageProps: UpdatePageParameters = {
        page_id: pageId,
      }
      if (flags.archived) {
        pageProps.archived = true
      }
      if (flags.unarchive) {
        pageProps.archived = false
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
