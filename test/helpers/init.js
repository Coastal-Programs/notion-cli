const path = require('path')
process.env.TS_NODE_PROJECT = path.resolve('test/tsconfig.json')
process.env.NODE_ENV = 'development'

// CRITICAL: Polyfill fetch BEFORE any modules load
// This must happen before notion.ts imports @notionhq/client
const fetch = require('node-fetch')
global.fetch = fetch
global.Headers = fetch.Headers
global.Request = fetch.Request
global.Response = fetch.Response

if (process.env.TEST_DEBUG) {
  console.log('✓ fetch polyfilled with node-fetch in init.js')
  console.log('✓ fetch.name =', global.fetch.name)
}

global.oclif = global.oclif || {}
global.oclif.columns = 80
