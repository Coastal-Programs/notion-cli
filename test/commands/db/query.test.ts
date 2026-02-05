import { expect, test } from '@oclif/test'
import * as nock from 'nock'
import * as sinon from 'sinon'

const DATABASE_ID = '11111111-2222-3333-4444-555555555555'
const DATABASE_ID_NO_DASHES = DATABASE_ID.replace(/-/g, '')
const PAGE_ID = '11111111-2222-3333-4444-555555555556'
const PAGE_ID_NO_DASHES = PAGE_ID.replace(/-/g, '')

const response = {
  object: 'list',
  results: [
    {
      object: 'page',
      id: PAGE_ID,
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
    },
  ],
}

const titleEmptyResponse = {
  object: 'list',
  results: [
    {
      object: 'page',
      id: PAGE_ID,
      properties: {
        Name: {
          id: 'title',
          type: 'title',
          title: [],
        },
      },
      url: `https://www.notion.so/${PAGE_ID_NO_DASHES}`,
    },
  ],
}

describe('db:query', () => {
  let processExitStub: sinon.SinonStub

  beforeEach(() => {
    nock.cleanAll()
    // Stub process.exit to prevent tests from hanging
    processExitStub = sinon.stub(process, 'exit' as any)
  })

  afterEach(() => {
    nock.cleanAll()
    processExitStub.restore()
  })

  describe('with raw filter flags', () => {
    test
      .do(() => {
        // Mock the data_sources endpoint that resolveNotionId calls
        nock('https://api.notion.com')
          .get(`/v1/data_sources/${DATABASE_ID_NO_DASHES}`)
          .reply(200, { id: DATABASE_ID, type: 'database' })
        // Mock the actual query endpoint (uses data_sources not databases)
        nock('https://api.notion.com')
          .post(`/v1/data_sources/${DATABASE_ID_NO_DASHES}/query`)
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['db:query', DATABASE_ID, '-a', '{"and": []}'])
      .it('shows query result table', (ctx) => {
        expect(ctx.stdout).to.match(/title.*object.*id.*url/)
        expect(ctx.stdout).to.match(
          new RegExp(`dummy page title.*page.*${PAGE_ID}.*https://www\\.notion\\.so/${PAGE_ID_NO_DASHES}`)
        )
      })
  })

  describe('with --raw flags', () => {
    test
      .do(() => {
        // Mock the data_sources endpoint that resolveNotionId calls
        nock('https://api.notion.com')
          .get(`/v1/data_sources/${DATABASE_ID_NO_DASHES}`)
          .reply(200, { id: DATABASE_ID, type: 'database' })
        // Mock the actual query endpoint (uses data_sources not databases)
        nock('https://api.notion.com')
          .post(`/v1/data_sources/${DATABASE_ID_NO_DASHES}/query`)
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['db:query', DATABASE_ID, '-a', '{"and": []}', '--raw'])
      .it('shows query result page objects', (ctx) => {
        expect(ctx.stdout).to.contain(PAGE_ID)
        expect(ctx.stdout).to.contain('dummy page title')
      })
  })

  describe('return title is []', () => {
    test
      .do(() => {
        // Mock the data_sources endpoint that resolveNotionId calls
        nock('https://api.notion.com')
          .get(`/v1/data_sources/${DATABASE_ID_NO_DASHES}`)
          .reply(200, { id: DATABASE_ID, type: 'database' })
        // Mock the actual query endpoint (uses data_sources not databases)
        nock('https://api.notion.com')
          .post(`/v1/data_sources/${DATABASE_ID_NO_DASHES}/query`)
          .reply(200, titleEmptyResponse)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['db:query', DATABASE_ID, '-a', '{"and": []}'])
      .it('shows query result table', (ctx) => {
        expect(ctx.stdout).to.match(/title.*object.*id.*url/)
        expect(ctx.stdout).to.match(
          new RegExp(`Untitled.*page.*${PAGE_ID}.*https://www\\.notion\\.so/${PAGE_ID_NO_DASHES}`)
        )
      })
  })
})
