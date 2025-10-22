import { GetDataSourceResponse } from '@notionhq/client/build/src/api-endpoints'

/**
 * Property schema for AI agents - simplified and easy to parse
 */
export interface PropertySchema {
  name: string
  type: string
  description?: string
  required?: boolean
  options?: string[] // For select/multi-select
  config?: Record<string, any> // Additional configuration
}

/**
 * Database schema in AI-friendly format
 */
export interface DataSourceSchema {
  id: string
  title: string
  description?: string
  properties: PropertySchema[]
  url?: string
}

/**
 * Extract clean, AI-parseable schema from Notion data source response
 *
 * This transforms the complex nested Notion API structure into a flat,
 * easy-to-understand format that AI agents can work with directly.
 *
 * @param dataSource - Raw Notion data source response
 * @returns Simplified schema object
 */
export function extractSchema(dataSource: GetDataSourceResponse): DataSourceSchema {
  const properties: PropertySchema[] = []

  // Extract title from data source
  const title = extractTitle(dataSource)

  // Extract description if available
  const description = extractDescription(dataSource)

  // Process each property in the data source
  if (dataSource.properties) {
    for (const [propName, propConfig] of Object.entries(dataSource.properties)) {
      const schema = extractPropertySchema(propName, propConfig)
      if (schema) {
        properties.push(schema)
      }
    }
  }

  return {
    id: dataSource.id,
    title,
    description,
    properties,
    url: 'url' in dataSource ? dataSource.url : undefined,
  }
}

/**
 * Extract title from data source
 */
function extractTitle(dataSource: GetDataSourceResponse): string {
  if ('title' in dataSource && Array.isArray(dataSource.title)) {
    return dataSource.title
      .map((t: any) => t.plain_text || '')
      .join('')
      .trim() || 'Untitled'
  }
  return 'Untitled'
}

/**
 * Extract description from data source
 */
function extractDescription(dataSource: GetDataSourceResponse): string | undefined {
  if ('description' in dataSource && Array.isArray(dataSource.description)) {
    const desc = dataSource.description
      .map((d: any) => d.plain_text || '')
      .join('')
      .trim()
    return desc || undefined
  }
  return undefined
}

/**
 * Extract individual property schema
 */
function extractPropertySchema(name: string, config: any): PropertySchema | null {
  if (!config || !config.type) {
    return null
  }

  const schema: PropertySchema = {
    name,
    type: config.type,
  }

  // Handle select and multi-select with options
  if (config.type === 'select' && config.select?.options) {
    schema.options = config.select.options.map((opt: any) => opt.name)
    schema.description = `Select one: ${schema.options.join(', ')}`
  }

  if (config.type === 'multi_select' && config.multi_select?.options) {
    schema.options = config.multi_select.options.map((opt: any) => opt.name)
    schema.description = `Select multiple: ${schema.options.join(', ')}`
  }

  // Handle status property (similar to select)
  if (config.type === 'status' && config.status?.options) {
    schema.options = config.status.options.map((opt: any) => opt.name)
    schema.description = `Status: ${schema.options.join(', ')}`
  }

  // Handle formula properties
  if (config.type === 'formula' && config.formula?.expression) {
    schema.config = {
      expression: config.formula.expression,
    }
    schema.description = `Formula: ${config.formula.expression}`
  }

  // Handle rollup properties
  if (config.type === 'rollup') {
    schema.config = {
      relation_property: config.rollup?.relation_property_name,
      rollup_property: config.rollup?.rollup_property_name,
      function: config.rollup?.function,
    }
    schema.description = 'Rollup from related database'
  }

  // Handle relation properties
  if (config.type === 'relation') {
    schema.config = {
      database_id: config.relation?.database_id,
      type: config.relation?.type,
    }
    schema.description = 'Relation to another database'
  }

  // Handle number properties with format
  if (config.type === 'number' && config.number?.format) {
    schema.config = {
      format: config.number.format,
    }
    schema.description = `Number (${config.number.format})`
  }

  // Mark title property as required
  if (config.type === 'title') {
    schema.required = true
    schema.description = 'Title (required)'
  }

  return schema
}

/**
 * Filter properties by names
 *
 * @param schema - Full schema
 * @param propertyNames - Array of property names to include
 * @returns Filtered schema
 */
export function filterProperties(
  schema: DataSourceSchema,
  propertyNames: string[]
): DataSourceSchema {
  const lowerNames = propertyNames.map(n => n.toLowerCase())
  return {
    ...schema,
    properties: schema.properties.filter(
      p => lowerNames.includes(p.name.toLowerCase())
    ),
  }
}

/**
 * Format schema as human-readable table data
 *
 * @param schema - Schema to format
 * @returns Array of objects for table display
 */
export function formatSchemaForTable(schema: DataSourceSchema): Array<Record<string, string>> {
  return schema.properties.map(prop => ({
    name: prop.name,
    type: prop.type,
    required: prop.required ? 'Yes' : 'No',
    options: prop.options?.join(', ') || '-',
    description: prop.description || '-',
  }))
}

/**
 * Format schema as markdown documentation
 *
 * @param schema - Schema to format
 * @returns Markdown string
 */
export function formatSchemaAsMarkdown(schema: DataSourceSchema): string {
  const lines: string[] = []

  lines.push(`# ${schema.title}`)
  lines.push('')

  if (schema.description) {
    lines.push(schema.description)
    lines.push('')
  }

  lines.push(`**Database ID:** \`${schema.id}\``)
  if (schema.url) {
    lines.push(`**URL:** ${schema.url}`)
  }
  lines.push('')

  lines.push('## Properties')
  lines.push('')
  lines.push('| Name | Type | Required | Options/Details |')
  lines.push('|------|------|----------|-----------------|')

  for (const prop of schema.properties) {
    const required = prop.required ? '✓' : ''
    const details = prop.options?.join(', ') || prop.description || ''
    lines.push(`| ${prop.name} | ${prop.type} | ${required} | ${details} |`)
  }

  return lines.join('\n')
}

/**
 * Validate that a data object matches the schema
 *
 * @param schema - Schema to validate against
 * @param data - Data object to validate
 * @returns Validation result with errors
 */
export function validateAgainstSchema(
  schema: DataSourceSchema,
  data: Record<string, any>
): { valid: boolean; errors: string[] } {
  const errors: string[] = []

  // Check required properties
  for (const prop of schema.properties) {
    if (prop.required && !(prop.name in data)) {
      errors.push(`Missing required property: ${prop.name}`)
    }
  }

  // Check property types and options
  for (const [key, value] of Object.entries(data)) {
    const propSchema = schema.properties.find(p => p.name === key)

    if (!propSchema) {
      errors.push(`Unknown property: ${key}`)
      continue
    }

    // Validate select/multi-select options
    if (propSchema.options && propSchema.options.length > 0) {
      if (propSchema.type === 'select') {
        if (typeof value === 'string' && !propSchema.options.includes(value)) {
          errors.push(`Invalid option for ${key}: ${value}. Must be one of: ${propSchema.options.join(', ')}`)
        }
      }

      if (propSchema.type === 'multi_select') {
        if (Array.isArray(value)) {
          const invalidOptions = value.filter(v => !propSchema.options!.includes(v))
          if (invalidOptions.length > 0) {
            errors.push(`Invalid options for ${key}: ${invalidOptions.join(', ')}`)
          }
        }
      }
    }
  }

  return {
    valid: errors.length === 0,
    errors,
  }
}
