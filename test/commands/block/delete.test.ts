import { expect, test } from '@oclif/test'
import * as nock from 'nock'
import * as sinon from 'sinon'

const BLOCK_ID = '87654321-4321-4321-4321-210987654321'
const PAGE_ID = '87654321-4321-4321-4321-210987654322'

const response = {
  object: 'block',
  id: BLOCK_ID,
  parent: {
    type: 'page_id',
    page_id: PAGE_ID,
  },
  has_children: false,
  archived: true,
  type: 'heading_2',
  heading_2: {
    rich_text: [
      {
        type: 'text',
        plain_text: 'dummy-heading-2-content',
      },
    ],
  },
}

describe('block:delete', () => {
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
          .delete(`/v1/blocks/${BLOCK_ID}`)
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['block:delete', BLOCK_ID])
      .it('shows deleted block object when success', (ctx) => {
        expect(ctx.stdout).to.match(/Object.*Id.*Type.*Parent.*Content/)
        expect(ctx.stdout).to.match(new RegExp(`block.*${BLOCK_ID}.*heading_2.*dummy-heading-2-content`))
      })
  })

  describe('shows raw json result', () => {
    test
      .do(() => {
        nock('https://api.notion.com')
          .delete(`/v1/blocks/${BLOCK_ID}`)
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['block:delete', BLOCK_ID, '--raw'])
      .it('shows deleted block object when success', (ctx) => {
        expect(ctx.stdout).to.contain('object": "block')
        expect(ctx.stdout).to.contain(`id": "${BLOCK_ID}`)
        expect(ctx.stdout).to.contain('archived": true')
      })
  })
})
