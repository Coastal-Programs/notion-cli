import { Command, Flags } from '@oclif/core'
import { tableFlags, formatTable } from '../../utils/table-formatter'
import * as notion from '../../notion'
import {
  AppendBlockChildrenParameters,
  BlockObjectResponse,
} from '@notionhq/client/build/src/api-endpoints'
import { getBlockPlainText, outputRawJson, buildBlocksFromTextFlags } from '../../helper'
import { resolveNotionId } from '../../utils/notion-resolver'
import { AutomationFlags } from '../../base-flags'
import {
  NotionCLIError,
  NotionCLIErrorFactory,
  wrapNotionError
} from '../../errors'

export default class BlockAppend extends Command {
  static description = 'Append block children'

  static aliases: string[] = ['block:a']

  static examples = [
    {
      description: 'Append a simple paragraph',
      command: `$ notion-cli block append -b BLOCK_ID --text "Hello world!"`,
    },
    {
      description: 'Append a heading',
      command: `$ notion-cli block append -b BLOCK_ID --heading-1 "Chapter Title"`,
    },
    {
      description: 'Append a bullet point',
      command: `$ notion-cli block append -b BLOCK_ID --bullet "First item"`,
    },
    {
      description: 'Append a code block',
      command: `$ notion-cli block append -b BLOCK_ID --code "console.log('test')" --language javascript`,
    },
    {
      description: 'Append block children with complex JSON (for advanced cases)',
      command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]'`,
    },
    {
      description: 'Append block children via URL',
      command: `$ notion-cli block append -b https://notion.so/BLOCK_ID --text "Hello world!"`,
    },
    {
      description: 'Append block children after a block',
      command: `$ notion-cli block append -b BLOCK_ID --text "Hello world!" -a AFTER_BLOCK_ID`,
    },
    {
      description: 'Append block children and output raw json',
      command: `$ notion-cli block append -b BLOCK_ID --text "Hello world!" -r`,
    },
    {
      description: 'Append block children and output JSON for automation',
      command: `$ notion-cli block append -b BLOCK_ID --text "Hello world!" --json`,
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
      description: 'Block children (JSON array) - for complex cases',
    }),
    // Simple text-based flags
    text: Flags.string({
      description: 'Paragraph text',
    }),
    'heading-1': Flags.string({
      description: 'H1 heading text',
    }),
    'heading-2': Flags.string({
      description: 'H2 heading text',
    }),
    'heading-3': Flags.string({
      description: 'H3 heading text',
    }),
    bullet: Flags.string({
      description: 'Bulleted list item text',
    }),
    numbered: Flags.string({
      description: 'Numbered list item text',
    }),
    todo: Flags.string({
      description: 'To-do item text',
    }),
    toggle: Flags.string({
      description: 'Toggle block text',
    }),
    code: Flags.string({
      description: 'Code block content',
    }),
    language: Flags.string({
      description: 'Code block language (used with --code)',
      default: 'plain text',
    }),
    quote: Flags.string({
      description: 'Quote block text',
    }),
    callout: Flags.string({
      description: 'Callout block text',
    }),
    after: Flags.string({
      char: 'a',
      description: 'Block ID or URL to append after (optional)',
    }),
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...tableFlags,
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { flags } = await this.parse(BlockAppend)

    try {
      // Resolve block ID from URL or direct ID
      const blockId = await resolveNotionId(flags.block_id, 'page')

      let children: any[]

      // Check if using simple text-based flags or complex JSON
      const hasTextFlags = flags.text || flags['heading-1'] || flags['heading-2'] || flags['heading-3'] ||
                          flags.bullet || flags.numbered || flags.todo || flags.toggle ||
                          flags.code || flags.quote || flags.callout

      if (hasTextFlags && flags.children) {
        this.error('Cannot use both text-based flags (--text, --heading-1, etc.) and --children flag together. Choose one approach.')
      }

      if (hasTextFlags) {
        // Use simple text-based flags
        children = buildBlocksFromTextFlags({
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

        if (children.length === 0) {
          this.error('No content provided. Use text-based flags (--text, --heading-1, etc.) or --children flag.')
        }
      } else if (flags.children) {
        // Use complex JSON
        try {
          children = JSON.parse(flags.children)
        } catch (error: any) {
          throw NotionCLIErrorFactory.invalidJson(flags.children, error)
        }
      } else {
        this.error('No content provided. Use text-based flags (--text, --heading-1, etc.) or --children flag.')
      }

      const params: AppendBlockChildrenParameters = {
        block_id: blockId,
        children: children,
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
      formatTable(res.results, columns, options)
      process.exit(0)
    } catch (error) {
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            resourceType: 'block',
            attemptedId: flags.block_id,
            endpoint: 'blocks.children.append',
            userInput: flags.children
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
