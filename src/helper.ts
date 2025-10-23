import {
  QueryDataSourceResponse,
  GetDatabaseResponse,
  GetDataSourceResponse,
  DatabaseObjectResponse,
  DataSourceObjectResponse,
  PageObjectResponse,
  BlockObjectResponse,
} from '@notionhq/client/build/src/api-endpoints'
import { IPromptChoice } from './interface'
import * as notion from './notion'
import { isFullPage, isFullDatabase, isFullDataSource } from '@notionhq/client'

export const outputRawJson = async (res: any) => {
  console.log(JSON.stringify(res, null, 2))
}

/**
 * Output data as compact JSON (single-line, no formatting)
 * Useful for piping to other tools or scripts
 */
export const outputCompactJson = (res: any) => {
  console.log(JSON.stringify(res))
}

/**
 * Strip unnecessary metadata from Notion API responses to reduce size
 * Removes created_by, last_edited_by, object fields, request_id, empty values, etc.
 * Keeps timestamps (created_time, last_edited_time) and essential data
 *
 * @param data The data to strip metadata from (single object or array)
 * @returns The stripped data
 */
export const stripMetadata = (data: any): any => {
  if (Array.isArray(data)) {
    return data.map(item => stripMetadata(item))
  }

  if (data === null || typeof data !== 'object') {
    return data
  }

  const result: any = {}

  for (const [key, value] of Object.entries(data)) {
    // Skip fields that should be removed
    if (
      key === 'created_by' ||
      key === 'last_edited_by' ||
      key === 'request_id' ||
      key === 'object' ||
      (key === 'has_more' && value === false)
    ) {
      continue
    }

    // Skip empty arrays
    if (Array.isArray(value) && value.length === 0) {
      continue
    }

    // Skip empty objects (but keep objects with properties)
    if (value && typeof value === 'object' && !Array.isArray(value) && Object.keys(value).length === 0) {
      continue
    }

    // Recursively strip metadata from nested objects and arrays
    if (value && typeof value === 'object') {
      result[key] = stripMetadata(value)
    } else {
      result[key] = value
    }
  }

  return result
}

/**
 * Output data as a markdown table
 * Converts column data into GitHub-flavored markdown table format
 */
export const outputMarkdownTable = (data: any[], columns: Record<string, any>) => {
  if (!data || data.length === 0) {
    console.log('No data to display')
    return
  }

  // Extract column headers
  const headers = Object.keys(columns)

  // Build header row
  const headerRow = '| ' + headers.join(' | ') + ' |'
  const separatorRow = '| ' + headers.map(() => '---').join(' | ') + ' |'

  console.log(headerRow)
  console.log(separatorRow)

  // Build data rows
  data.forEach((row) => {
    const values = headers.map((header) => {
      const column = columns[header]
      let value: any

      // Handle column getter function
      if (column.get && typeof column.get === 'function') {
        value = column.get(row)
      } else if (column.header) {
        // If column has a header property, use the key to get value
        value = row[header]
      } else {
        // Direct property access
        value = row[header]
      }

      // Format value for markdown (escape pipes and handle nulls)
      if (value === null || value === undefined) {
        return ''
      }

      const stringValue = String(value).replace(/\|/g, '\\|').replace(/\n/g, ' ')
      return stringValue
    })

    console.log('| ' + values.join(' | ') + ' |')
  })
}

/**
 * Output data as a pretty table with borders
 * Enhanced table format with better visual separation
 */
