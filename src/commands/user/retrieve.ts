import { Args, Command, Flags, ux } from '@oclif/core'
import { UserObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import * as notion from '../../notion'
import { outputRawJson } from '../../helper'
import { AutomationFlags } from '../../base-flags'
import { wrapNotionError } from '../../errors'

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
    ...ux.table.flags(),
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(UserRetrieve)

    try {
      const res = await notion.retrieveUser(args.user_id)

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
      ux.table([res], columns, options)
      process.exit(0)
    } catch (error) {
      const cliError = wrapNotionError(error)
      if (flags.json) {
        this.log(JSON.stringify(cliError.toJSON(), null, 2))
      } else {
        this.error(cliError.message)
      }
      process.exit(1)
    }
  }
}
