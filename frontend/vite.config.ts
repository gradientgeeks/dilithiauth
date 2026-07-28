import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { federation } from '@module-federation/vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  // Remote URL:
  //  - Dev:        http://localhost:5174/assets/remoteEntry.js
  //  - Production: set VITE_DOCS_REMOTE_URL in your CI / Firebase env
  const docsRemoteUrl =
    env.VITE_DOCS_REMOTE_URL ||
    (mode === 'production'
      ? 'https://docs.dilithiauth.gradientgeeks.tech/assets/remoteEntry.js'
      : 'http://localhost:5174/assets/remoteEntry.js')

  return {
    plugins: [
      tailwindcss(),
      react(),
      federation({
        name: 'host',
        remotes: {
          docs: {
            type: 'module',
            name: 'docs',
            entry: docsRemoteUrl,
            shareScope: 'default',
          },
        },
        shared: ['react', 'react-dom', 'react-router-dom'],
      }),
    ],
    server: {
      proxy: {
        '/api': 'https://dilithiauthapis.gradientgeeks.tech',
      },
    },
    build: {
      target: 'esnext',
      outDir: 'dist',
      emptyOutDir: true,
    },
  }
})
