import { Command, Flags } from '@oclif/core'
import { tableFlags, formatTable } from '../../utils/table-formatter'
import { UserObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import * as notion from '../../notion'
import { outputRawJson, stripMetadata } from '../../helper'
import { AutomationFlags } from '../../base-flags'
import {
  NotionCLIError,
  wrapNotionError
} from '../../errors'

export default class UserList extends Command {
  static description = 'List all users'

  static aliases: string[] = ['user:l']

  static examples = [
    {
      description: 'List all users',
      command: `$ notion-cli user list`,
    },
    {
      description: 'List all users and output raw json',
      command: `$ notion-cli user list -r`,
    },
    {
      description: 'List all users and output JSON for automation',
      command: `$ notion-cli user list --json`,
    },
  ]

  static flags = {
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...tableFlags,
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { flags } = await this.parse(UserList)

    try {
      let res = await notion.listUser()

      // Apply minimal flag to strip metadata
      if (flags.minimal) {
        res = stripMetadata(res)
      }

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
      formatTable(res.results, columns, options)
      process.exit(0)
    } catch (error) {
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            resourceType: 'user',
            endpoint: 'users.list'
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
