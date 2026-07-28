import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import './index.css'
import DocsApp from './DocsApp'

// Standalone entrypoint — used for `npm run dev` and direct preview.
// When consumed as a micro frontend by the host, main.tsx is NOT used;
// only DocsApp.tsx is imported via the remoteEntry.js federation module.
ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <DocsApp standalone />
    </BrowserRouter>
  </React.StrictMode>
)
