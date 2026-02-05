/**
 * Table formatting utility to replace oclif v2's ux.table
 * Provides backward-compatible table flags and formatting
 */

import { Flags } from '@oclif/core'
import * as Table from 'cli-table3'

/**
 * Table flags compatible with oclif v2's ux.table.flags()
 */
export const tableFlags = {
  columns: Flags.string({
    description: 'Only show provided columns (comma-separated)',
    exclusive: ['extended'],
  }),
  sort: Flags.string({
    description: 'Property to sort by (prepend with - for descending)',
  }),
  filter: Flags.string({
    description: 'Filter property by substring match',
  }),
  csv: Flags.boolean({
    description: 'Output in CSV format',
    exclusive: ['no-truncate'],
  }),
  extended: Flags.boolean({
    char: 'x',
    description: 'Show extra columns',
  }),
  'no-truncate': Flags.boolean({
    description: 'Do not truncate output to fit screen',
    exclusive: ['csv'],
  }),
  'no-header': Flags.boolean({
    description: 'Hide table header from output',
  }),
}

export interface ColumnOptions<T> {
  header?: string
  get?: (row: T) => any
  extended?: boolean
  minWidth?: number
}

export interface TableOptions {
  columns?: string
  sort?: string
  filter?: string
  csv?: boolean
  extended?: boolean
  'no-truncate'?: boolean
  'no-header'?: boolean
  printLine?: (s: string) => void
}

/**
 * Format and display a table (compatible with oclif v2's ux.table)
 */
export function formatTable<T extends Record<string, any>>(
  data: T[],
  columns: Record<string, ColumnOptions<T>>,
  options: TableOptions = {}
): void {
  if (data.length === 0) {
    return
  }

  const printLine = options.printLine || console.log

  // Filter columns based on options
  let selectedColumns = Object.keys(columns)

  if (options.columns) {
    const requestedCols = options.columns.split(',').map(c => c.trim())
    selectedColumns = selectedColumns.filter(col => requestedCols.includes(col))
  }

  if (!options.extended) {
    selectedColumns = selectedColumns.filter(col => !columns[col].extended)
  }

  // Filter rows
  let filteredData = data
  if (options.filter) {
    const [filterCol, filterVal] = options.filter.split('=')
    if (filterVal) {
      filteredData = data.filter(row => {
        const val = columns[filterCol]?.get ? columns[filterCol].get!(row) : row[filterCol]
        return String(val).includes(filterVal)
      })
    }
  }

  // Sort data
  if (options.sort) {
    const descending = options.sort.startsWith('-')
    const sortCol = descending ? options.sort.slice(1) : options.sort
    filteredData = [...filteredData].sort((a, b) => {
      const aVal = columns[sortCol]?.get ? columns[sortCol].get!(a) : a[sortCol]
      const bVal = columns[sortCol]?.get ? columns[sortCol].get!(b) : b[sortCol]
      const comparison = String(aVal).localeCompare(String(bVal))
      return descending ? -comparison : comparison
    })
  }

  // Output as CSV
  if (options.csv) {
    if (!options['no-header']) {
      const headers = selectedColumns.map(col => columns[col].header || col)
      printLine(headers.join(','))
    }
    filteredData.forEach(row => {
      const values = selectedColumns.map(col => {
        const val = columns[col].get ? columns[col].get(row) : row[col]
        const str = String(val || '')
        // Escape CSV values
        return str.includes(',') || str.includes('"') ? `"${str.replace(/"/g, '""')}"` : str
      })
      printLine(values.join(','))
    })
    return
  }

  // Output as table
  const headers = selectedColumns.map(col => columns[col].header || col)
  const table = new Table({
    head: options['no-header'] ? [] : headers,
    style: {
      head: ['cyan'],
      border: ['gray']
    },
    wordWrap: !options['no-truncate'],
    colWidths: selectedColumns.map(col => {
      if (options['no-truncate']) return undefined
      return columns[col].minWidth || undefined
    }),
  })

  filteredData.forEach(row => {
    const values = selectedColumns.map(col => {
      const val = columns[col].get ? columns[col].get(row) : row[col]
      return String(val !== undefined && val !== null ? val : '')
    })
    table.push(values)
  })

  printLine(table.toString())
}
