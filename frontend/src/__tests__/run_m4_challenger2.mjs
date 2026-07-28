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
    console.log('Loading Challenger 2 M4 test suite via Vite SSR...')
    const testModule = await vite.ssrLoadModule('/src/__tests__/m4_challenger2_test.tsx')
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
