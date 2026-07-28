// Allow CSS/asset side-effect imports (e.g. import './index.css')
declare module '*.css' {
  const content: string
  export default content
}
