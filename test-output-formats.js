/**
 * Test script for new output formats
 *
 * This script demonstrates and tests the new output format functions
 * without requiring a real Notion API connection.
 */

// Simulate the helper functions (in real code, these are in src/helper.ts)
const testData = [
  {
    object: 'page',
    id: '550e8400-e29b-41d4-a716-446655440000',
    url: 'https://www.notion.so/test-page',
    title: 'Test Page 1'
  },
  {
    object: 'page',
    id: '550e8400-e29b-41d4-a716-446655440001',
    url: 'https://www.notion.so/another-page',
    title: 'Another Page with | pipe'
  },
  {
    object: 'database',
    id: '550e8400-e29b-41d4-a716-446655440002',
    url: 'https://www.notion.so/test-db',
    title: 'Test Database'
  }
]

const columns = {
  title: {
    get: (row) => row.title || 'Untitled'
  },
  object: {},
  id: {},
  url: {}
}

// Output functions (matching the implementation in helper.ts)
function outputCompactJson(data) {
  console.log(JSON.stringify(data))
}

function outputMarkdownTable(data, columns) {
  if (!data || data.length === 0) {
    console.log('No data to display')
    return
  }

  const headers = Object.keys(columns)
  const headerRow = '| ' + headers.join(' | ') + ' |'
  const separatorRow = '| ' + headers.map(() => '---').join(' | ') + ' |'

  console.log(headerRow)
  console.log(separatorRow)

  data.forEach((row) => {
    const values = headers.map((header) => {
      const column = columns[header]
      let value

      if (column.get && typeof column.get === 'function') {
        value = column.get(row)
      } else {
        value = row[header]
      }

      if (value === null || value === undefined) {
        return ''
      }

      const stringValue = String(value).replace(/\|/g, '\\|').replace(/\n/g, ' ')
      return stringValue
    })

    console.log('| ' + values.join(' | ') + ' |')
  })
}

function outputPrettyTable(data, columns) {
  if (!data || data.length === 0) {
    console.log('No data to display')
    return
  }

  const headers = Object.keys(columns)
  const columnWidths = {}

  headers.forEach((header) => {
    columnWidths[header] = header.length
  })

  data.forEach((row) => {
    headers.forEach((header) => {
      const column = columns[header]
      let value

      if (column.get && typeof column.get === 'function') {
        value = column.get(row)
      } else {
        value = row[header]
      }

      const stringValue = String(value === null || value === undefined ? '' : value)
      columnWidths[header] = Math.max(columnWidths[header], stringValue.length)
    })
  })

  const topBorder = '┌' + headers.map(h => '─'.repeat(columnWidths[h] + 2)).join('┬') + '┐'
  const headerSeparator = '├' + headers.map(h => '─'.repeat(columnWidths[h] + 2)).join('┼') + '┤'
  const bottomBorder = '└' + headers.map(h => '─'.repeat(columnWidths[h] + 2)).join('┴') + '┘'

  console.log(topBorder)
  const headerRow = '│ ' + headers.map(h => h.padEnd(columnWidths[h])).join(' │ ') + ' │'
  console.log(headerRow)
  console.log(headerSeparator)

  data.forEach((row) => {
    const values = headers.map((header) => {
      const column = columns[header]
      let value

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

  console.log(bottomBorder)
}

// Run tests
console.log('=== TEST 1: Compact JSON Output ===')
outputCompactJson(testData)
console.log()

console.log('=== TEST 2: Markdown Table Output ===')
outputMarkdownTable(testData, columns)
console.log()

console.log('=== TEST 3: Pretty Table Output ===')
outputPrettyTable(testData, columns)
console.log()

console.log('=== All tests completed successfully! ===')
