import { expect, test } from '@oclif/test'
import * as nock from 'nock'
import * as sinon from 'sinon'
import { cacheManager } from '../../../dist/cache.js'
import { diskCacheManager } from '../../../dist/utils/disk-cache.js'

const DATABASE_ID = '11111111-2222-3333-4444-555555555555'
const DATABASE_ID_NO_DASHES = DATABASE_ID.replace(/-/g, '')
const DATABASE_ID_TEST3 = '11111111-2222-3333-4444-666666666666'
const DATABASE_ID_TEST3_NO_DASHES = DATABASE_ID_TEST3.replace(/-/g, '')

const response = {
  object: 'data_source',
  id: DATABASE_ID,
  title: [
    {
      type: 'text',
      text: {
        content: 'dummy database title',
      },
      plain_text: 'dummy database title',
    },
  ],
  url: `https://www.notion.so/${DATABASE_ID_NO_DASHES}`,
}

const titleEmptyResponse = {
  object: 'data_source',
  id: DATABASE_ID_TEST3,
  title: [],
  url: `https://www.notion.so/${DATABASE_ID_TEST3_NO_DASHES}`,
}

describe('db:retrieve', () => {
  let processExitStub: sinon.SinonStub
  let originalDiskCacheEnabled: string | undefined

  before(() => {
    // Disable disk cache for all tests in this suite
    originalDiskCacheEnabled = process.env.NOTION_CLI_DISK_CACHE_ENABLED
    process.env.NOTION_CLI_DISK_CACHE_ENABLED = 'false'
  })

  after(() => {
    // Restore disk cache setting
    if (originalDiskCacheEnabled === undefined) {
      delete process.env.NOTION_CLI_DISK_CACHE_ENABLED
    } else {
      process.env.NOTION_CLI_DISK_CACHE_ENABLED = originalDiskCacheEnabled
    }
  })

  beforeEach(async () => {
    // Clean all nock mocks and abort any pending requests
    nock.abortPendingRequests()
    nock.cleanAll()
    nock.restore()
    nock.activate()
    // Clear both in-memory and disk cache
    cacheManager.clear()
    await diskCacheManager.clear()
    // Disable cache by directly modifying the config
    ;(cacheManager as any).config.enabled = false
    // Stub process.exit to prevent tests from hanging
    processExitStub = sinon.stub(process, 'exit' as any)
  })

  afterEach(async () => {
    nock.cleanAll()
    processExitStub.restore()
    // Clear disk cache again
    await diskCacheManager.clear()
    // Re-enable cache for other tests
    ;(cacheManager as any).config.enabled = true
  })

  describe('with no flags', () => {
    test
      .do(() => {
        // Mock the data_sources endpoint
        // NOTE: Only called once due to caching. First call from resolveNotionId caches the result,
        // second call from retrieve uses cache instead of API.
        nock('https://api.notion.com')
          .get(`/v1/data_sources/${DATABASE_ID_NO_DASHES}`)
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['db:retrieve', '--no-truncate', DATABASE_ID])
      .it('shows retrieved result table', (ctx) => {
        expect(ctx.stdout).to.match(/title.*object.*id.*url/)
        expect(ctx.stdout).to.match(
          new RegExp(`dummy database title.*data_source.*${DATABASE_ID}.*https://www\\.notion\\.so/${DATABASE_ID_NO_DASHES}`)
        )
      })
  })

  describe('with --raw flags', () => {
    test
      .do(() => {
        // Mock the data_sources endpoint
        // NOTE: Only called once due to caching (same as first test)
        nock('https://api.notion.com')
          .get(`/v1/data_sources/${DATABASE_ID_NO_DASHES}`)
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['db:retrieve', DATABASE_ID, '--raw'])
      .it('shows a database object', (ctx) => {
        expect(ctx.stdout).to.contain(DATABASE_ID)
        expect(ctx.stdout).to.contain('dummy database title')
      })
  })

  describe('response title is []', () => {
    test
      .do(() => {
        // CRITICAL: Clean nock and cache before setting up mocks to prevent cross-test pollution
        nock.cleanAll()
        cacheManager.clear()
        // Mock the data_sources endpoint with a DIFFERENT ID to avoid test interference
        // NOTE: Only called once due to caching (same pattern as other tests)
        nock('https://api.notion.com')
          .get(`/v1/data_sources/${DATABASE_ID_TEST3_NO_DASHES}`)
          .reply(200, titleEmptyResponse)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['db:retrieve', '--no-truncate', DATABASE_ID_TEST3])
      .it('shows retrieved result table', (ctx) => {
        expect(ctx.stdout).to.match(/title.*object.*id.*url/)
        expect(ctx.stdout).to.match(
          new RegExp(`Untitled.*data_source.*${DATABASE_ID_TEST3}.*https://www\\.notion\\.so/${DATABASE_ID_TEST3_NO_DASHES}`)
        )
      })
  })
})