export const outputPrettyTable = (data: any[], columns: Record<string, any>) => {
  if (!data || data.length === 0) {
    console.log('No data to display')
    return
  }

  const headers = Object.keys(columns)

  // Calculate column widths
  const columnWidths: Record<string, number> = {}

  headers.forEach((header) => {
    columnWidths[header] = header.length
  })

  // Calculate max width for each column based on data
  data.forEach((row) => {
    headers.forEach((header) => {
      const column = columns[header]
      let value: any

      if (column.get && typeof column.get === 'function') {
        value = column.get(row)
      } else {
        value = row[header]
      }

      const stringValue = String(value === null || value === undefined ? '' : value)
      columnWidths[header] = Math.max(columnWidths[header], stringValue.length)
    })
  })

  // Build separator line
  const topBorder = '┌' + headers.map(h => '─'.repeat(columnWidths[h] + 2)).join('┬') + '┐'
  const headerSeparator = '├' + headers.map(h => '─'.repeat(columnWidths[h] + 2)).join('┼') + '┤'
  const bottomBorder = '└' + headers.map(h => '─'.repeat(columnWidths[h] + 2)).join('┴') + '┘'

  // Print top border
  console.log(topBorder)

  // Print headers
  const headerRow = '│ ' + headers.map(h => h.padEnd(columnWidths[h])).join(' │ ') + ' │'
  console.log(headerRow)
  console.log(headerSeparator)

  // Print data rows
  data.forEach((row) => {
    const values = headers.map((header) => {
      const column = columns[header]
      let value: any

      if (column.get && typeof column.get === 'function') {
        value = column.get(row)
      } else {
        value = row[header]
      }

      const stringValue = String(value === null || value === undefined ? '' : value)
      return stringValue.padEnd(columnWidths[header])
    })

    console.log('│ ' + values.join(' │ ') + ' │')
  })

  // Print bottom border
  console.log(bottomBorder)
}

/**
 * Show a hint to users (especially AI assistants) that more data is available with the -r flag
 * This makes the -r flag more discoverable for automation and AI use cases
 *
 * @param itemCount Number of items displayed in the table
 * @param item The item object to count total fields from
 * @param visibleFields Number of fields shown in the table (default: 4 for title, object, id, url)
 */
export function showRawFlagHint(itemCount: number, item: any, visibleFields: number = 4): void {
  // Count total fields in the item
  let totalFields = visibleFields // Start with the visible fields (title, object, id, url)

  if (item) {
    // For pages and databases, count properties
    if (item.properties) {
      totalFields += Object.keys(item.properties).length
    }
    // Add other top-level metadata fields
    const metadataFields = ['created_time', 'last_edited_time', 'created_by', 'last_edited_by', 'parent', 'archived', 'icon', 'cover']
    metadataFields.forEach(field => {
      if (item[field] !== undefined) {
        totalFields++
      }
    })
  }

  const hiddenFields = totalFields - visibleFields

  if (hiddenFields > 0) {
    const itemText = itemCount === 1 ? 'item' : 'items'
    console.log(`\nTip: Showing ${visibleFields} of ${totalFields} fields for ${itemCount} ${itemText}.`)
    console.log(`Use -r flag for full JSON output with all properties (recommended for AI assistants and automation).`)
  }
}

export const getFilterFields = async (type: string) => {
  switch (type) {
    case 'checkbox':
      return [{ title: 'equals' }, { title: 'does_not_equal' }]
    case 'created_time':
    case 'last_edited_time':
    case 'date':
      return [
        { title: 'after' },
        { title: 'before' },
        { title: 'equals' },
        { title: 'is_empty' },
        { title: 'is_not_empty' },
        { title: 'next_month' },
        { title: 'next_week' },
        { title: 'next_year' },
        { title: 'on_or_after' },
        { title: 'on_or_before' },
        { title: 'past_month' },
        { title: 'past_week' },
        { title: 'past_year' },
        { title: 'this_week' },
      ]
    case 'rich_text':
    case 'title':
      return [
        { title: 'contains' },
        { title: 'does_not_contain' },
        { title: 'does_not_equal' },
        { title: 'ends_with' },
        { title: 'equals' },
        { title: 'is_empty' },
        { title: 'is_not_empty' },
        { title: 'starts_with' },
      ]
    case 'number':
      return [
        { title: 'equals' },
        { title: 'does_not_equal' },
        { title: 'greater_than' },
        { title: 'greater_than_or_equal_to' },
        { title: 'less_than' },
        { title: 'less_than_or_equal_to' },
        { title: 'is_empty' },
        { title: 'is_not_empty' },
      ]
    case 'select':
      return [
        { title: 'equals' },
        { title: 'does_not_equal' },
        { title: 'is_empty' },
        { title: 'is_not_empty' },
      ]
    case 'multi_select':
    case 'relation':
      return [
        { title: 'contains' },
        { title: 'does_not_contain' },
        { title: 'is_empty' },
        { title: 'is_not_empty' },
      ]
    case 'status':
      return [
        { title: 'equals' },
        { title: 'does_not_equal' },
        { title: 'is_empty' },
        { title: 'is_not_empty' },
      ]
    case 'files':
    case 'formula':
    case 'people':
    case 'rollup':
    default:
      console.error(`type: ${type} is not support type`)
      return null
  }
}



