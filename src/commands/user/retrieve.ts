import { Args, Command, Flags } from '@oclif/core'
import { tableFlags, formatTable } from '../../utils/table-formatter'
import { UserObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import * as notion from '../../notion'
import { outputRawJson, stripMetadata } from '../../helper'
import { AutomationFlags } from '../../base-flags'
import {
  NotionCLIError,
  wrapNotionError
} from '../../errors'

export default class UserRetrieve extends Command {
  static description = 'Retrieve a user'

  static aliases: string[] = ['user:r']

  static examples = [
    {
      description: 'Retrieve a user',
      command: `$ notion-cli user retrieve USER_ID`,
    },
    {
      description: 'Retrieve a user and output raw json',
      command: `$ notion-cli user retrieve USER_ID -r`,
    },
    {
      description: 'Retrieve a user and output JSON for automation',
      command: `$ notion-cli user retrieve USER_ID --json`,
    },
  ]

  static args = {
    user_id: Args.string(),
  }

  static flags = {
    raw: Flags.boolean({
      char: 'r',
      description: 'output raw json',
    }),
    ...tableFlags,
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(UserRetrieve)

    try {
      let res = await notion.retrieveUser(args.user_id)

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
      formatTable([res], columns, options)
      process.exit(0)
    } catch (error) {
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            resourceType: 'user',
            attemptedId: args.user_id,
            endpoint: 'users.retrieve'
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
