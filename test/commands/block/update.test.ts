import { expect, test } from '@oclif/test'
import * as nock from 'nock'
import * as sinon from 'sinon'

const BLOCK_ID = '11111111-2222-3333-4444-555555555555'
const BLOCK_ID_NO_DASHES = BLOCK_ID.replace(/-/g, '')
const PAGE_ID = '11111111-2222-3333-4444-555555555556'

const response = {
  object: 'block',
  id: BLOCK_ID,
  parent: {
    type: 'page_id',
    page_id: PAGE_ID,
  },
  has_children: false,
  archived: true,
  in_trash: false,
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

describe('block:update', () => {
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
          .patch(`/v1/blocks/${BLOCK_ID_NO_DASHES}`)
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['block:update', BLOCK_ID, '--no-truncate'])
      .it('shows deleted block object when success', (ctx) => {
        expect(ctx.stdout).to.match(/object.*id.*type.*parent.*content/)
        expect(ctx.stdout).to.match(new RegExp(`block.*${BLOCK_ID}.*heading_2.*dummy-heading-2-content`))
      })
  })
  describe('shows raw json result', () => {
    test
      .do(() => {
        nock('https://api.notion.com')
          .patch(`/v1/blocks/${BLOCK_ID_NO_DASHES}`)
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['block:update', BLOCK_ID, '--raw'])
      .it('shows updated block object when success', (ctx) => {
        expect(ctx.stdout).to.contain('object": "block')
        expect(ctx.stdout).to.contain(`id": "${BLOCK_ID}`)
        expect(ctx.stdout).to.contain('archived": true')
      })
  })
})