export const buildDatabaseQueryFilter = async (
  name: string,
  type: string,
  field: string,
  value: string | string[] | boolean
): Promise<object | null> => {
  let filter = null
  switch (type) {
    case 'checkbox':
      filter = {
        property: name,
        [type]: {
          // boolean value
          [field]: value == 'true',
        },
      }
      break
    case 'date':
    case 'created_time':
    case 'last_edited_time':
    case 'rich_text':
    case 'number':
    case 'select':
    case 'status':
    case 'title':
      filter = {
        property: name,
        [type]: {
          [field]: value,
        },
      }
      break
    case 'multi_select':
    case 'relation':
      const values = value as string[]
      if (values.length == 1) {
        filter = {
          property: name,
          [type]: {
            [field]: value[0],
          },
        }
      } else {
        filter = { and: [] }
        for (const v of values) {
          filter.and.push({
            property: name,
            [type]: {
              [field]: v,
            },
          })
        }
      }
      break

    case 'files':
    case 'formula':
    case 'people':
    case 'rollup':
    default:
      console.error(`type: ${type} is not support type`)
  }
  return filter
}

export const buildPagePropUpdateData = async (
  name: string,
  type: string,
  value: string
): Promise<object | null> => {
  switch (type) {
    case 'number':
      return {
        [name]: {
          [type]: value,
        },
      }
    case 'select':
      return {
        [name]: {
          [type]: {
            name: value,
          },
        },
      }
    case 'multi_select':
      const nameObjects = []
      for (const val of value) {
        nameObjects.push({
          name: val,
        })
      }
      return {
        [name]: {
          [type]: nameObjects,
        },
      }
    case 'relation':
      const relationPageIds = []
      for (const id of value) {
        relationPageIds.push({ id: id })
      }
      return {
        [name]: {
          [type]: relationPageIds,
        },
      }
  }
  return null
}

export const buildOneDepthJson = async (pages: QueryDataSourceResponse['results']) => {
  const oneDepthJson = []
  const relationJson = []
  for (const page of pages) {
    if (page.object != 'page') {
      continue
    }
    if (!isFullPage(page)) {
      continue
    }
    const pageData = {}
    pageData['page_id'] = page.id
    Object.entries(page.properties).forEach(([key, prop]) => {
      switch (prop.type) {
        case 'number':
          pageData[key] = prop.number
          break
        case 'select':
          pageData[key] = prop.select === null ? '' : prop.select.name
          break
        case 'multi_select':
          const multiSelects = []
          for (const select of prop.multi_select) {
            multiSelects.push(select.name)
          }
          pageData[key] = multiSelects.join(',')
          break
        case 'relation':
          const relationPages = []
          // relationJsonにkeyがなければ作成
          if (relationJson[key] == null) {
            relationJson[key] = []
          }
          for (const relation of prop.relation) {
            relationPages.push(relation.id)
            relationJson[key].push({
              page_id: page.id,
              relation_page_id: relation.id,
            })
          }
          pageData[key] = relationPages.join(',')
          break
        case 'created_time':
          pageData[key] = prop.created_time
          break
        case 'last_edited_time':
          pageData[key] = prop.last_edited_time
          break
        case 'formula':
          switch (prop.formula.type) {
            case 'string':
              pageData[key] = prop.formula.string
              break
            case 'number':
              pageData[key] = prop.formula.number
              break
            case 'boolean':
              pageData[key] = prop.formula.boolean
              break
            case 'date':
              pageData[key] = prop.formula.date.start
              break
            default:
            // console.error(`${prop.formula.type} is not supported`)
          }
          break
        case 'url':
          pageData[key] = prop.url
          break
        case 'date':
          pageData[key] = prop.date === null ? '' : prop.date.start
          break
        case 'email':
          pageData[key] = prop.email
          break
        case 'phone_number':
          pageData[key] = prop.phone_number
          break
        case 'created_by':
          pageData[key] = prop.created_by.id
          break
        case 'last_edited_by':
          pageData[key] = prop.last_edited_by.id
          break
        case 'people':
          const people = []
          for (const person of prop.people) {
            people.push(person.id)
          }
          pageData[key] = people.join(',')
          break
        case 'files':
          const files = []
          for (const file of prop.files) {
            files.push(file.name)
          }
          pageData[key] = files.join(',')
          break
        case 'checkbox':
          pageData[key] = prop.checkbox
          break
        // @ts-ignore
        case 'unique_id':
          // @ts-ignore
          pageData[key] = `${prop.unique_id.prefix}-${prop.unique_id.number}`
          break
        case 'title':
          pageData[key] = prop.title[0].plain_text
          break
        case 'rich_text':
          const richTexts = []
          for (const richText of prop.rich_text) {
            richTexts.push(richText.plain_text)
          }
          pageData[key] = richTexts.join(',')
          break
        case 'status':
          pageData[key] = prop.status === null ? '' : prop.status.name
          break
        default:
          console.error(`${key}(type: ${prop.type}) is not supported`)
      }
    })
    oneDepthJson.push(pageData)
  }

  return { oneDepthJson, relationJson }
}


