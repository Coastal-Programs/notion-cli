import { expect, test } from '@oclif/test'
import * as nock from 'nock'
import * as sinon from 'sinon'
import { cacheManager } from '../../../dist/cache.js'

// Use valid UUID format for IDs
const PAGE_ID = '11111111-2222-3333-4444-555555555555'
const PAGE_ID_NO_DASHES = PAGE_ID.replace(/-/g, '')
const PAGE_ID_EMPTY_TITLE_1 = '11111111-2222-3333-4444-777777777777'
const PAGE_ID_EMPTY_TITLE_1_NO_DASHES = PAGE_ID_EMPTY_TITLE_1.replace(/-/g, '')
const PAGE_ID_EMPTY_TITLE_2 = '11111111-2222-3333-4444-888888888888'
const PAGE_ID_EMPTY_TITLE_2_NO_DASHES = PAGE_ID_EMPTY_TITLE_2.replace(/-/g, '')
const PARENT_PAGE_ID = '22222222-3333-4444-5555-666666666666'
const PARENT_DB_ID = '33333333-4444-5555-6666-777777777777'

const retrieveOnPageResponse = {
  object: 'page',
  id: PAGE_ID,
  parent: {
    type: 'page_id',
    page_id: PARENT_PAGE_ID,
  },
  archived: false,
  properties: {
    title: {
      id: 'title',
      type: 'title',
      title: [
        {
          type: 'text',
          text: {
            content: 'dummy page title',
          },
          plain_text: 'dummy page title',
        },
      ],
    },
  },
  url: `https://www.notion.so/${PAGE_ID_NO_DASHES}`,
}

const retrieveOnPageResponseWithEmptyTitle = {
  object: 'page',
  id: PAGE_ID_EMPTY_TITLE_1,
  parent: {
    type: 'page_id',
    page_id: PARENT_PAGE_ID,
  },
  archived: false,
  properties: {
    title: {
      id: 'title',
      type: 'title',
      title: [],
    },
  },
  url: `https://www.notion.so/${PAGE_ID_EMPTY_TITLE_1_NO_DASHES}`,
}

const retrieveOnDbResponse = {
  object: 'page',
  id: PAGE_ID,
  parent: {
    type: 'database_id',
    database_id: PARENT_DB_ID,
    data_source_id: PARENT_DB_ID,
  },
  archived: false,
  properties: {
    title: {
      id: 'title',
      type: 'title',
      title: [
        {
          type: 'text',
          text: {
            content: 'dummy page title',
          },
          plain_text: 'dummy page title',
        },
      ],
    },
  },
  url: `https://www.notion.so/${PAGE_ID_NO_DASHES}`,
}

const retrieveOnDbResponseWithEmptyTitle = {
  object: 'page',
  id: PAGE_ID_EMPTY_TITLE_2,
  parent: {
    type: 'database_id',
    database_id: PARENT_DB_ID,
    data_source_id: PARENT_DB_ID,
  },
  archived: false,
  properties: {
    title: {
      id: 'title',
      type: 'title',
      title: [],
    },
  },
  url: `https://www.notion.so/${PAGE_ID_EMPTY_TITLE_2_NO_DASHES}`,
}

