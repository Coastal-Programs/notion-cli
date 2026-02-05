import { Command, Flags } from '@oclif/core'
import { UserObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import * as notion from '../../../notion'
import { outputRawJson } from '../../../helper'
import { AutomationFlags } from '../../../base-flags'
import {
  NotionCLIError,
  wrapNotionError
} from '../../../errors'
import { tableFlags, formatTable } from '../../../utils/table-formatter'

export default class UserRetrieveBot extends Command {
  static description = 'Retrieve a bot user'

  static aliases: string[] = ['user:r:b']

  static examples = [
    {
      description: 'Retrieve a bot user',
      command: `$ notion-cli user retrieve:bot`,
    },
    {
      description: 'Retrieve a bot user and output raw json',
      command: `$ notion-cli user retrieve:bot -r`,
    },
    {
      description: 'Retrieve a bot user and output JSON for automation',
      command: `$ notion-cli user retrieve:bot --json`,
    },
  ]

  static args = {}

  static flags = {
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...tableFlags,
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { flags } = await this.parse(UserRetrieveBot)

    try {
      const res = await notion.botUser()

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
        id: {},
        name: {},
        object: {},
        type: {},
        person_or_bot: {
          header: 'person/bot',
          get: (row: UserObjectResponse) => {
            if (row.type === 'person') {
              return row.person
            }
            return row.bot
          },
        },
        avatar_url: {},
      }
      const options = {
        printLine: this.log.bind(this),
        ...flags,
      }
      formatTable([res], columns, options)
      process.exit(0)
    } catch (error) {
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            resourceType: 'user',
            endpoint: 'users.me'
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
