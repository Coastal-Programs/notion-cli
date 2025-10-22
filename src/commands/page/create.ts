import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import * as fs from 'fs'
import * as path from 'path'
import { markdownToBlocks } from '@tryfabric/martian'
import {
  CreatePageParameters,
  PageObjectResponse,
  BlockObjectRequest,
} from '@notionhq/client/build/src/api-endpoints'
import { getPageTitle, outputRawJson } from '../../helper'
import { resolveNotionId } from '../../utils/notion-resolver'
import { AutomationFlags } from '../../base-flags'
import { wrapNotionError } from '../../errors'

export default class PageCreate extends Command {
  static description = 'Create a page'

  static aliases: string[] = ['page:c']

  static examples = [
    {
      description: 'Create a page via interactive mode',
      command: `$ notion-cli page create`,
    },
    {
      description: 'Create a page with a specific parent_page_id',
      command: `$ notion-cli page create -p PARENT_PAGE_ID`,
    },
    {
      description: 'Create a page with a parent page URL',
      command: `$ notion-cli page create -p https://notion.so/PARENT_PAGE_ID`,
    },
    {
      description: 'Create a page with a specific parent_db_id',
      command: `$ notion-cli page create -d PARENT_DB_ID`,
    },
    {
      description: 'Create a page with a specific source markdown file and parent_page_id',
      command: `$ notion-cli page create -f ./path/to/source.md -p PARENT_PAGE_ID`,
    },
    {
      description: 'Create a page with a specific source markdown file and parent_db_id',
      command: `$ notion-cli page create -f ./path/to/source.md -d PARENT_DB_ID`,
    },
    {
      description:
        'Create a page with a specific source markdown file and output raw json with parent_page_id',
      command: `$ notion-cli page create -f ./path/to/source.md -p PARENT_PAGE_ID -r`,
    },
    {
      description: 'Create a page and output JSON for automation',
      command: `$ notion-cli page create -p PARENT_PAGE_ID --json`,
    },
  ]

  static flags = {
    parent_page_id: Flags.string({
      char: 'p',
      description: 'Parent page ID or URL (to create a sub-page)',
    }),
    parent_data_source_id: Flags.string({
      char: 'd',
      description: 'Parent data source ID or URL (to create a page in a table)',
    }),
    file_path: Flags.string({
      char: 'f',
      description: 'Path to a source markdown file',
    }),
    title_property: Flags.string({
      char: 't',
      description: 'Name of the title property (defaults to "Name" if not specified)',
      default: 'Name',
    }),
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...ux.table.flags(),
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(PageCreate)

    try {
      let pageProps: CreatePageParameters
      let pageParent: CreatePageParameters['parent']

      if (flags.parent_page_id) {
        // Resolve parent page ID from URL, direct ID, or name (future)
        const parentPageId = await resolveNotionId(flags.parent_page_id, 'page')
        pageParent = {
          page_id: parentPageId,
        }
      } else {
        // Resolve parent database ID from URL, direct ID, or name (future)
        const parentDataSourceId = await resolveNotionId(flags.parent_data_source_id!, 'database')
        pageParent = {
          data_source_id: parentDataSourceId,
        }
      }

      if (flags.file_path) {
        const p = path.join('./', flags.file_path)
        const fileName = path.basename(flags.file_path)
        const md = fs.readFileSync(p, { encoding: 'utf-8' })
        const blocks = markdownToBlocks(md)

        // TODO: Add support for creating a page from a template
        pageProps = {
          parent: pageParent,
          properties: {
            [flags.title_property]: {
              title: [{ text: { content: fileName } }],
            },
          },
          children: blocks as BlockObjectRequest[],
        }
      } else {
        pageProps = {
          parent: pageParent,
          properties: {},
        }
      }

      const res = await notion.createPage(pageProps)

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
