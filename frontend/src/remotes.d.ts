// Type declarations for Module Federation remotes.
// These tell TypeScript about modules loaded at runtime from the docs micro frontend.

declare module 'docs/DocsApp' {
  import type { ComponentType } from 'react'

  interface DocsAppProps {
    standalone?: boolean
  }

  const DocsApp: ComponentType<DocsAppProps>
  export default DocsApp
}
