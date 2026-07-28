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
    console.log('Running Milestone 2 Empirical Challenge Test Suite via Vite SSR...\n')
    const testModule = await vite.ssrLoadModule('/src/__tests__/m2_challenge_test.tsx')
    if (testModule.default) {
      await testModule.default()
    }
  } catch (err) {
    console.error('Test execution error:', err)
  } finally {
    await vite.close()
  }
}

run()
