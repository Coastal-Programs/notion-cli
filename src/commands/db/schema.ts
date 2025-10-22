import { Args, Command, Flags } from '@oclif/core'
import * as notion from '../../notion'
import {
  extractSchema,
  filterProperties,
  formatSchemaForTable,
  formatSchemaAsMarkdown,
  DataSourceSchema,
} from '../../utils/schema-extractor'
import { wrapNotionError } from '../../errors'

export default class DbSchema extends Command {
  static description =
    'Extract clean, AI-parseable schema from a Notion data source (table). ' +
    'This command is optimized for AI agents and automation - it returns property names, ' +
    'types, options (for select/multi-select), and configuration in an easy-to-parse format.'

  static aliases: string[] = ['db:s', 'ds:schema', 'ds:s']

  static examples = [
    {
      description: 'Get full schema in JSON format (recommended for AI agents)',
      command: '<%= config.bin %> db schema abc123def456 --output json',
    },
    {
      description: 'Get schema as formatted table',
      command: '<%= config.bin %> db schema abc123def456',
    },
    {
      description: 'Get schema in YAML format',
      command: '<%= config.bin %> db schema abc123def456 --output yaml',
    },
    {
      description: 'Get only specific properties',
      command: '<%= config.bin %> db schema abc123def456 --properties Name,Status,Tags --output json',
    },
    {
      description: 'Get schema as markdown documentation',
      command: '<%= config.bin %> db schema abc123def456 --markdown',
    },
    {
      description: 'Parse schema with jq (extract property names)',
      command: '<%= config.bin %> db schema abc123def456 --output json | jq \'.data.properties[].name\'',
    },
    {
      description: 'Find all select/multi-select properties and their options',
      command: '<%= config.bin %> db schema abc123def456 --output json | jq \'.data.properties[] | select(.options) | {name, options}\'',
    },
  ]

  static args = {
    data_source_id: Args.string({
      required: true,
      description: 'Data source ID (the table whose schema you want to extract)',
    }),
  }

  static flags = {
    output: Flags.string({
      char: 'o',
      description: 'Output format',
      options: ['json', 'yaml', 'table'],
      default: 'table',
    }),
    properties: Flags.string({
      char: 'p',
      description: 'Comma-separated list of properties to include (default: all)',
    }),
    markdown: Flags.boolean({
      char: 'm',
      description: 'Output as markdown documentation',
      default: false,
    }),
    json: Flags.boolean({
      char: 'j',
      description: 'Output as JSON (shorthand for --output json)',
      default: false,
    }),
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(DbSchema)

    try {
      // Fetch data source from Notion (uses caching)
      const dataSource = await notion.retrieveDataSource(args.data_source_id)

      // Extract clean schema
      let schema: DataSourceSchema = extractSchema(dataSource)

      // Filter properties if specified
      if (flags.properties) {
        const propertyNames = flags.properties.split(',').map(p => p.trim())
        schema = filterProperties(schema, propertyNames)
      }

      // Determine output format
      const outputFormat = flags.json ? 'json' : flags.output

      // Handle markdown output
      if (flags.markdown) {
        const markdown = formatSchemaAsMarkdown(schema)
        this.log(markdown)
        process.exit(0)
        return
      }

      // Handle JSON output (for AI agents)
      if (outputFormat === 'json') {
        this.log(
          JSON.stringify(
            {
              success: true,
              data: schema,
              timestamp: new Date().toISOString(),
            },
            null,
            2
          )
        )
        process.exit(0)
        return
      }

      // Handle YAML output
      if (outputFormat === 'yaml') {
        const yaml = this.formatAsYaml(schema)
        this.log(yaml)
        process.exit(0)
        return
      }

      // Handle table output (default)
      this.outputTable(schema)
      process.exit(0)
    } catch (error) {
      const cliError = wrapNotionError(error)

      if (flags.json || flags.output === 'json') {
        this.log(JSON.stringify(cliError.toJSON(), null, 2))
      } else {
        this.error(cliError.message)
      }

      process.exit(1)
    }
  }

  /**
   * Output schema as formatted table
   */
  private outputTable(schema: DataSourceSchema): void {
    this.log(`\n📋 ${schema.title}`)
    if (schema.description) {
      this.log(`   ${schema.description}`)
    }
    this.log(`   ID: ${schema.id}`)
    if (schema.url) {
      this.log(`   URL: ${schema.url}`)
    }
    this.log('')

    if (schema.properties.length === 0) {
      this.log('   No properties found.')
      return
    }

    // Calculate column widths
    const nameWidth = Math.max(
      20,
      ...schema.properties.map(p => p.name.length)
    )
    const typeWidth = Math.max(
      12,
      ...schema.properties.map(p => p.type.length)
    )

    // Print header
    this.log(
      `   ${'Name'.padEnd(nameWidth)} | ${'Type'.padEnd(typeWidth)} | Req | Details`
    )
    this.log(
      `   ${'-'.repeat(nameWidth)}-+-${'-'.repeat(typeWidth)}-+-----+---------`
    )

    // Print properties
    for (const prop of schema.properties) {
      const name = prop.name.padEnd(nameWidth)
      const type = prop.type.padEnd(typeWidth)
      const required = prop.required ? ' ✓ ' : '   '
      const details = prop.options
        ? prop.options.slice(0, 3).join(', ') +
          (prop.options.length > 3 ? '...' : '')
        : prop.description || ''

      this.log(`   ${name} | ${type} | ${required} | ${details}`)
    }

    this.log('')
  }

  /**
   * Format schema as YAML
   */
  private formatAsYaml(schema: DataSourceSchema): string {
    const lines: string[] = []

    lines.push('id: ' + schema.id)
    lines.push('title: ' + schema.title)
    if (schema.description) {
      lines.push('description: ' + schema.description)
    }
    if (schema.url) {
      lines.push('url: ' + schema.url)
    }
    lines.push('properties:')

    for (const prop of schema.properties) {
      lines.push(`  - name: ${prop.name}`)
      lines.push(`    type: ${prop.type}`)
      if (prop.required) {
        lines.push(`    required: true`)
      }
      if (prop.options && prop.options.length > 0) {
        lines.push(`    options:`)
        for (const opt of prop.options) {
          lines.push(`      - ${opt}`)
        }
      }
      if (prop.description) {
        lines.push(`    description: ${prop.description}`)
      }
      if (prop.config) {
        lines.push(`    config:`)
        for (const [key, value] of Object.entries(prop.config)) {
          lines.push(`      ${key}: ${value}`)
        }
      }
    }

    return lines.join('\n')
  }
}
