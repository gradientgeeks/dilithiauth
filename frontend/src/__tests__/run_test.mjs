import { createServer } from 'vite'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(__dirname, '../..')

async function run() {
  const vite = await createServer({
    root,
    server: { middlewareMode: true },
    appType: 'custom',
    plugins: [
      {
        name: 'test-auth-state-injector',
        transform(code, id) {
          if (id.includes('AuthContext')) {
            console.log('[PLUGIN] Transforming AuthContext id:', id)
            return code.replace(
              'return ctx',
              'const mockDev = globalThis.__MOCK_DEVELOPER__; return { ...ctx, developer: mockDev !== undefined ? mockDev : ctx.developer }'
            )
          }
        }
      }
    ]
  })

  try {
    console.log('Loading test suite via Vite SSR with dynamic auth injection...')
    const testModule = await vite.ssrLoadModule('/src/__tests__/test_suite.tsx')
    if (testModule.default) {
      await testModule.default()
    }
  } catch (err) {
    console.error('Test execution error:', err)
    process.exit(1)
  } finally {
    await vite.close()
  }
}

run()
