import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import { GetPageParameters, PageObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import {
  getPageTitle,
  outputRawJson,
  outputCompactJson,
  outputMarkdownTable,
  outputPrettyTable,
  showRawFlagHint,
  stripMetadata
} from '../../helper'
import { NotionToMarkdown } from 'notion-to-md'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'
import { resolveNotionId } from '../../utils/notion-resolver'
import { wrapNotionError } from '../../errors'

export default class PageRetrieve extends Command {
  static description = 'Retrieve a page'

  static aliases: string[] = ['page:r']

  static examples = [
    {
      description: 'Retrieve a page with full data (recommended for AI assistants)',
      command: `$ notion-cli page retrieve PAGE_ID -r`,
    },
    {
      description: 'Retrieve entire page tree with all nested content (35% token reduction)',
      command: `$ notion-cli page retrieve PAGE_ID --recursive --compact-json`,
    },
    {
      description: 'Retrieve page tree with custom depth limit',
      command: `$ notion-cli page retrieve PAGE_ID -R --max-depth 5 --json`,
    },
    {
      description: 'Retrieve a page and output table',
      command: `$ notion-cli page retrieve PAGE_ID`,
    },
    {
      description: 'Retrieve a page via URL',
      command: `$ notion-cli page retrieve https://notion.so/PAGE_ID`,
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
    {
      description: 'Retrieve a page and output JSON for automation',
      command: `$ notion-cli page retrieve PAGE_ID --json`,
    },
  ]

  static args = {
    page_id: Args.string({
      required: true,
      description: 'Page ID or full Notion URL (e.g., https://notion.so/...)',
    }),
  }

  static flags = {
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json (recommended for AI assistants - returns all fields)',
    }),
    markdown: Flags.boolean({
      char: 'm',
      description: 'output page content as markdown',
    }),
    recursive: Flags.boolean({
      char: 'R',
      description: 'recursively fetch all blocks and nested pages (reduces API calls)',
      default: false,
    }),
    'max-depth': Flags.integer({
      description: 'maximum recursion depth for --recursive (default: 3)',
      default: 3,
      min: 1,
      max: 10,
      dependsOn: ['recursive'],
    }),
    ...ux.table.flags(),
    ...OutputFormatFlags,
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(PageRetrieve)

    try {
      // Resolve ID from URL, direct ID, or name (future)
      const pageId = await resolveNotionId(args.page_id, 'page')

      // Handle page content as markdown (uses NotionToMarkdown)
      if (flags.markdown) {
        const n2m = new NotionToMarkdown({ notionClient: notion.client })
        const mdBlocks = await n2m.pageToMarkdown(pageId)
        const mdString = n2m.toMarkdownString(mdBlocks)
        console.log(mdString.parent)
        process.exit(0)
        return
      }

      // Handle recursive fetching
      if (flags.recursive) {
        const recursiveData = await notion.retrievePageRecursive(
          pageId,
          0,
          flags['max-depth']
        )

        // Handle JSON output for automation (takes precedence)
        if (flags.json) {
          this.log(JSON.stringify({
            success: true,
            data: recursiveData,
            timestamp: new Date().toISOString()
          }, null, 2))
          process.exit(0)
          return
        }

        // Handle compact JSON output
        if (flags['compact-json']) {
          outputCompactJson(recursiveData)
          process.exit(0)
          return
        }

        // Handle raw JSON output
        if (flags.raw) {
          outputRawJson(recursiveData)
          process.exit(0)
          return
        }

        // For other formats, show a message that they're not supported with recursive
        this.error('Recursive mode only supports --json, --compact-json, or --raw output formats')
        process.exit(1)
        return
      }

      const pageProps: GetPageParameters = {
        page_id: pageId,
      }

      let res = await notion.retrievePage(pageProps)

      // Apply minimal flag to strip metadata
      if (flags.minimal) {
        res = stripMetadata(res)
      }

      // Handle JSON output for automation (takes precedence)
      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          data: res,
          timestamp: new Date().toISOString()
        }, null, 2))
        process.exit(0)
        return
      }

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

      // Handle raw JSON output
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
