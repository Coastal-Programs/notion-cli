import { expect, test } from '@oclif/test'
import * as nock from 'nock'
import * as sinon from 'sinon'

const BLOCK_ID = '11111111-2222-3333-4444-555555555555'
const BLOCK_ID_NO_DASHES = BLOCK_ID.replace(/-/g, '')
const PAGE_ID = '11111111-2222-3333-4444-555555555556'

const response = {
  object: 'list',
  results: [
    {
      object: 'block',
      id: BLOCK_ID,
      parent: {
        type: 'page_id',
        page_id: PAGE_ID,
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

describe('block:append', () => {
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
          .patch(`/v1/blocks/${BLOCK_ID_NO_DASHES}/children`, (_body) => {
            // Accept any valid request body
            return true
          })
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command([
        'block:append',
        '--no-truncate',
        '-b',
        BLOCK_ID,
        '-c',
        '[{"type": "heading_2", "heading_2": {"rich_text": [{"type": "text", "text": {"content": "dummy-heading-2-content"}}]}}]',
      ])
      .it('shows appended block object when success', (ctx) => {
        expect(ctx.stdout).to.match(/Object.*Id.*Type.*Parent.*Content/)
        expect(ctx.stdout).to.match(new RegExp(`block.*${BLOCK_ID}.*heading_2.*dummy-heading-2-content`))
      })
  })

  describe('shows raw json result', () => {
    test
      .do(() => {
        nock('https://api.notion.com')
          .patch(`/v1/blocks/${BLOCK_ID_NO_DASHES}/children`, (_body) => {
            // Accept any valid request body
            return true
          })
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command([
        'block:append',
        '--raw',
        '-b',
        BLOCK_ID,
        '-c',
        '[{"type": "heading_2", "heading_2": {"rich_text": [{"type": "text", "text": {"content": "dummy-heading-2-content"}}]}}]',
      ])
      .it('shows updated block object when success', (ctx) => {
        expect(ctx.stdout).to.contain('object": "list')
        expect(ctx.stdout).to.contain('results": [')
        expect(ctx.stdout).to.contain('type": "heading_2')
      })
  })
})
