import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import {
  outputRawJson,
  outputCompactJson,
  getPageTitle,
  getDataSourceTitle,
  getBlockPlainText,
} from '../../helper'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'
import {
  NotionCLIError,
  NotionCLIErrorCode,
  wrapNotionError
} from '../../errors'
import { PageObjectResponse, BlockObjectResponse, GetDataSourceResponse } from '@notionhq/client/build/src/api-endpoints'
import * as readline from 'readline'

type RetrieveResult = {
  id: string
  success: boolean
  data?: PageObjectResponse | BlockObjectResponse | GetDataSourceResponse
  error?: string
  message?: string
}

export default class BatchRetrieve extends Command {
  static description = 'Batch retrieve multiple pages, blocks, or data sources'

  static aliases: string[] = ['batch:r']

  static examples = [
    {
      description: 'Retrieve multiple pages via --ids flag',
      command: '$ notion-cli batch retrieve --ids PAGE_ID_1,PAGE_ID_2,PAGE_ID_3 --compact-json',
    },
    {
      description: 'Retrieve multiple pages from stdin (one ID per line)',
      command: '$ cat page_ids.txt | notion-cli batch retrieve --compact-json',
    },
    {
      description: 'Retrieve multiple blocks',
      command: '$ notion-cli batch retrieve --ids BLOCK_ID_1,BLOCK_ID_2 --type block --json',
    },
    {
      description: 'Retrieve multiple data sources',
      command: '$ notion-cli batch retrieve --ids DS_ID_1,DS_ID_2 --type database --json',
    },
    {
      description: 'Retrieve with raw output',
      command: '$ notion-cli batch retrieve --ids ID1,ID2,ID3 -r',
    },
  ]

  static args = {
    ids: Args.string({
      required: false,
      description: 'Comma-separated list of IDs to retrieve (or use --ids flag or stdin)',
    }),
  }

  static flags = {
    ids: Flags.string({
      description: 'Comma-separated list of IDs to retrieve',
    }),
    type: Flags.string({
      description: 'Resource type to retrieve (page, block, database)',
      options: ['page', 'block', 'database'],
      default: 'page',
    }),
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json (recommended for AI assistants - returns all fields)',
    }),
    ...ux.table.flags(),
    ...OutputFormatFlags,
    ...AutomationFlags,
  }

  /**
   * Read IDs from stdin
   */
  private async readStdin(): Promise<string[]> {
    return new Promise((resolve, reject) => {
      const ids: string[] = []
      const rl = readline.createInterface({
        input: process.stdin,
        output: process.stdout,
        terminal: false,
      })

      rl.on('line', (line) => {
        const trimmed = line.trim()
        if (trimmed) {
          ids.push(trimmed)
        }
      })

      rl.on('close', () => {
        resolve(ids)
      })

      rl.on('error', (err) => {
        reject(err)
      })

      // Timeout after 5 seconds if no input
      setTimeout(() => {
        rl.close()
        resolve(ids)
      }, 5000)
    })
  }

  /**
   * Retrieve a single resource and handle errors
   */
  private async retrieveResource(id: string, type: string): Promise<RetrieveResult> {
    try {
      let data: PageObjectResponse | BlockObjectResponse | GetDataSourceResponse

      switch (type) {
        case 'page':
          data = await notion.retrievePage({ page_id: id })
          break
        case 'block':
          data = await notion.retrieveBlock(id)
          break
        case 'database':
          data = await notion.retrieveDataSource(id)
          break
        default:
          throw new NotionCLIError(
            NotionCLIErrorCode.VALIDATION_ERROR,
            `Invalid resource type: ${type}`,
            [],
            { userInput: type, resourceType: type as any }
          )
      }

      return {
        id,
        success: true,
        data,
      }
    } catch (error) {
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            attemptedId: id,
            userInput: id
          })

      return {
        id,
        success: false,
        error: cliError.code,
        message: cliError.userMessage,
      }
    }
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(BatchRetrieve)

    try {
      // Get IDs from args, flags, or stdin
      let ids: string[] = []

      if (args.ids) {
        // From positional argument
        ids = args.ids.split(',').map(id => id.trim()).filter(id => id)
      } else if (flags.ids) {
        // From --ids flag
        ids = flags.ids.split(',').map(id => id.trim()).filter(id => id)
      } else if (!process.stdin.isTTY) {
        // From stdin
        ids = await this.readStdin()
      }

      if (ids.length === 0) {
        throw new NotionCLIError(
          NotionCLIErrorCode.VALIDATION_ERROR,
          'No IDs provided. Use --ids flag, positional argument, or pipe IDs via stdin',
          [
            {
              description: 'Provide IDs via --ids flag',
              command: 'notion-cli batch retrieve --ids ID1,ID2,ID3'
            },
            {
              description: 'Or pipe IDs from a file',
              command: 'cat ids.txt | notion-cli batch retrieve'
            }
          ]
        )
      }

      // Fetch all resources in parallel
      const results = await Promise.all(
        ids.map(id => this.retrieveResource(id, flags.type))
      )

      // Count successes and failures
      const successCount = results.filter(r => r.success).length
      const failureCount = results.filter(r => !r.success).length

      // Handle JSON output for automation (takes precedence)
      if (flags.json) {
        this.log(JSON.stringify({
          success: successCount > 0,
          total: results.length,
          succeeded: successCount,
          failed: failureCount,
          results: results,
          timestamp: new Date().toISOString(),
        }, null, 2))
        process.exit(failureCount === 0 ? 0 : 1)
        return
      }

      // Handle compact JSON output
      if (flags['compact-json']) {
        outputCompactJson({
          total: results.length,
          succeeded: successCount,
          failed: failureCount,
          results: results,
        })
        process.exit(failureCount === 0 ? 0 : 1)
        return
      }

      // Handle raw JSON output
      if (flags.raw) {
        outputRawJson(results)
        process.exit(failureCount === 0 ? 0 : 1)
        return
      }

      // Handle table output (default)
      const tableData = results.map(result => {
        if (result.success && result.data) {
          let title = ''
          if ('object' in result.data) {
            if (result.data.object === 'page') {
              title = getPageTitle(result.data as PageObjectResponse)
            } else if (result.data.object === 'database') {
              title = getDataSourceTitle(result.data as GetDataSourceResponse)
            } else if (result.data.object === 'block') {
              title = getBlockPlainText(result.data as BlockObjectResponse)
            }
          }

          return {
            id: result.id,
            status: 'success',
            type: result.data.object || flags.type,
            title: title || '-',
          }
        } else {
          return {
            id: result.id,
            status: 'failed',
            type: flags.type,
            title: result.message || result.error || 'Unknown error',
          }
        }
      })

      const columns = {
        id: {},
        status: {},
        type: {},
        title: {},
      }

      const options = {
        printLine: this.log.bind(this),
        ...flags,
      }

      ux.table(tableData, columns, options)

      // Print summary
      this.log(`\nTotal: ${results.length} | Succeeded: ${successCount} | Failed: ${failureCount}`)

      process.exit(failureCount === 0 ? 0 : 1)
    } catch (error) {
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            endpoint: 'batch.retrieve'
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
