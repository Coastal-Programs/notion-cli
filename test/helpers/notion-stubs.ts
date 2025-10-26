import sinon from 'sinon'
import * as notion from '../../src/notion'

let currentSandbox: sinon.SinonSandbox | null = null

/**
 * Stub Notion client functions for testing
 * @param stubs - Object mapping function names to their stub implementations
 * @returns Object with restore method to clean up stubs
 */
export function stubNotionClient(stubs: Partial<Record<keyof typeof notion, (...args: any[]) => any>>) {
  // Clean up any existing sandbox
  if (currentSandbox) {
    currentSandbox.restore()
  }

  // Create new sandbox
  currentSandbox = sinon.createSandbox()

  // Stub each provided function
  Object.entries(stubs).forEach(([methodName, implementation]) => {
    const method = methodName as keyof typeof notion
    if (typeof notion[method] === 'function') {
      currentSandbox!.stub(notion, method).callsFake(implementation as any)
    }
  })

  return {
    restore: () => {
      if (currentSandbox) {
        currentSandbox.restore()
        currentSandbox = null
      }
    }
  }
}

/**
 * Mock block data factory
 */
export const mockBlock = (id = 'test-block-id', overrides: any = {}) => ({
  object: 'block' as const,
  id,
  type: 'paragraph' as const,
  paragraph: {
    rich_text: [
      {
        type: 'text' as const,
        text: { content: 'Mock content' },
        plain_text: 'Mock content',
        href: null,
        annotations: {
          bold: false,
          italic: false,
          strikethrough: false,
          underline: false,
          code: false,
          color: 'default' as const,
        },
      },
    ],
    color: 'default' as const,
  },
  parent: {
    type: 'page_id' as const,
    page_id: 'test-page-id',
  },
  has_children: false,
  archived: false,
  created_time: '2025-01-01T00:00:00.000Z',
  last_edited_time: '2025-01-01T00:00:00.000Z',
  created_by: { object: 'user' as const, id: 'user-id' },
  last_edited_by: { object: 'user' as const, id: 'user-id' },
  ...overrides,
})

/**
 * Mock page data factory
 */
export const mockPage = (id = 'test-page-id', overrides: any = {}) => ({
  object: 'page' as const,
  id,
  created_time: '2025-01-01T00:00:00.000Z',
  last_edited_time: '2025-01-01T00:00:00.000Z',
  created_by: { object: 'user' as const, id: 'user-id' },
  last_edited_by: { object: 'user' as const, id: 'user-id' },
  cover: null,
  icon: null,
  parent: {
    type: 'database_id' as const,
    database_id: 'test-db-id',
  },
  archived: false,
  properties: {},
  url: `https://notion.so/${id}`,
  ...overrides,
})

/**
 * Mock database data factory
 */
export const mockDatabase = (id = 'test-db-id', overrides: any = {}) => ({
  object: 'database' as const,
  id,
  created_time: '2025-01-01T00:00:00.000Z',
  last_edited_time: '2025-01-01T00:00:00.000Z',
  created_by: { object: 'user' as const, id: 'user-id' },
  last_edited_by: { object: 'user' as const, id: 'user-id' },
  title: [
    {
      type: 'text' as const,
      text: { content: 'Test Database' },
      plain_text: 'Test Database',
      href: null,
      annotations: {
        bold: false,
        italic: false,
        strikethrough: false,
        underline: false,
        code: false,
        color: 'default' as const,
      },
    },
  ],
  description: [],
  icon: null,
  cover: null,
  properties: {},
  parent: {
    type: 'page_id' as const,
    page_id: 'test-page-id',
  },
  url: `https://notion.so/${id}`,
  archived: false,
  is_inline: false,
  ...overrides,
})

/**
 * Clean up all stubs (call in afterEach or test cleanup)
 */
export function restoreAllStubs() {
  if (currentSandbox) {
    currentSandbox.restore()
    currentSandbox = null
  }
}
