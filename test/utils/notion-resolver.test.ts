import { expect, test } from '@oclif/test'
import { resolveNotionId } from '../../src/utils/notion-resolver'

/**
 * Tests for Smart ID Resolution
 *
 * These tests verify that the resolver can handle:
 * 1. Direct data_source_id (should work immediately)
 * 2. database_id → data_source_id conversion (smart resolution)
 * 3. URLs with database_id (auto-conversion)
 * 4. Invalid IDs (proper error handling)
 */

describe('Smart Database ID Resolution', () => {
  // Note: These tests require a real Notion API token and workspace
  // They are integration tests that verify the actual API behavior

  describe('resolveNotionId with data_source_id', () => {
    test
      .skip() // Skip by default since it requires real credentials
      .it('should accept valid data_source_id directly', async () => {
        // Replace with actual data_source_id from your test workspace
        const dataSourceId = 'your-data-source-id-here'
        const result = await resolveNotionId(dataSourceId, 'database')
        expect(result).to.equal(dataSourceId)
      })
  })

  describe('resolveNotionId with database_id', () => {
    test
      .skip() // Skip by default since it requires real credentials
      .it('should convert database_id to data_source_id', async () => {
        // This test verifies the smart resolution feature
        // When given a database_id (from parent.database_id), it should:
        // 1. Try to retrieve as data_source_id (fail)
        // 2. Search for pages with this database_id
        // 3. Extract data_source_id from page parent
        // 4. Return the correct data_source_id

        const databaseId = 'your-database-id-here'
        const expectedDataSourceId = 'expected-data-source-id-here'

        const result = await resolveNotionId(databaseId, 'database')
        expect(result).to.equal(expectedDataSourceId)
      })
  })

  describe('resolveNotionId error handling', () => {
    test
      .it('should reject invalid input', async () => {
        try {
          await resolveNotionId('', 'database')
          expect.fail('Should have thrown error')
        } catch (error: any) {
          expect(error.message).to.include('Invalid input')
        }
      })

    test
      .it('should reject null input', async () => {
        try {
          await resolveNotionId(null as any, 'database')
          expect.fail('Should have thrown error')
        } catch (error: any) {
          expect(error.message).to.include('Invalid input')
        }
      })
  })

  describe('ID format validation', () => {
    test
      .it('should accept 32 hex characters', async () => {
        const validId = '1fb79d4c71bb8032b722c82305b63a00'
        // This will fail at API call since ID doesn't exist
        // But it should pass format validation
        try {
          await resolveNotionId(validId, 'database')
        } catch (error: any) {
          // Should fail with "not found" not "invalid format"
          expect(error.message).to.not.include('Invalid')
        }
      })

    test
      .it('should accept UUID format with dashes', async () => {
        const validId = '1fb79d4c-71bb-8032-b722-c82305b63a00'
        // This will fail at API call since ID doesn't exist
        // But it should pass format validation
        try {
          await resolveNotionId(validId, 'database')
        } catch (error: any) {
          // Should fail with "not found" not "invalid format"
          expect(error.message).to.not.include('Invalid')
        }
      })
  })
})

/**
 * Manual Testing Guide
 * ====================
 *
 * To manually test the smart ID resolution feature:
 *
 * 1. Set up your NOTION_TOKEN environment variable:
 *    export NOTION_TOKEN="your-integration-token"
 *
 * 2. Get a database_id from a page:
 *    notion-cli page retrieve <PAGE_ID> --raw | jq '.parent.database_id'
 *
 * 3. Try to use that database_id directly (will auto-convert):
 *    notion-cli db retrieve <DATABASE_ID>
 *
 * Expected Output:
 * ----------------
 * Info: Resolved database_id to data_source_id
 *   database_id:    1fb79d4c71bb8032b722c82305b63a00
 *   data_source_id: 2gc80e5d82cc9043c833d93416c74b11
 *
 * Note: Use data_source_id for database operations.
 *       The database_id from parent.database_id won't work directly.
 *
 * [Database information displayed here...]
 *
 * 4. Verify it works with other database commands:
 *    notion-cli db query <DATABASE_ID>
 *    notion-cli db update <DATABASE_ID> --title "New Title"
 *
 * Testing Edge Cases:
 * -------------------
 *
 * 1. Invalid ID format:
 *    notion-cli db retrieve "invalid-id"
 *    Expected: "Invalid input: expected a database name, ID, or URL"
 *
 * 2. Non-existent ID:
 *    notion-cli db retrieve "1fb79d4c71bb8032b722c82305b63a00"
 *    Expected: "Database not found" (after trying smart resolution)
 *
 * 3. URL with database_id:
 *    notion-cli db retrieve "https://notion.so/1fb79d4c71bb8032b722c82305b63a00"
 *    Expected: Auto-conversion if it's a database_id
 *
 * 4. Database name (existing feature):
 *    notion-cli db retrieve "Tasks"
 *    Expected: Works as before (name lookup)
 */
