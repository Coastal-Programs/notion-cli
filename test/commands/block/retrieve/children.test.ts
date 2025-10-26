import { expect, test } from '@oclif/test'
import * as nock from 'nock'
import * as sinon from 'sinon'

const BLOCK_ID = '12345678-1234-1234-1234-123456789012'

const response = {
  object: 'list',
  results: [
    {
      object: 'block',
      id: BLOCK_ID,
      parent: {
        type: 'page_id',
        page_id: '12345678-1234-1234-1234-123456789013',
      },
      has_children: true,
      archived: false,
      type: 'heading_2',
      heading_2: {
        rich_text: [
          {
            type: 'text',
            plain_text: 'dummy-heading-2-content',
          },
        ],
      },
    },
  ],
  next_cursor: null,
  has_more: false,
  type: 'block',
  block: {},
}

describe('block:retrieve:children', () => {
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

  describe('shows ux.table result', () => {
    test
      .do(() => {
        nock('https://api.notion.com')
          .get(`/v1/blocks/${BLOCK_ID}/children`)
          .query(true) // Accept any query parameters
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['block:retrieve:children', BLOCK_ID, '--no-truncate'])
      .it('shows retrieved block children when success', (ctx) => {
        expect(ctx.stdout).to.match(/Object.*Id.*Type.*Content/)
        expect(ctx.stdout).to.match(new RegExp(`block.*${BLOCK_ID}.*heading_2.*dummy-heading-2-content`))
      })
  })

  describe('shows raw json result', () => {
    test
      .do(() => {
        nock('https://api.notion.com')
          .get(`/v1/blocks/${BLOCK_ID}/children`)
          .query(true) // Accept any query parameters
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['block:retrieve:children', BLOCK_ID, '--raw'])
      .it('shows retrieved block children when success', (ctx) => {
        expect(ctx.stdout).to.contain('object": "list')
        expect(ctx.stdout).to.contain('results": [')
        expect(ctx.stdout).to.contain('type": "block')
      })
  })
})
