import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import {
  UpdateBlockParameters,
  BlockObjectResponse,
} from '@notionhq/client/build/src/api-endpoints'
import { outputRawJson, getBlockPlainText, buildBlockUpdateFromTextFlags } from '../../helper'
import { resolveNotionId } from '../../utils/notion-resolver'
import { AutomationFlags } from '../../base-flags'
import {
  NotionCLIError,
  NotionCLIErrorFactory,
  wrapNotionError
} from '../../errors'

export default class BlockUpdate extends Command {
  static description = 'Update a block'

  static aliases: string[] = ['block:u']

  static examples = [
    {
      description: 'Update block with simple text',
      command: `$ notion-cli block update BLOCK_ID --text "Updated content"`,
    },
    {
      description: 'Update heading content',
      command: `$ notion-cli block update BLOCK_ID --heading-1 "New Title"`,
    },
    {
      description: 'Update code block',
      command: `$ notion-cli block update BLOCK_ID --code "const x = 42;" --language javascript`,
    },
    {
      description: 'Archive a block',
      command: `$ notion-cli block update BLOCK_ID -a`,
    },
    {
      description: 'Archive a block via URL',
      command: `$ notion-cli block update https://notion.so/BLOCK_ID -a`,
    },
    {
      description: 'Update block content with complex JSON (for advanced cases)',
      command: `$ notion-cli block update BLOCK_ID -c '{"paragraph":{"rich_text":[{"text":{"content":"Updated text"}}]}}'`,
    },
    {
      description: 'Update block color',
      command: `$ notion-cli block update BLOCK_ID --color blue`,
    },
    {
      description: 'Update a block and output raw json',
      command: `$ notion-cli block update BLOCK_ID --text "Updated" -r`,
    },
    {
      description: 'Update a block and output JSON for automation',
      command: `$ notion-cli block update BLOCK_ID --text "Updated" --json`,
    },
  ]

  static args = {
    block_id: Args.string({ description: 'Block ID or URL', required: true }),
  }

  static flags = {
    archived: Flags.boolean({
      char: 'a',
      description: 'Archive the block',
    }),
    content: Flags.string({
      char: 'c',
      description: 'Updated block content (JSON object with block type properties) - for complex cases',
    }),
    // Simple text-based flags
    text: Flags.string({
      description: 'Update paragraph text',
    }),
    'heading-1': Flags.string({
      description: 'Update H1 heading text',
    }),
    'heading-2': Flags.string({
      description: 'Update H2 heading text',
    }),
    'heading-3': Flags.string({
      description: 'Update H3 heading text',
    }),
    bullet: Flags.string({
      description: 'Update bulleted list item text',
    }),
    numbered: Flags.string({
      description: 'Update numbered list item text',
    }),
    todo: Flags.string({
      description: 'Update to-do item text',
    }),
    toggle: Flags.string({
      description: 'Update toggle block text',
    }),
    code: Flags.string({
      description: 'Update code block content',
    }),
    language: Flags.string({
      description: 'Update code block language (used with --code)',
      default: 'plain text',
    }),
    quote: Flags.string({
      description: 'Update quote block text',
    }),
    callout: Flags.string({
      description: 'Update callout block text',
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
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(BlockUpdate)

    try {
      // Resolve block ID from URL or direct ID
      const blockId = await resolveNotionId(args.block_id, 'page')

      const params: any = {
        block_id: blockId,
      }

      // Handle archived flag
      if (flags.archived !== undefined) {
        params.archived = flags.archived
      }

      // Check if using simple text-based flags or complex JSON
      const hasTextFlags = flags.text || flags['heading-1'] || flags['heading-2'] || flags['heading-3'] ||
                          flags.bullet || flags.numbered || flags.todo || flags.toggle ||
                          flags.code || flags.quote || flags.callout

      if (hasTextFlags && flags.content) {
        this.error('Cannot use both text-based flags (--text, --heading-1, etc.) and --content flag together. Choose one approach.')
      }

      // Handle content updates
      if (hasTextFlags) {
        // Use simple text-based flags
        const blockUpdate = buildBlockUpdateFromTextFlags('', {
          text: flags.text,
          heading1: flags['heading-1'],
          heading2: flags['heading-2'],
          heading3: flags['heading-3'],
          bullet: flags.bullet,
          numbered: flags.numbered,
          todo: flags.todo,
          toggle: flags.toggle,
          code: flags.code,
          language: flags.language,
          quote: flags.quote,
          callout: flags.callout,
        })

        if (blockUpdate) {
          Object.assign(params, blockUpdate)
        }
      } else if (flags.content) {
        // Use complex JSON
        try {
          const content = JSON.parse(flags.content)
          Object.assign(params, content)
        } catch (error: any) {
          throw NotionCLIErrorFactory.invalidJson(flags.content, error)
        }
      }

      // Handle color updates
      if (flags.color) {
        // Retrieve the block to determine its type
        const block = await notion.retrieveBlock(blockId)

        // Color is only supported for certain block types
        const colorSupportedTypes = [
          'paragraph', 'heading_1', 'heading_2', 'heading_3',
          'bulleted_list_item', 'numbered_list_item', 'toggle',
          'quote', 'callout'
        ]

        if (!colorSupportedTypes.includes(block.type)) {
          this.error(`Color property is not supported for block type: ${block.type}. Supported types: ${colorSupportedTypes.join(', ')}`)
        }

        // Color must be nested within the block type property
        params[block.type] = {
          ...params[block.type],
          color: flags.color
        }
      }

      const res = await notion.updateBlock(params as UpdateBlockParameters)

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
      ux.table([res], columns, options)
      process.exit(0)
    } catch (error) {
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            resourceType: 'block',
            attemptedId: args.block_id,
            endpoint: 'blocks.update',
            userInput: flags.content
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
