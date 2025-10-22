import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import {
  UpdateBlockParameters,
  BlockObjectResponse,
} from '@notionhq/client/build/src/api-endpoints'
import { outputRawJson, getBlockPlainText } from '../../helper'

export default class BlockUpdate extends Command {
  static description = 'Update a block'

  static aliases: string[] = ['block:u']

  static examples = [
    {
      description: 'Archive a block',
      command: `$ notion-cli block update BLOCK_ID -a`,
    },
    {
      description: 'Update block content',
      command: `$ notion-cli block update BLOCK_ID -c '{"paragraph":{"rich_text":[{"text":{"content":"Updated text"}}]}}'`,
    },
    {
      description: 'Update block color',
      command: `$ notion-cli block update BLOCK_ID --color blue`,
    },
    {
      description: 'Update a block and output raw json',
      command: `$ notion-cli block update BLOCK_ID -a -r`,
    },
  ]

  static args = {
    block_id: Args.string({ description: 'block_id', required: true }),
  }

  static flags = {
    archived: Flags.boolean({
      char: 'a',
      description: 'Archive the block',
    }),
    content: Flags.string({
      char: 'c',
      description: 'Updated block content (JSON object with block type properties)',
    }),
    color: Flags.string({
      description: 'Block color (for supported block types)',
      options: ['default', 'gray', 'brown', 'orange', 'yellow', 'green', 'blue', 'purple', 'pink', 'red'],
    }),
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...ux.table.flags(),
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(BlockUpdate)

    const params: any = {
      block_id: args.block_id,
    }

    // Handle archived flag
    if (flags.archived !== undefined) {
      params.archived = flags.archived
    }

    // Handle content updates
    if (flags.content) {
      try {
        const content = JSON.parse(flags.content)
        Object.assign(params, content)
      } catch (error) {
        this.error('Invalid JSON in --content flag. Please provide valid JSON.')
      }
    }

    // Handle color updates
    if (flags.color) {
      params.color = flags.color
    }

    const res = await notion.updateBlock(params as UpdateBlockParameters)
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
    ux.table([res], columns, options)
  }
}
