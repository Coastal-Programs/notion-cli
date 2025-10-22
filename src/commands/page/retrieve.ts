import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import { GetPageParameters, PageObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import {
  getPageTitle,
  outputRawJson,
  outputCompactJson,
  outputMarkdownTable,
  outputPrettyTable
} from '../../helper'
import { NotionToMarkdown } from 'notion-to-md'
import { OutputFormatFlags } from '../../base-flags'

export default class PageRetrieve extends Command {
  static description = 'Retrieve a page'

  static aliases: string[] = ['page:r']

  static examples = [
    {
      description: 'Retrieve a page and output table',
      command: `$ notion-cli page retrieve PAGE_ID`,
    },
    {
      description: 'Retrieve a page and output raw json',
      command: `$ notion-cli page retrieve PAGE_ID -r`,
    },
    {
      description: 'Retrieve a page and output markdown',
      command: `$ notion-cli page retrieve PAGE_ID -m`,
    },
    {
      description: 'Retrieve a page metadata and output as markdown table',
      command: `$ notion-cli page retrieve PAGE_ID --markdown`,
    },
    {
      description: 'Retrieve a page metadata and output as compact JSON',
      command: `$ notion-cli page retrieve PAGE_ID --compact-json`,
    },
  ]

  static args = {
    page_id: Args.string({ required: true }),
  }

  static flags = {
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    markdown: Flags.boolean({
      char: 'm',
      description: 'output page content as markdown',
    }),
    ...ux.table.flags(),
    ...OutputFormatFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(PageRetrieve)

    // Handle page content as markdown (uses NotionToMarkdown)
    if (flags.markdown) {
      const n2m = new NotionToMarkdown({ notionClient: notion.client })
      const mdBlocks = await n2m.pageToMarkdown(args.page_id)
      const mdString = n2m.toMarkdownString(mdBlocks)
      console.log(mdString.parent)
      this.exit(0)
    }

    const pageProps: GetPageParameters = {
      page_id: args.page_id,
    }

    const res = await notion.retrievePage(pageProps)

    // Define columns for table output
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

    // Handle compact JSON output
    if (flags['compact-json']) {
      outputCompactJson(res)
      this.exit(0)
    }

    // Handle pretty table output
    if (flags.pretty) {
      outputPrettyTable([res], columns)
      this.exit(0)
    }

    // Handle raw JSON output
    if (flags.raw) {
      outputRawJson(res)
      this.exit(0)
    }

    // Handle table output (default)
    const options = {
      printLine: this.log.bind(this),
      ...flags,
    }
    ux.table([res], columns, options)
  }
}
