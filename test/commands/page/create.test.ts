import { expect, test } from '@oclif/test'
import * as nock from 'nock'
import * as sinon from 'sinon'
import { cacheManager } from '../../../dist/cache.js'

// Use valid UUID format for IDs
const PAGE_ID = '11111111-2222-3333-4444-555555555555'
const PAGE_ID_NO_DASHES = PAGE_ID.replace(/-/g, '')
const PARENT_PAGE_ID = '22222222-3333-4444-5555-666666666666'
const PARENT_PAGE_ID_NO_DASHES = PARENT_PAGE_ID.replace(/-/g, '')
const PARENT_DB_ID = '33333333-4444-5555-6666-777777777777'
const PARENT_DB_ID_NO_DASHES = PARENT_DB_ID.replace(/-/g, '')

const createOnPageResponse = {
  object: 'page',
  id: PAGE_ID,
  parent: {
    type: 'page_id',
    page_id: PARENT_PAGE_ID,
  },
  archived: false,
  properties: {
    Name: {
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

const createOnPageResponseWithEmptyTitle = {
  object: 'page',
  id: PAGE_ID,
  parent: {
    type: 'page_id',
    page_id: PARENT_PAGE_ID,
  },
  archived: false,
  properties: {
    Name: {
      id: 'title',
      type: 'title',
      title: [],
    },
  },
  url: `https://www.notion.so/${PAGE_ID_NO_DASHES}`,
}

const createOnDbResponse = {
  object: 'page',
  id: PAGE_ID,
  parent: {
    type: 'database_id',
    database_id: PARENT_DB_ID,
    data_source_id: PARENT_DB_ID,
  },
  archived: false,
  properties: {
    Name: {
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

const createOnDbResponseWithEmptyTitle = {
  object: 'page',
  id: PAGE_ID,
  parent: {
    type: 'database_id',
    database_id: PARENT_DB_ID,
    data_source_id: PARENT_DB_ID,
  },
  archived: false,
  properties: {
    Name: {
      id: 'title',
      type: 'title',
      title: [],
    },
  },
  url: `https://www.notion.so/${PAGE_ID_NO_DASHES}`,
}

describe('page:create', () => {
  let processExitStub: sinon.SinonStub

  beforeEach(() => {
    nock.cleanAll()
    // Clear cache to prevent test interference
    cacheManager.clear()
    // Stub process.exit to prevent tests from hanging
    processExitStub = sinon.stub(process, 'exit' as any)
  })

  afterEach(() => {
    nock.cleanAll()
    processExitStub.restore()
  })

  describe('with parent_page_id flags', () => {
    test
      .do(() => {
        // Mock the pages endpoint for creation
        nock('https://api.notion.com')
          .post('/v1/pages')
          .reply(200, createOnPageResponse)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['page:create', '--no-truncate', '-p', PARENT_PAGE_ID])
      .it('shows create page result table', (ctx) => {
        expect(ctx.stdout).to.match(/title.*object.*id.*url/)
        expect(ctx.stdout).to.match(
          new RegExp(`dummy page title.*page.*${PAGE_ID}.*https://www\\.notion\\.so/${PAGE_ID_NO_DASHES}`)
        )
      })

    describe('with --raw flags', () => {
      test
        .do(() => {
          nock('https://api.notion.com')
            .post('/v1/pages')
            .reply(200, createOnPageResponse)
        })
        .stdout({ print: process.env.TEST_DEBUG ? true : false })
        .command(['page:create', '-p', PARENT_PAGE_ID, '--raw'])
        .it('shows a page object', (ctx) => {
          expect(ctx.stdout).to.contain(PARENT_PAGE_ID)
        })
    })

    describe('response title is []', () => {
      test
        .do(() => {
          nock('https://api.notion.com')
            .post('/v1/pages')
            .reply(200, createOnPageResponseWithEmptyTitle)
        })
        .stdout({ print: process.env.TEST_DEBUG ? true : false })
        .command(['page:create', '--no-truncate', '-p', PARENT_PAGE_ID])
        .it('shows create page result table', (ctx) => {
          expect(ctx.stdout).to.match(/title.*object.*id.*url/)
          expect(ctx.stdout).to.match(
            new RegExp(`Untitled.*page.*${PAGE_ID}.*https://www\\.notion\\.so/${PAGE_ID_NO_DASHES}`)
          )
        })
    })
  })

  describe('with parent_db_id flags', () => {
    test
      .do(() => {
        // Mock the data_sources endpoint for resolveNotionId validation
        nock('https://api.notion.com')
          .get(`/v1/data_sources/${PARENT_DB_ID_NO_DASHES}`)
          .reply(200, {
            object: 'data_source',
            id: PARENT_DB_ID,
            title: []
          })
        // Mock the pages endpoint for creation
        nock('https://api.notion.com')
          .post('/v1/pages')
          .reply(200, createOnDbResponse)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['page:create', '--no-truncate', '-d', PARENT_DB_ID])
      .it('shows create page result table', (ctx) => {
        expect(ctx.stdout).to.match(/title.*object.*id.*url/)
        expect(ctx.stdout).to.match(
          new RegExp(`dummy page title.*page.*${PAGE_ID}.*https://www\\.notion\\.so/${PAGE_ID_NO_DASHES}`)
        )
      })

    describe('with --raw flags', () => {
      test
        .do(() => {
          // Mock the data_sources endpoint for resolveNotionId validation
          nock('https://api.notion.com')
            .get(`/v1/data_sources/${PARENT_DB_ID_NO_DASHES}`)
            .reply(200, {
              object: 'data_source',
              id: PARENT_DB_ID,
              title: []
            })
          // Mock the pages endpoint for creation
          nock('https://api.notion.com')
            .post('/v1/pages')
            .reply(200, createOnDbResponse)
        })
        .stdout({ print: process.env.TEST_DEBUG ? true : false })
        .command(['page:create', '-d', PARENT_DB_ID, '--raw'])
        .it('shows a page object', (ctx) => {
          expect(ctx.stdout).to.contain(PARENT_DB_ID)
        })
    })

    describe('response title is []', () => {
      test
        .do(() => {
          // Mock the data_sources endpoint for resolveNotionId validation
          nock('https://api.notion.com')
            .get(`/v1/data_sources/${PARENT_DB_ID_NO_DASHES}`)
            .reply(200, {
              object: 'data_source',
              id: PARENT_DB_ID,
              title: []
            })
          // Mock the pages endpoint for creation
          nock('https://api.notion.com')
            .post('/v1/pages')
            .reply(200, createOnDbResponseWithEmptyTitle)
        })
        .stdout({ print: process.env.TEST_DEBUG ? true : false })
        .command(['page:create', '--no-truncate', '-d', PARENT_DB_ID])
        .it('shows create page result table', (ctx) => {
          expect(ctx.stdout).to.match(/title.*object.*id.*url/)
          expect(ctx.stdout).to.match(
            new RegExp(`Untitled.*page.*${PAGE_ID}.*https://www\\.notion\\.so/${PAGE_ID_NO_DASHES}`)
          )
        })
    })
  })
})
