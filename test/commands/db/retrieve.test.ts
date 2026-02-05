import { expect, test } from '@oclif/test'
import * as nock from 'nock'
import * as sinon from 'sinon'
import { cacheManager } from '../../../dist/cache.js'

const DATABASE_ID = '11111111-2222-3333-4444-555555555555'
const DATABASE_ID_NO_DASHES = DATABASE_ID.replace(/-/g, '')

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
  id: DATABASE_ID,
  title: [],
  url: `https://www.notion.so/${DATABASE_ID_NO_DASHES}`,
}

describe('db:retrieve', () => {
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

  describe('with no flags', () => {
    test
      .do(() => {
        // Mock the data_sources endpoint - called twice (once for resolveNotionId, once for retrieve)
        nock('https://api.notion.com')
          .get(`/v1/data_sources/${DATABASE_ID_NO_DASHES}`)
          .times(2)
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
        // Mock the data_sources endpoint - called twice (once for resolveNotionId, once for retrieve)
        nock('https://api.notion.com')
          .get(`/v1/data_sources/${DATABASE_ID_NO_DASHES}`)
          .times(2)
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
        // Mock the data_sources endpoint - called twice (once for resolveNotionId, once for retrieve)
        nock('https://api.notion.com')
          .get(`/v1/data_sources/${DATABASE_ID_NO_DASHES}`)
          .times(2)
          .reply(200, titleEmptyResponse)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['db:retrieve', '--no-truncate', DATABASE_ID])
      .it('shows retrieved result table', (ctx) => {
        expect(ctx.stdout).to.match(/title.*object.*id.*url/)
        expect(ctx.stdout).to.match(
          new RegExp(`Untitled.*data_source.*${DATABASE_ID}.*https://www\\.notion\\.so/${DATABASE_ID_NO_DASHES}`)
        )
      })
  })
})