describe('page:retrieve', () => {
  let processExitStub: sinon.SinonStub

  beforeEach(async () => {
    // Clean all nock mocks and abort any pending requests
    nock.abortPendingRequests()
    nock.cleanAll()
    nock.restore()
    nock.activate()
    // Clear cache to prevent test interference
    cacheManager.clear()
    // Disable cache by directly modifying the config (environment variables don't work after instantiation)
    ;(cacheManager as any).config.enabled = false
    // Stub process.exit to prevent tests from hanging
    processExitStub = sinon.stub(process, 'exit' as any)
  })

  afterEach(() => {
    nock.cleanAll()
    processExitStub.restore()
    // Re-enable cache for other tests
    ;(cacheManager as any).config.enabled = true
  })

  describe('with page_id on a page flags', () => {
    test
      .do(() => {
        nock('https://api.notion.com')
          .get(`/v1/pages/${PAGE_ID_NO_DASHES}`)
          .reply(200, retrieveOnPageResponse)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['page:retrieve', '--no-truncate', PAGE_ID])
      .it('shows retrieve page result table', (ctx) => {
        expect(ctx.stdout).to.match(/title.*object.*id.*url/)
        expect(ctx.stdout).to.match(
          new RegExp(`dummy page title.*page.*${PAGE_ID}.*https://www\\.notion\\.so/${PAGE_ID_NO_DASHES}`)
        )
      })

    describe('with --raw flags', () => {
      test
        .do(() => {
          nock('https://api.notion.com')
            .get(`/v1/pages/${PAGE_ID_NO_DASHES}`)
            .reply(200, retrieveOnPageResponse)
        })
        .stdout({ print: process.env.TEST_DEBUG ? true : false })
        .command(['page:retrieve', PAGE_ID, '--raw'])
        .it('shows a page object', (ctx) => {
          expect(ctx.stdout).to.contain('object": "page')
          expect(ctx.stdout).to.contain(`id": "${PAGE_ID}`)
          expect(ctx.stdout).to.contain(`url": "https://www.notion.so/${PAGE_ID_NO_DASHES}`)
        })
    })

    describe('response title is []', () => {
      test
        .do(() => {
          // CRITICAL: Clear nock and cache before setting up mocks to prevent cross-test pollution
          nock.cleanAll()
          cacheManager.clear()
          // Use unique ID to avoid test interference
          nock('https://api.notion.com')
            .get(`/v1/pages/${PAGE_ID_EMPTY_TITLE_1_NO_DASHES}`)
            .reply(200, retrieveOnPageResponseWithEmptyTitle)
        })
        .stdout({ print: process.env.TEST_DEBUG ? true : false })
        .command(['page:retrieve', '--no-truncate', PAGE_ID_EMPTY_TITLE_1])
        .it('shows retrieve page result table', (ctx) => {
          expect(ctx.stdout).to.match(/title.*object.*id.*url/)
          expect(ctx.stdout).to.match(
            new RegExp(`Untitled.*page.*${PAGE_ID_EMPTY_TITLE_1}.*https://www\\.notion\\.so/${PAGE_ID_EMPTY_TITLE_1_NO_DASHES}`)
          )
        })
    })
  })

  describe('with page_id on a db flags', () => {
    test
      .do(() => {
        nock('https://api.notion.com')
          .get(`/v1/pages/${PAGE_ID_NO_DASHES}`)
          .reply(200, retrieveOnDbResponse)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['page:retrieve', '--no-truncate', PAGE_ID])
      .it('shows retrieve page result table', (ctx) => {
        expect(ctx.stdout).to.match(/title.*object.*id.*url/)
        expect(ctx.stdout).to.match(
          new RegExp(`dummy page title.*page.*${PAGE_ID}.*https://www\\.notion\\.so/${PAGE_ID_NO_DASHES}`)
        )
      })

    describe('with --raw flags', () => {
      test
        .do(() => {
          nock('https://api.notion.com')
            .get(`/v1/pages/${PAGE_ID_NO_DASHES}`)
            .reply(200, retrieveOnDbResponse)
        })
        .stdout({ print: process.env.TEST_DEBUG ? true : false })
        .command(['page:retrieve', PAGE_ID, '--raw'])
        .it('shows a page object', (ctx) => {
          expect(ctx.stdout).to.contain('object": "page')
          expect(ctx.stdout).to.contain(`id": "${PAGE_ID}`)
          expect(ctx.stdout).to.contain(`url": "https://www.notion.so/${PAGE_ID_NO_DASHES}`)
        })
    })

    describe('response title is []', () => {
      test
        .do(() => {
          // CRITICAL: Clear nock and cache before setting up mocks to prevent cross-test pollution
          nock.cleanAll()
          cacheManager.clear()
          // Use unique ID to avoid test interference
          nock('https://api.notion.com')
            .get(`/v1/pages/${PAGE_ID_EMPTY_TITLE_2_NO_DASHES}`)
            .reply(200, retrieveOnDbResponseWithEmptyTitle)
        })
        .stdout({ print: process.env.TEST_DEBUG ? true : false })
        .command(['page:retrieve', '--no-truncate', PAGE_ID_EMPTY_TITLE_2])
        .it('shows retrieve page result table', (ctx) => {
          expect(ctx.stdout).to.match(/title.*object.*id.*url/)
          expect(ctx.stdout).to.match(
            new RegExp(`Untitled.*page.*${PAGE_ID_EMPTY_TITLE_2}.*https://www\\.notion\\.so/${PAGE_ID_EMPTY_TITLE_2_NO_DASHES}`)
          )
        })
    })
  })
})