export const getDbTitle = (row: DatabaseObjectResponse) => {
  if (row.title && row.title.length > 0) {
    return row.title[0].plain_text
  }
  return 'Untitled'
}

export const getDataSourceTitle = (row: GetDataSourceResponse | DataSourceObjectResponse) => {
  // Check if it's a full data source response
  if (isFullDataSource(row)) {
    if (row.title && row.title.length > 0) {
      return row.title[0].plain_text
    }
  }
  return 'Untitled'
}

export const getPageTitle = (row: PageObjectResponse) => {
  let title = 'Untitled'
  Object.entries(row.properties).find(([_, prop]) => {
    if (prop.type === 'title' && prop.title.length > 0) {
      title = prop.title[0].plain_text
      return true
    }
  })
  return title
}

export const getBlockPlainText = (row: BlockObjectResponse) => {
  try {
    switch (row.type) {
      case 'bookmark':
        return row[row.type].url
      case 'breadcrumb':
        return ''
      case 'child_database':
        return row[row.type].title
      case 'child_page':
        return row[row.type].title
      case 'column_list':
        return ''
      case 'divider':
        return ''
      case 'embed':
        return row[row.type].url
      case 'equation':
        return row[row.type].expression
      case 'file':
      case 'image':
        if (row[row.type].type == 'file') {
          return row[row.type].file.url
        } else {
          return row[row.type].external.url
        }
      case 'link_preview':
        return row[row.type].url
      case 'synced_block':
        return ''
      case 'table_of_contents':
        return ''
      case 'table':
        return ''

      case 'bulleted_list_item':
      case 'callout':
      case 'code':
      case 'heading_1':
      case 'heading_2':
      case 'heading_3':
      case 'numbered_list_item':
      case 'paragraph':
      case 'quote':
      case 'to_do':
      case 'toggle':
        let plainText = ''
        if (row[row.type].rich_text.length > 0) {
          plainText = row[row.type].rich_text[0].plain_text
        }
        return plainText

      default:
        return row[row.type]
    }
  } catch (e) {
    console.error(`${row.type} is not supported`)
    console.error(e)
    return ''
  }
}

/**
 * Helper to create rich text array from plain text string
 */
const createRichText = (text: string) => {
  return [
    {
      type: 'text' as const,
      text: {
        content: text,
      },
    },
  ]
}

/**
 * Build block JSON from simple text-based flags
 * Returns an array of block objects ready for Notion API
 */
