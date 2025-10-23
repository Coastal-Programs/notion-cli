import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import {
  AppendBlockChildrenParameters,
  BlockObjectResponse,
} from '@notionhq/client/build/src/api-endpoints'
import { getBlockPlainText, outputRawJson } from '../../helper'
import { resolveNotionId } from '../../utils/notion-resolver'
import { AutomationFlags } from '../../base-flags'
import { handleCliError } from '../../errors'

export default class BlockAppend extends Command {
  static description = 'Append block children'

  static aliases: string[] = ['block:a']

  static examples = [
    {
      description: 'Append block children',
      command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]'`,
    },
    {
      description: 'Append block children via URL',
      command: `$ notion-cli block append -b https://notion.so/BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]'`,
    },
    {
      description: 'Append block children after a block',
      command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]' -a AFTER_BLOCK_ID`,
    },
    {
      description: 'Append block children and output raw json',
      command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]' -r`,
    },
    {
      description: 'Append block children and output JSON for automation',
      command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]' --json`,
    },
  ]

  static flags = {
    block_id: Flags.string({
      char: 'b',
      description: 'Parent block ID or URL',
      required: true,
    }),
    children: Flags.string({
      char: 'c',
      description: 'Block children (JSON array)',
      required: true,
    }),
    after: Flags.string({
      char: 'a',
      description: 'Block ID or URL to append after (optional)',
    }),
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...ux.table.flags(),
    ...AutomationFlags,
  }

  // TODO: Add support children params building prompt
  public async run(): Promise<void> {
    const { flags } = await this.parse(BlockAppend)

    try {
      // Resolve block ID from URL or direct ID
      const blockId = await resolveNotionId(flags.block_id, 'page')

      const params: AppendBlockChildrenParameters = {
        block_id: blockId,
        children: JSON.parse(flags.children),
      }

      if (flags.after) {
        // Resolve after block ID from URL or direct ID
        const afterBlockId = await resolveNotionId(flags.after, 'page')
        params.after = afterBlockId
      }

      const res = await notion.appendBlockChildren(params)

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
        object: {},
        id: {},
        type: {},
        parent: {},
        content: {
          get: (row: BlockObjectResponse) => {
            return getBlockPlainText(row)
          },
        },
      }
      const options = {
        printLine: this.log.bind(this),
        ...flags,
      }
      ux.table(res.results, columns, options)
      process.exit(0)
    } catch (error) {
      handleCliError(error, flags.json, {
        resourceType: 'block',
        attemptedId: flags['block-id'],
        endpoint: 'blocks.children.append'
      })
    }
  }
}
