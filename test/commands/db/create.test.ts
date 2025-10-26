import { expect, test } from '@oclif/test'
import * as nock from 'nock'
import * as sinon from 'sinon'

const DATABASE_ID = '11111111-2222-3333-4444-555555555555'
const DATABASE_ID_NO_DASHES = DATABASE_ID.replace(/-/g, '')
const PAGE_ID = '11111111-2222-3333-4444-555555555556'
const PAGE_ID_NO_DASHES = PAGE_ID.replace(/-/g, '')

describe('db:create', () => {
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

  const response = {
    object: 'database',
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

  describe('with no flags', () => {
    test
      .do(() => {
        nock('https://api.notion.com')
          .post('/v1/databases')
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['db:create', '--no-truncate', '-t', 'dummy database title', PAGE_ID])
      .it('shows created result table', (ctx) => {
        expect(ctx.stdout).to.match(/Title.*Object.*Id.*Url/)
        expect(ctx.stdout).to.match(
          new RegExp(`dummy database title.*database.*${DATABASE_ID}.*https://www\\.notion\\.so/${DATABASE_ID_NO_DASHES}`)
        )
      })
  })

  describe('with --raw flags', () => {
    test
      .do(() => {
        nock('https://api.notion.com')
          .post('/v1/databases')
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['db:create', PAGE_ID, '-t', 'dummy database title', '--raw'])
      .it('shows created database object when success with title flags', (ctx) => {
        expect(ctx.stdout).to.contain(DATABASE_ID)
        expect(ctx.stdout).to.contain('dummy database title')
      })
  })

  describe('response title is []', () => {
    const titleEmptyResponse = {
      object: 'database',
      id: DATABASE_ID,
      title: [],
      url: `https://www.notion.so/${DATABASE_ID_NO_DASHES}`,
    }

    test
      .do(() => {
        nock('https://api.notion.com')
          .post('/v1/databases')
          .reply(200, titleEmptyResponse)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['db:create', '--no-truncate', '-t', 'dummy database title', PAGE_ID])
      .it('shows created result table', (ctx) => {
        expect(ctx.stdout).to.match(/Title.*Object.*Id.*Url/)
        expect(ctx.stdout).to.match(
          new RegExp(`Untitled.*database.*${DATABASE_ID}.*https://www\\.notion\\.so/${DATABASE_ID_NO_DASHES}`)
        )
      })
  })
})
