import { expect, test } from '@oclif/test'
import * as nock from 'nock'
import * as sinon from 'sinon'

// Use a valid UUID format for block ID
const BLOCK_ID = '12345678-1234-1234-1234-123456789012'

const response = {
  object: 'block',
  id: BLOCK_ID,
  parent: {
    type: 'page_id',
    page_id: '12345678-1234-1234-1234-123456789013',
  },
  has_children: true,
  archived: false,
  type: 'child_page',
  child_page: {
    title: 'dummy child page title',
  },
}

describe('block:retrieve', () => {
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
          .get(`/v1/blocks/${BLOCK_ID}`)
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['block:retrieve', BLOCK_ID])
      .it('shows retrieved block object when success', (ctx) => {
        expect(ctx.stdout).to.match(/object.*id.*type.*parent.*content/)
        expect(ctx.stdout).to.match(new RegExp(`block.*${BLOCK_ID}.*child_page.*dummy child page title`))
      })
  })

  describe('shows raw json result', () => {
    test
      .do(() => {
        nock('https://api.notion.com')
          .get(`/v1/blocks/${BLOCK_ID}`)
          .reply(200, response)
      })
      .stdout({ print: process.env.TEST_DEBUG ? true : false })
      .command(['block:retrieve', BLOCK_ID, '--raw'])
      .it('shows retrieved block object when success', (ctx) => {
        expect(ctx.stdout).to.contain('object": "block')
        expect(ctx.stdout).to.contain('type": "child_page')
      })
  })
})
