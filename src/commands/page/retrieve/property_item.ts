import { Args, Command, Flags } from '@oclif/core'
import * as notion from '../../../notion'
import { outputRawJson } from '../../../helper'
import { AutomationFlags } from '../../../base-flags'
import {
  NotionCLIError,
  wrapNotionError
} from '../../../errors'

export default class PageRetrievePropertyItem extends Command {
  static description = 'Retrieve a page property item'

  static aliases: string[] = ['page:r:pi']

  static examples = [
    {
      description: 'Retrieve a page property item',
      command: `$ notion-cli page retrieve:property_item PAGE_ID PROPERTY_ID`,
    },
    {
      description: 'Retrieve a page property item and output raw json',
      command: `$ notion-cli page retrieve:property_item PAGE_ID PROPERTY_ID -r`,
    },
    {
      description: 'Retrieve a page property item and output JSON for automation',
      command: `$ notion-cli page retrieve:property_item PAGE_ID PROPERTY_ID --json`,
    },
  ]

  static args = {
    page_id: Args.string({ required: true }),
    property_id: Args.string({ required: true }),
  }

  static flags = {
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(PageRetrievePropertyItem)

    try {
      const res = await notion.retrievePageProperty(args.page_id, args.property_id)

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

      // Handle raw JSON output (default for this command)
      outputRawJson(res)
      process.exit(0)
    } catch (error) {
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            resourceType: 'page',
            attemptedId: args.page_id,
            endpoint: 'pages.properties.retrieve'
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
