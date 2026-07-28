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
  })

  try {
    console.log('Running Milestone 3 Empirical Challenge Test Suite via Vite SSR...\n')
    const testModule = await vite.ssrLoadModule('/src/__tests__/m3_challenge_test.tsx')
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