export const buildBlocksFromTextFlags = (flags: {
  text?: string
  heading1?: string
  heading2?: string
  heading3?: string
  bullet?: string
  numbered?: string
  todo?: string
  toggle?: string
  code?: string
  language?: string
  quote?: string
  callout?: string
}): any[] => {
  const blocks: any[] = []

  if (flags.text) {
    blocks.push({
      object: 'block',
      type: 'paragraph',
      paragraph: {
        rich_text: createRichText(flags.text),
      },
    })
  }

  if (flags.heading1) {
    blocks.push({
      object: 'block',
      type: 'heading_1',
      heading_1: {
        rich_text: createRichText(flags.heading1),
      },
    })
  }

  if (flags.heading2) {
    blocks.push({
      object: 'block',
      type: 'heading_2',
      heading_2: {
        rich_text: createRichText(flags.heading2),
      },
    })
  }

  if (flags.heading3) {
    blocks.push({
      object: 'block',
      type: 'heading_3',
      heading_3: {
        rich_text: createRichText(flags.heading3),
      },
    })
  }

  if (flags.bullet) {
    blocks.push({
      object: 'block',
      type: 'bulleted_list_item',
      bulleted_list_item: {
        rich_text: createRichText(flags.bullet),
      },
    })
  }

  if (flags.numbered) {
    blocks.push({
      object: 'block',
      type: 'numbered_list_item',
      numbered_list_item: {
        rich_text: createRichText(flags.numbered),
      },
    })
  }

  if (flags.todo) {
    blocks.push({
      object: 'block',
      type: 'to_do',
      to_do: {
        rich_text: createRichText(flags.todo),
        checked: false,
      },
    })
  }

  if (flags.toggle) {
    blocks.push({
      object: 'block',
      type: 'toggle',
      toggle: {
        rich_text: createRichText(flags.toggle),
      },
    })
  }

  if (flags.code) {
    blocks.push({
      object: 'block',
      type: 'code',
      code: {
        rich_text: createRichText(flags.code),
        language: flags.language || 'plain text',
      },
    })
  }

  if (flags.quote) {
    blocks.push({
      object: 'block',
      type: 'quote',
      quote: {
        rich_text: createRichText(flags.quote),
      },
    })
  }

  if (flags.callout) {
    blocks.push({
      object: 'block',
      type: 'callout',
      callout: {
        rich_text: createRichText(flags.callout),
        icon: {
          type: 'emoji',
          emoji: '💡',
        },
      },
    })
  }

  return blocks
}

/**
 * Build block update content from simple text flags
 * Returns an object with the block type properties for updating
 */
export const buildBlockUpdateFromTextFlags = (
  blockType: string,
  flags: {
    text?: string
    heading1?: string
    heading2?: string
    heading3?: string
    bullet?: string
    numbered?: string
    todo?: string
    toggle?: string
    code?: string
    language?: string
    quote?: string
    callout?: string
  }
): any => {
  // For updates, we need to know the block type and provide the appropriate content
  // The text flags can update any compatible block type

  if (flags.text) {
    return {
      paragraph: {
        rich_text: createRichText(flags.text),
      },
    }
  }

  if (flags.heading1) {
    return {
      heading_1: {
        rich_text: createRichText(flags.heading1),
      },
    }
  }

  if (flags.heading2) {
    return {
      heading_2: {
        rich_text: createRichText(flags.heading2),
      },
    }
  }

  if (flags.heading3) {
    return {
      heading_3: {
        rich_text: createRichText(flags.heading3),
      },
    }
  }

  if (flags.bullet) {
    return {
      bulleted_list_item: {
        rich_text: createRichText(flags.bullet),
      },
    }
  }

  if (flags.numbered) {
    return {
      numbered_list_item: {
        rich_text: createRichText(flags.numbered),
      },
    }
  }

  if (flags.todo) {
    return {
      to_do: {
        rich_text: createRichText(flags.todo),
      },
    }
  }

  if (flags.toggle) {
    return {
      toggle: {
        rich_text: createRichText(flags.toggle),
      },
    }
  }

  if (flags.code) {
    return {
      code: {
        rich_text: createRichText(flags.code),
        language: flags.language || 'plain text',
      },
    }
  }

  if (flags.quote) {
    return {
      quote: {
        rich_text: createRichText(flags.quote),
      },
    }
  }

  if (flags.callout) {
    return {
      callout: {
        rich_text: createRichText(flags.callout),
      },
    }
  }

  return null
}
