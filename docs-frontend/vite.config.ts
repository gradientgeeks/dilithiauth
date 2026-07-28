import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { federation } from '@module-federation/vite'

// The host app URL — used for CORS headers in dev
const HOST_ORIGIN = process.env.VITE_HOST_ORIGIN || 'http://localhost:5173'

export default defineConfig({
  plugins: [
    tailwindcss(),
    react(),
    federation({
      name: 'docs',
      filename: 'remoteEntry.js',
      exposes: {
        // The host imports via: import('docs/DocsApp')
        './DocsApp': './src/DocsApp.tsx',
      },
      shared: ['react', 'react-dom', 'react-router-dom'],
    }),
  ],
  server: {
    port: 5174,
    cors: {
      origin: HOST_ORIGIN,
      credentials: true,
    },
  },
  preview: {
    port: 5174,
    cors: {
      origin: HOST_ORIGIN,
      credentials: true,
    },
  },
  build: {
    target: 'esnext',
    minify: false,
    cssCodeSplit: false,
    outDir: 'dist',
    emptyOutDir: true,
  },
})
