import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import {
  AppendBlockChildrenParameters,
  BlockObjectResponse,
} from '@notionhq/client/build/src/api-endpoints'
import { getBlockPlainText, outputRawJson } from '../../helper'

export default class BlockAppend extends Command {
  static description = 'Append block children'

  static aliases: string[] = ['block:a']

  static examples = [
    {
      description: 'Append block children',
      command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]'`,
    },
    {
      description: 'Append block children after a block',
      command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]' -a AFTER_BLOCK_ID`,
    },
    {
      description: 'Append block children and output raw json',
      command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]' -r`,
    },
  ]

  static flags = {
    block_id: Flags.string({
      char: 'b',
      description: 'Parent block ID',
      required: true,
    }),
    children: Flags.string({
      char: 'c',
      description: 'Block children (JSON array)',
      required: true,
    }),
    after: Flags.string({
      char: 'a',
      description: 'Block ID to append after (optional)',
    }),
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...ux.table.flags(),
  }

  // TODO: Add support children params building prompt
  public async run(): Promise<void> {
    const { flags } = await this.parse(BlockAppend)
    const params: AppendBlockChildrenParameters = {
      block_id: flags.block_id,
      children: JSON.parse(flags.children),
    }
    if (flags.after) {
      params.after = flags.after
    }
    const res = await notion.appendBlockChildren(params)
    if (flags.raw) {
      outputRawJson(res)
      this.exit(0)
    }

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
  }
}
