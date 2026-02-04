import { expect, test } from '@oclif/test'
import * as nock from 'nock'
import * as sinon from 'sinon'
import { cacheManager } from '../../../dist/cache.js'

// Use valid UUID format for IDs
const PAGE_ID = '11111111-2222-3333-4444-555555555555'
const PAGE_ID_NO_DASHES = PAGE_ID.replace(/-/g, '')
const PARENT_PAGE_ID = '22222222-3333-4444-5555-666666666666'
const PARENT_DB_ID = '33333333-4444-5555-6666-777777777777'

const responseOnPage = {
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

const responseOnPageWithEmptyTitle = {
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
      title: [],
    },
  },
  url: `https://www.notion.so/${PAGE_ID_NO_DASHES}`,
}

const responseOnDb = {
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

const responseOnDbWithEmptyTitle = {
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
      title: [],
    },
  },
  url: `https://www.notion.so/${PAGE_ID_NO_DASHES}`,
}

describe('page:update', () => {
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

  describe('with page_id on a page flags', () => {
    test
      .do(() => {
        nock('https://api.notion.com')
          .patch(`/v1/pages/${PAGE_ID_NO_DASHES}`)
          .reply(200, responseOnPage)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['page:update', '--no-truncate', PAGE_ID])
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
            .patch(`/v1/pages/${PAGE_ID_NO_DASHES}`)
            .reply(200, responseOnPage)
        })
        .stdout({ print: process.env.TEST_DEBUG ? true : false })
        .command(['page:update', PAGE_ID, '--raw'])
        .it('shows a page object', (ctx) => {
          expect(ctx.stdout).to.contain('object": "page')
          expect(ctx.stdout).to.contain(`id": "${PAGE_ID}`)
          expect(ctx.stdout).to.contain(`url": "https://www.notion.so/${PAGE_ID_NO_DASHES}`)
        })
    })

    describe('response title is []', () => {
      test
        .do(() => {
          nock('https://api.notion.com')
            .patch(`/v1/pages/${PAGE_ID_NO_DASHES}`)
            .reply(200, responseOnPageWithEmptyTitle)
        })
        .stdout({ print: process.env.TEST_DEBUG ? true : false })
        .command(['page:update', '--no-truncate', PAGE_ID])
        .it('shows retrieve page result table', (ctx) => {
          expect(ctx.stdout).to.match(/title.*object.*id.*url/)
          expect(ctx.stdout).to.match(
            new RegExp(`Untitled.*page.*${PAGE_ID}.*https://www\\.notion\\.so/${PAGE_ID_NO_DASHES}`)
          )
        })
    })
  })

  describe('with page_id on a db flags', () => {
    test
      .do(() => {
        nock('https://api.notion.com')
          .patch(`/v1/pages/${PAGE_ID_NO_DASHES}`)
          .reply(200, responseOnDb)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['page:update', '--no-truncate', PAGE_ID])
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
            .patch(`/v1/pages/${PAGE_ID_NO_DASHES}`)
            .reply(200, responseOnDb)
        })
        .stdout({ print: process.env.TEST_DEBUG ? true : false })
        .command(['page:update', PAGE_ID, '--raw'])
        .it('shows a page object', (ctx) => {
          expect(ctx.stdout).to.contain('object": "page')
          expect(ctx.stdout).to.contain(`id": "${PAGE_ID}`)
          expect(ctx.stdout).to.contain(`url": "https://www.notion.so/${PAGE_ID_NO_DASHES}`)
        })
    })

    describe('response title is []', () => {
      test
        .do(() => {
          nock('https://api.notion.com')
            .patch(`/v1/pages/${PAGE_ID_NO_DASHES}`)
            .reply(200, responseOnDbWithEmptyTitle)
        })
        .stdout({ print: process.env.TEST_DEBUG ? true : false })
        .command(['page:update', '--no-truncate', PAGE_ID])
        .it('shows retrieve page result table', (ctx) => {
          expect(ctx.stdout).to.match(/title.*object.*id.*url/)
          expect(ctx.stdout).to.match(
            new RegExp(`Untitled.*page.*${PAGE_ID}.*https://www\\.notion\\.so/${PAGE_ID_NO_DASHES}`)
          )
        })
    })
  })
})
