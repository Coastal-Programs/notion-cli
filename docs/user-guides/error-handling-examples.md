# Error Handling Integration Examples

**Version**: 1.0.0
**Last Updated**: 2025-10-22

Real-world examples of integrating the enhanced error handling system into Notion CLI commands.

---

## Table of Contents

1. [Basic Command Integration](#basic-command-integration)
2. [Database Query Command](#database-query-command)
3. [Page Create Command](#page-create-command)
4. [Block Update Command](#block-update-command)
5. [Custom Validation](#custom-validation)
6. [Retry Logic Integration](#retry-logic-integration)
7. [Testing Examples](#testing-examples)

---

## Basic Command Integration

### Minimal Example

```typescript
import { Args, Command, Flags } from '@oclif/core'
import { handleCliError, ErrorContext } from '../errors'
import * as notion from '../notion'

export default class DbRetrieve extends Command {
  static description = 'Retrieve a database'

  static args = {
    database_id: Args.string({ required: true })
  }

  static flags = {
    json: Flags.boolean({ char: 'j', description: 'Output as JSON' })
  }

  async run() {
    const { args, flags } = await this.parse(DbRetrieve)

    try {
      // Your command logic
      const database = await notion.retrieveDb(args.database_id)

      // Success output
      if (flags.json) {
        this.log(JSON.stringify({ success: true, data: database }))
      } else {
        console.log(`Database: ${database.title}`)
      }

    } catch (error) {
      // Context helps generate better error messages
      const context: ErrorContext = {
        resourceType: 'database',
        attemptedId: args.database_id,
        userInput: args.database_id,
        endpoint: 'databases.retrieve'
      }

      // handleCliError never returns - exits with code 1
      handleCliError(error, flags.json, context)
    }
  }
}
```

---

## Database Query Command

### Complete Example with Validation

```typescript
import { Args, Command, Flags, ux } from '@oclif/core'
import {
  handleCliError,
  NotionCLIErrorFactory,
  ErrorContext,
  NotionCLIError
} from '../errors'
import * as notion from '../notion'
import { resolveNotionId } from '../utils/notion-resolver'
import * as fs from 'fs'
import * as path from 'path'

export default class DbQuery extends Command {
  static description = 'Query a database with filters and sorting'

  static args = {
    database_id: Args.string({
      required: true,
      description: 'Database ID, URL, or name'
    })
  }

  static flags = {
    json: Flags.boolean({
      char: 'j',
      description: 'Output as JSON'
    }),
    rawFilter: Flags.string({
      char: 'a',
      description: 'JSON stringified filter'
    }),
    fileFilter: Flags.string({
      char: 'f',
      description: 'Path to JSON filter file'
    }),
    sortProperty: Flags.string({
      char: 's',
      description: 'Property name to sort by'
    }),
    sortDirection: Flags.string({
      char: 'd',
      options: ['asc', 'desc'],
      default: 'asc',
      description: 'Sort direction'
    }),
    pageSize: Flags.integer({
      char: 'p',
      min: 1,
      max: 100,
      default: 10,
      description: 'Number of results per page'
    })
  }

  async run() {
    const { args, flags } = await this.parse(DbQuery)

    try {
      // Step 1: Resolve database ID (handles URLs, names, IDs)
      const databaseId = await resolveNotionId(args.database_id, 'database')

      // Step 2: Build query parameters with validation
      const queryParams = await this.buildQueryParams(databaseId, flags)

      // Step 3: Execute query
      const response = await notion.client.dataSources.query(queryParams)

      // Step 4: Output results
      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          data: response.results,
          count: response.results.length,
          hasMore: response.has_more
        }))
      } else {
        // Table output
        ux.table(response.results, {
          id: {},
          title: { get: (row) => this.getTitle(row) },
          url: {}
        })
      }

    } catch (error) {
      const context: ErrorContext = {
        resourceType: 'database',
        attemptedId: args.database_id,
        userInput: args.database_id,
        endpoint: 'dataSources.query',
        metadata: {
          filterProvided: !!(flags.rawFilter || flags.fileFilter),
          sortProvided: !!flags.sortProperty
        }
      }

      handleCliError(error, flags.json, context)
    }
  }

  /**
   * Build query parameters with proper validation
   */
  private async buildQueryParams(databaseId: string, flags: any) {
    const params: any = {
      data_source_id: databaseId,
      page_size: flags.pageSize
    }

    // Handle filter (mutually exclusive: raw or file)
    if (flags.rawFilter) {
      params.filter = this.parseFilterJson(flags.rawFilter, 'inline')
    } else if (flags.fileFilter) {
      params.filter = this.parseFilterFromFile(flags.fileFilter)
    }

    // Handle sorting with property validation
    if (flags.sortProperty) {
      await this.validateProperty(databaseId, flags.sortProperty)

      params.sorts = [{
        property: flags.sortProperty,
        direction: flags.sortDirection === 'desc' ? 'descending' : 'ascending'
      }]
    }

    return params
  }

  /**
   * Parse JSON filter string with helpful error on failure
   */
  private parseFilterJson(jsonString: string, source: 'inline' | 'file'): object {
    try {
      return JSON.parse(jsonString)
    } catch (parseError) {
      throw NotionCLIErrorFactory.invalidJson(jsonString, parseError as Error)
    }
  }

  /**
   * Read and parse filter from file
   */
  private parseFilterFromFile(filePath: string): object {
    try {
      const absolutePath = path.resolve(filePath)
      const fileContent = fs.readFileSync(absolutePath, 'utf-8')
      return this.parseFilterJson(fileContent, 'file')
    } catch (error: any) {
      if (error instanceof NotionCLIError) {
        throw error // Re-throw parsing errors
      }

      // File read error
      throw new NotionCLIError(
        'VALIDATION_ERROR' as any,
        `Failed to read filter file: ${filePath}`,
        [
          {
            description: 'Verify the file path is correct'
          },
          {
            description: 'Check file permissions'
          },
          {
            description: 'Use absolute path or path relative to current directory'
          }
        ],
        {
          userInput: filePath,
          originalError: error,
          metadata: { errorCode: error.code }
        }
      )
    }
  }

  /**
   * Validate that property exists in database schema
   */
  private async validateProperty(databaseId: string, propertyName: string): Promise<void> {
    try {
      const database = await notion.retrieveDb(databaseId)

      if (!database.properties || !database.properties[propertyName]) {
        throw NotionCLIErrorFactory.invalidProperty(propertyName, databaseId)
      }
    } catch (error) {
      // If database retrieval fails, let it bubble up
      // If it's our validation error, re-throw
      throw error
    }
  }

  /**
   * Extract title from page/database object
   */
  private getTitle(row: any): string {
    if (row.properties) {
      // It's a page - find title property
      const titleProp = Object.values(row.properties).find(
        (prop: any) => prop.type === 'title'
      )
      if (titleProp && titleProp.title?.[0]?.plain_text) {
        return titleProp.title[0].plain_text
      }
    }
    return 'Untitled'
  }
}
```

---

## Page Create Command

### Example with Property Validation

```typescript
import { Args, Command, Flags } from '@oclif/core'
import {
  handleCliError,
  NotionCLIErrorFactory,
  ErrorContext
} from '../errors'
import * as notion from '../notion'
import { resolveNotionId } from '../utils/notion-resolver'

export default class PageCreate extends Command {
  static description = 'Create a new page in a database'

  static args = {
    database_id: Args.string({ required: true }),
    title: Args.string({ required: true })
  }

  static flags = {
    json: Flags.boolean({ char: 'j' }),
    properties: Flags.string({
      char: 'p',
      description: 'JSON object of properties to set'
    })
  }

  async run() {
    const { args, flags } = await this.parse(PageCreate)

    try {
      // Resolve database
      const databaseId = await resolveNotionId(args.database_id, 'database')

      // Build page properties
      const properties = this.buildProperties(args.title, flags.properties)

      // Validate properties against schema
      await this.validateProperties(databaseId, properties)

      // Create page
      const page = await notion.createPage({
        parent: { database_id: databaseId },
        properties
      })

      // Success output
      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          data: { id: page.id, url: page.url }
        }))
      } else {
        console.log(`✓ Page created: ${page.url}`)
      }

    } catch (error) {
      const context: ErrorContext = {
        resourceType: 'database',
        attemptedId: args.database_id,
        userInput: args.database_id,
        endpoint: 'pages.create',
        metadata: {
          title: args.title,
          propertiesProvided: !!flags.properties
        }
      }

      handleCliError(error, flags.json, context)
    }
  }

  /**
   * Build page properties from inputs
   */
  private buildProperties(title: string, propertiesJson?: string): Record<string, any> {
    const properties: Record<string, any> = {
      // Find and set title property (we'll validate schema has one)
      title: {
        title: [{ text: { content: title } }]
      }
    }

    if (propertiesJson) {
      try {
        const additionalProps = JSON.parse(propertiesJson)
        Object.assign(properties, additionalProps)
      } catch (parseError) {
        throw NotionCLIErrorFactory.invalidJson(propertiesJson, parseError as Error)
      }
    }

    return properties
  }

  /**
   * Validate properties against database schema
   */
  private async validateProperties(
    databaseId: string,
    properties: Record<string, any>
  ): Promise<void> {
    const database = await notion.retrieveDb(databaseId)
    const schema = database.properties

    // Check each property exists in schema
    for (const [propName, propValue] of Object.entries(properties)) {
      // Skip 'title' - it's handled specially
      if (propName === 'title') continue

      if (!schema[propName]) {
        throw NotionCLIErrorFactory.invalidProperty(propName, databaseId)
      }

      // Could add type validation here
      // e.g., if trying to set number property to string
    }
  }
}
```

---

## Block Update Command

### Example with Nested Resource Handling

```typescript
import { Args, Command, Flags } from '@oclif/core'
import {
  handleCliError,
  NotionCLIErrorFactory,
  ErrorContext,
  NotionCLIError
} from '../errors'
import * as notion from '../notion'

export default class BlockUpdate extends Command {
  static description = 'Update a block'

  static args = {
    block_id: Args.string({ required: true }),
    content: Args.string({ required: true })
  }

  static flags = {
    json: Flags.boolean({ char: 'j' }),
    type: Flags.string({
      char: 't',
      options: ['paragraph', 'heading_1', 'heading_2', 'heading_3', 'bulleted_list_item'],
      default: 'paragraph',
      description: 'Block type'
    })
  }

  async run() {
    const { args, flags } = await this.parse(BlockUpdate)

    try {
      // Validate block ID format
      const blockId = this.validateBlockId(args.block_id)

      // Verify block exists and is accessible
      await this.verifyBlockAccess(blockId)

      // Build update parameters
      const updateParams = this.buildUpdateParams(blockId, args.content, flags.type)

      // Update block
      const block = await notion.updateBlock(updateParams)

      // Success output
      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          data: { id: block.id, type: block.type }
        }))
      } else {
        console.log(`✓ Block updated: ${block.id}`)
      }

    } catch (error) {
      const context: ErrorContext = {
        resourceType: 'block',
        attemptedId: args.block_id,
        userInput: args.block_id,
        endpoint: 'blocks.update',
        metadata: {
          blockType: flags.type,
          contentLength: args.content.length
        }
      }

      handleCliError(error, flags.json, context)
    }
  }

  /**
   * Validate block ID format
   */
  private validateBlockId(input: string): string {
    const cleaned = input.replace(/-/g, '')

    if (!/^[a-f0-9]{32}$/i.test(cleaned)) {
      throw NotionCLIErrorFactory.invalidIdFormat(input, 'block')
    }

    return cleaned
  }

  /**
   * Verify block exists and is accessible
   */
  private async verifyBlockAccess(blockId: string): Promise<void> {
    try {
      await notion.retrieveBlock(blockId)
    } catch (error: any) {
      // Enhance 404 errors with block-specific context
      if (error.status === 404 || error.code === 'object_not_found') {
        throw NotionCLIErrorFactory.resourceNotFound('block', blockId)
      }

      // Re-throw other errors
      throw error
    }
  }

  /**
   * Build update parameters for block
   */
  private buildUpdateParams(blockId: string, content: string, type: string): any {
    return {
      block_id: blockId,
      [type]: {
        rich_text: [
          {
            text: { content }
          }
        ]
      }
    }
  }
}
```

---

## Custom Validation

### Reusable Validation Utilities

```typescript
import {
  NotionCLIError,
  NotionCLIErrorFactory,
  NotionCLIErrorCode
} from '../errors'

/**
 * Validation utilities for common scenarios
 */
export class ValidationUtils {

  /**
   * Validate Notion ID format (32 hex chars)
   */
  static validateNotionId(
    input: string,
    type: 'database' | 'page' | 'block' = 'database'
  ): string {
    const cleaned = input.replace(/-/g, '')

    if (!/^[a-f0-9]{32}$/i.test(cleaned)) {
      throw NotionCLIErrorFactory.invalidIdFormat(input, type)
    }

    return cleaned
  }

  /**
   * Validate and parse JSON input
   */
  static parseJson<T = any>(jsonString: string, fieldName: string = 'JSON input'): T {
    try {
      return JSON.parse(jsonString) as T
    } catch (parseError) {
      throw NotionCLIErrorFactory.invalidJson(jsonString, parseError as Error)
    }
  }

  /**
   * Validate property exists in database schema
   */
  static async validateProperty(
    database: any,
    propertyName: string,
    databaseId: string
  ): Promise<void> {
    if (!database.properties || !database.properties[propertyName]) {
      throw NotionCLIErrorFactory.invalidProperty(propertyName, databaseId)
    }
  }

  /**
   * Validate property type matches expected type
   */
  static validatePropertyType(
    database: any,
    propertyName: string,
    expectedType: string,
    databaseId: string
  ): void {
    const property = database.properties?.[propertyName]

    if (!property) {
      throw NotionCLIErrorFactory.invalidProperty(propertyName, databaseId)
    }

    if (property.type !== expectedType) {
      throw new NotionCLIError(
        NotionCLIErrorCode.VALIDATION_ERROR,
        `Property "${propertyName}" is type "${property.type}", expected "${expectedType}"`,
        [
          {
            description: 'Get the database schema to see property types',
            command: `notion-cli db schema ${databaseId}`
          },
          {
            description: `Available properties of type "${expectedType}":`,
          }
        ],
        {
          resourceType: 'database',
          attemptedId: databaseId,
          metadata: {
            propertyName,
            actualType: property.type,
            expectedType
          }
        }
      )
    }
  }

  /**
   * Validate required flags are provided
   */
  static validateRequiredFlags(
    flags: Record<string, any>,
    required: string[]
  ): void {
    const missing = required.filter(flag => !flags[flag])

    if (missing.length > 0) {
      throw new NotionCLIError(
        NotionCLIErrorCode.MISSING_REQUIRED_FIELD,
        `Missing required flags: ${missing.join(', ')}`,
        [
          {
            description: 'Provide all required flags',
            command: `--${missing.join(' --')}`
          }
        ],
        {
          metadata: { missingFlags: missing }
        }
      )
    }
  }

  /**
   * Validate mutually exclusive flags
   */
  static validateMutuallyExclusive(
    flags: Record<string, any>,
    exclusive: string[][]
  ): void {
    for (const group of exclusive) {
      const provided = group.filter(flag => flags[flag])

      if (provided.length > 1) {
        throw new NotionCLIError(
          NotionCLIErrorCode.VALIDATION_ERROR,
          `Flags ${provided.join(', ')} are mutually exclusive`,
          [
            {
              description: 'Use only one of these flags at a time'
            }
          ],
          {
            metadata: { conflictingFlags: provided }
          }
        )
      }
    }
  }

  /**
   * Validate URL format
   */
  static validateUrl(input: string): void {
    try {
      new URL(input)
    } catch {
      throw new NotionCLIError(
        NotionCLIErrorCode.INVALID_URL,
        `Invalid URL format: ${input}`,
        [
          {
            description: 'Ensure the URL is properly formatted',
          },
          {
            description: 'Example: https://notion.so/page-id',
          }
        ],
        {
          userInput: input
        }
      )
    }
  }
}
```

### Usage in Commands

```typescript
import { ValidationUtils } from '../utils/validation'

export default class MyCommand extends Command {
  async run() {
    const { args, flags } = await this.parse(MyCommand)

    try {
      // Validate ID
      const cleanId = ValidationUtils.validateNotionId(args.id, 'database')

      // Validate JSON
      const filter = ValidationUtils.parseJson(flags.filter, 'filter')

      // Validate mutually exclusive flags
      ValidationUtils.validateMutuallyExclusive(flags, [
        ['json', 'csv'],
        ['raw', 'pretty']
      ])

      // ... rest of command
    } catch (error) {
      handleCliError(error, flags.json)
    }
  }
}
```

---

## Retry Logic Integration

### Combining Retry with Enhanced Errors

```typescript
import { fetchWithRetry } from '../retry'
import { wrapNotionError, ErrorContext } from '../errors'

export async function queryDatabaseWithRetry(
  databaseId: string,
  params: any
): Promise<any> {
  const context: ErrorContext = {
    resourceType: 'database',
    attemptedId: databaseId,
    endpoint: 'dataSources.query'
  }

  try {
    return await fetchWithRetry(
      () => client.dataSources.query({
        data_source_id: databaseId,
        ...params
      }),
      {
        config: {
          maxRetries: 3,
          baseDelay: 1000
        },
        context: `query-database-${databaseId}`,
        onRetry: (retryContext) => {
          // Custom retry logging
          console.log(
            `Retrying query (${retryContext.attempt}/${retryContext.maxRetries})...`
          )
        }
      }
    )
  } catch (error) {
    // Wrap error with context before re-throwing
    throw wrapNotionError(error, context)
  }
}
```

---

## Testing Examples

### Unit Test Template

```typescript
import { describe, it, expect } from 'mocha'
import {
  NotionCLIError,
  NotionCLIErrorCode,
  NotionCLIErrorFactory
} from '../../src/errors'

describe('Error Handling - DbQuery Command', () => {
  describe('Invalid ID Format', () => {
    it('throws INVALID_ID_FORMAT for non-hex characters', () => {
      expect(() => {
        ValidationUtils.validateNotionId('not-valid-id', 'database')
      }).to.throw(NotionCLIError)
        .with.property('code', NotionCLIErrorCode.INVALID_ID_FORMAT)
    })

    it('includes helpful suggestions in error', () => {
      try {
        ValidationUtils.validateNotionId('123', 'database')
      } catch (error: any) {
        expect(error.suggestions).to.have.length.greaterThan(0)
        expect(error.suggestions[0].description).to.include('32 hexadecimal')
      }
    })
  })

  describe('Invalid JSON Filter', () => {
    it('throws INVALID_JSON for malformed JSON', () => {
      expect(() => {
        ValidationUtils.parseJson('{invalid}', 'filter')
      }).to.throw(NotionCLIError)
        .with.property('code', NotionCLIErrorCode.INVALID_JSON)
    })

    it('includes JSON validator link in suggestions', () => {
      try {
        ValidationUtils.parseJson('{invalid}', 'filter')
      } catch (error: any) {
        const hasValidatorLink = error.suggestions.some(
          (s: any) => s.link?.includes('jsonlint')
        )
        expect(hasValidatorLink).to.be.true
      }
    })
  })

  describe('Property Validation', () => {
    it('throws INVALID_PROPERTY for non-existent property', async () => {
      const mockDb = {
        properties: {
          Name: { type: 'title' },
          Status: { type: 'select' }
        }
      }

      expect(() => {
        ValidationUtils.validateProperty(mockDb, 'NonExistent', 'db-id')
      }).to.throw(NotionCLIError)
        .with.property('code', NotionCLIErrorCode.INVALID_PROPERTY)
    })
  })
})
```

### Integration Test Template

```typescript
import { describe, it, expect, before, after } from 'mocha'
import { exec } from 'child_process'
import { promisify } from 'util'

const execAsync = promisify(exec)

describe('CLI Error Integration Tests', () => {
  let originalToken: string | undefined

  before(() => {
    originalToken = process.env.NOTION_TOKEN
  })

  after(() => {
    if (originalToken) {
      process.env.NOTION_TOKEN = originalToken
    }
  })

  describe('Token Validation', () => {
    it('returns TOKEN_MISSING when token not set', async () => {
      delete process.env.NOTION_TOKEN

      try {
        await execAsync('notion-cli db query test-id')
        expect.fail('Should have thrown error')
      } catch (error: any) {
        expect(error.stdout).to.include('TOKEN_MISSING')
        expect(error.stdout).to.include('notion-cli config set-token')
        expect(error.code).to.equal(1)
      }
    })

    it('returns JSON error when --json flag used', async () => {
      delete process.env.NOTION_TOKEN

      try {
        await execAsync('notion-cli db query test-id --json')
        expect.fail('Should have thrown error')
      } catch (error: any) {
        const output = JSON.parse(error.stdout)
        expect(output).to.have.property('success', false)
        expect(output.error).to.have.property('code', 'TOKEN_MISSING')
        expect(output.error.suggestions).to.be.an('array')
      }
    })
  })

  describe('ID Format Validation', () => {
    it('returns INVALID_ID_FORMAT for bad ID', async () => {
      process.env.NOTION_TOKEN = 'test-token'

      try {
        await execAsync('notion-cli db query "not-a-valid-id"')
        expect.fail('Should have thrown error')
      } catch (error: any) {
        expect(error.stdout).to.include('INVALID_ID_FORMAT')
        expect(error.stdout).to.include('32 hexadecimal characters')
      }
    })
  })
})
```

---

## Summary

These examples demonstrate:

1. **Consistent Error Handling** - All commands use the same pattern
2. **Rich Context** - Every error includes relevant metadata
3. **Validation First** - Catch errors before API calls when possible
4. **Helpful Messages** - Users get actionable suggestions
5. **Testable** - Error handling is easy to unit test
6. **JSON Support** - All errors work in automation mode

**Next Steps:**
1. Review examples for your use case
2. Copy relevant patterns to your command
3. Add tests for error scenarios
4. Test both human and JSON output

---

**Document Version**: 1.0.0
**Last Updated**: 2025-10-22
**See Also**: [Architecture](./ERROR-HANDLING-ARCHITECTURE.md) | [Quick Reference](./ERROR-HANDLING-QUICK-REF.md)
