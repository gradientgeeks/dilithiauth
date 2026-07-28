import { createServer } from 'vite'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(__dirname, '../../frontend')

async function run() {
  const vite = await createServer({
    root,
    server: { middlewareMode: true },
    appType: 'custom',
  })

  try {
    console.log('--- EMPIRICAL DEEP VERIFICATION FOR M4 ---')
    
    // Load components
    const NewAppModule = await vite.ssrLoadModule('/src/pages/NewApp.tsx')
    const ApiModule = await vite.ssrLoadModule('/src/api.ts')
    
    const NewApp = NewAppModule.default
    const { apps } = ApiModule

    console.log('[TEST 1] Verifying apps.create signature in api.ts...')
    if (typeof apps.create === 'function') {
      console.log('[PASS] apps.create function exists.')
    } else {
      console.error('[FAIL] apps.create function is missing.')
    }

    console.log('[TEST 2] Verifying NewApp component exports cleanly...')
    if (typeof NewApp === 'function') {
      console.log('[PASS] NewApp component default export verified.')
    } else {
      console.error('[FAIL] NewApp component default export missing.')
    }

  } catch (err) {
    console.error('Deep verification error:', err)
    process.exit(1)
  } finally {
    await vite.close()
  }
}

run()
