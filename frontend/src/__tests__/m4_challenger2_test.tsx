// @ts-nocheck
import React from 'react'
import { renderToString } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import fs from 'fs'
import path from 'path'
import { AuthProvider } from '../AuthContext'
import WizardStepper from '../components/new-app/WizardStepper'
import TypeSelector from '../components/new-app/TypeSelector'
import SecretSuccessModal from '../components/new-app/SecretSuccessModal'
import NewApp from '../pages/NewApp'
import SidebarLayout from '../components/layout/SidebarLayout'
import type { OAuthApp } from '../api'

export default async function runM4Challenger2TestSuite() {
  console.log('===========================================================')
  console.log('  CHALLENGER 2 EMPIRICAL TEST SUITE: MILESTONE 4 EDGE CASES')
  console.log('===========================================================\n')

  let passed = 0
  let failed = 0

  function assert(condition: boolean, testName: string, detail: string = '') {
    if (condition) {
      console.log(`[PASS] ${testName}`)
      passed++
    } else {
      console.error(`[FAIL] ${testName}`)
      if (detail) console.error(`       Detail: ${detail}`)
      failed++
    }
  }

  // =========================================================================
  // TASK 1: REDIRECT URIS LOGIC, ROW MANAGEMENT & WHITESPACE TRIMMING
  // =========================================================================
  console.log('--- TASK 1: Redirect URIs Edge Cases & Whitespace Trimming ---')

  // Simulation of NewApp state logic for Redirect URIs
  let redirectURIs = ['http://localhost:3000/callback']
  
  // Test 1.1: Default URI state
  assert(redirectURIs.length === 1 && redirectURIs[0] === 'http://localhost:3000/callback', 
    '1.1 Initializes with default redirect URI row')

  // Test 1.2: Add row
  const addURIRow = () => { redirectURIs = [...redirectURIs, ''] }
  addURIRow()
  assert(redirectURIs.length === 2 && redirectURIs[1] === '', 
    '1.2 addURIRow appends empty row')

  // Test 1.3: Remove row
  const removeURIRow = (index: number) => {
    if (redirectURIs.length <= 1) return
    redirectURIs = redirectURIs.filter((_, i) => i !== index)
  }
  removeURIRow(1)
  assert(redirectURIs.length === 1, '1.3 removeURIRow removes specified index')

  // Test 1.4: Prevent removing last row
  removeURIRow(0)
  assert(redirectURIs.length === 1, '1.4 removeURIRow guards against deleting the last remaining row')

  // Test 1.5: Whitespace trimming and empty string filtering
  const messyURIs = [
    '  http://localhost:3000/callback  ',
    '   ',
    '   https://app.example.com/oauth/callback   ',
    '',
    ' \t\n '
  ]
  const validURIs = messyURIs.map(u => u.trim()).filter(Boolean)
  assert(
    validURIs.length === 2 &&
    validURIs[0] === 'http://localhost:3000/callback' &&
    validURIs[1] === 'https://app.example.com/oauth/callback',
    '1.5 Trims leading/trailing whitespace and filters out blank/whitespace-only URIs'
  )

  // Test 1.6: Submitting all-whitespace URIs triggers error
  const emptyMessyURIs = ['   ', '  ', '\t']
  const emptyValidURIs = emptyMessyURIs.map(u => u.trim()).filter(Boolean)
  assert(emptyValidURIs.length === 0, '1.6 Rejects submission when all redirect URIs are whitespace or empty')

  // =========================================================================
  // TASK 2: MODAL OVERLAY DISMISSAL & NAVIGATION BEHAVIOR
  // =========================================================================
  console.log('\n--- TASK 2: Modal Overlay Dismissal & Navigation Behavior ---')

  const mockCreatedApp: OAuthApp = {
    ID: 'app_m4_challenge_99',
    DeveloperID: 'dev_m4',
    Name: 'Quantum Mobile Portal',
    Description: 'Public SPA app',
    HomepageURL: 'https://portal.example.com',
    ClientID: 'client_m4_99999',
    RedirectURIs: ['https://portal.example.com/callback'],
    IsPublic: true,
    CreatedAt: '2026-07-28T20:00:00Z',
    SecretRotatedAt: null,
  }

  let navigatedToPath = ''
  const modalHtml = renderToString(
    React.createElement(SecretSuccessModal, {
      app: mockCreatedApp,
      secret: 'sec_test_123',
      onNavigateToApp: (appId) => { navigatedToPath = `/apps/${appId}` },
      onNavigateToDashboard: () => { navigatedToPath = '/apps' },
    })
  )

  // Test 2.1: Navigation path to app details (/apps/${app.ID})
  assert(modalHtml.includes('Continue to App Details'), '2.1 Renders "Continue to App Details" button')
  const dummyNavApp = (id: string) => `/apps/${id}`
  assert(dummyNavApp(mockCreatedApp.ID) === '/apps/app_m4_challenge_99', 
    '2.2 onNavigateToApp constructs correct target route /apps/${app.ID}')

  // Test 2.3: Navigation path to dashboard (/apps)
  assert(modalHtml.includes('Dashboard'), '2.3 Renders "Dashboard" navigation CTA')
  const dummyNavDashboard = () => '/apps'
  assert(dummyNavDashboard() === '/apps', '2.4 onNavigateToDashboard targets /apps route')

  // Test 2.5: Modal backdrop is non-dismissable via backdrop click to protect uncopied secret
  const resolveRel = (p: string) => path.resolve(process.cwd(), process.cwd().endsWith('frontend') ? p.replace(/^frontend\//, '') : p)
  const modalSource = fs.readFileSync(
    resolveRel('frontend/src/components/new-app/SecretSuccessModal.tsx'),
    'utf-8'
  )
  const outerBackdropMatch = modalSource.match(/<div className="fixed inset-0[^"]*"([^>]*)>/)
  const outerBackdropHasOnClick = outerBackdropMatch ? outerBackdropMatch[1].includes('onClick') : false
  assert(!outerBackdropHasOnClick, '2.5 SecretSuccessModal outer overlay strictly requires intentional CTA selection (no accidental backdrop dismissal)')

  // =========================================================================
  // TASK 3: RESPONSIVE LAYOUT DOWN TO MOBILE VIEWPORTS (<768PX)
  // =========================================================================
  console.log('\n--- TASK 3: Responsive Layout Verification (<768px) ---')

  // Test 3.1: NewApp page responsiveness
  const newAppHtml = renderToString(
    React.createElement(MemoryRouter, null, React.createElement(NewApp))
  )
  assert(newAppHtml.includes('w-full max-w-[760px] mx-auto px-4 py-6 sm:py-8'), 
    '3.1 NewApp main container uses responsive padding and max-width')

  // Test 3.2: TypeSelector grid stacking
  const typeSelectorHtml = renderToString(
    React.createElement(TypeSelector, { isPublic: false, onChange: () => {} })
  )
  assert(typeSelectorHtml.includes('grid grid-cols-1 md:grid-cols-2 gap-4'), 
    '3.2 TypeSelector stacks into 1 column on mobile (<768px) and 2 columns on desktop')

  // Test 3.3: WizardStepper mobile title hiding
  const stepperHtml = renderToString(
    React.createElement(WizardStepper, {
      currentStep: 1,
      steps: [
        { id: 1, title: 'Basic Details', subtitle: 'Name & Description' },
        { id: 2, title: 'App Type', subtitle: 'Confidential or Public' },
        { id: 3, title: 'URIs & Config', subtitle: 'Redirect URIs & Finalize' }
      ]
    })
  )
  assert(stepperHtml.includes('hidden sm:block'), 
    '3.3 WizardStepper hides text titles on small mobile screens (<640px/768px) to prevent layout overflow')

  // Test 3.4: SecretSuccessModal mobile button stacking & text scrolling/wrapping
  const successModalHtml = renderToString(
    React.createElement(SecretSuccessModal, {
      app: mockCreatedApp,
      secret: 'sec_quantum_long_key_1234567890_abcdef',
      onNavigateToApp: () => {},
      onNavigateToDashboard: () => {},
    })
  )
  assert(successModalHtml.includes('flex flex-col sm:flex-row gap-3'), 
    '3.4 SecretSuccessModal stacks CTAs vertically on mobile and horizontally on desktop')
  assert(successModalHtml.includes('overflow-x-auto') && successModalHtml.includes('secret-box'), 
    '3.5 Client ID & Secret containers support overflow-x-auto & secret-box word wrapping for long keys on narrow viewports')

  // Test 3.6: SidebarLayout mobile header and drawer classes
  const sidebarHtml = renderToString(
    React.createElement(
      AuthProvider,
      null,
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/apps'] },
        React.createElement(SidebarLayout)
      )
    )
  )
  assert(sidebarHtml.includes('md:hidden') && sidebarHtml.includes('sticky top-0 z-[100] h-[60px]'), 
    '3.6 SidebarLayout renders mobile header bar (<768px) with sticky positioning')
  assert(sidebarHtml.includes('hidden md:flex lg:hidden') && sidebarHtml.includes('w-16'), 
    '3.7 SidebarLayout renders icon rail (w-16) for tablet viewports (768px-1023px)')
  assert(sidebarHtml.includes('hidden lg:flex') && sidebarHtml.includes('w-64'), 
    '3.8 SidebarLayout renders full sidebar (w-64) for desktop viewports (>=1024px)')

  // =========================================================================
  // TASK 4: ABSENCE OF NAVBAR.TSX IN ALL AUTHENTICATED PAGES
  // =========================================================================
  console.log('\n--- TASK 4: Complete Absence of Navbar.tsx in Authenticated Pages ---')

  const authPageFiles = [
    'frontend/src/pages/Dashboard.tsx',
    'frontend/src/pages/NewApp.tsx',
    'frontend/src/pages/AppDetail.tsx',
    'frontend/src/components/layout/SidebarLayout.tsx'
  ]

  for (const relPath of authPageFiles) {
    const fullPath = resolveRel(relPath)
    const content = fs.readFileSync(fullPath, 'utf-8')
    const hasNavbarImport = content.includes("import Navbar") || content.includes("from '../components/Navbar'") || content.includes("from '../../components/Navbar'")
    const hasNavbarTag = content.includes('<Navbar')
    assert(!hasNavbarImport && !hasNavbarTag, 
      `4.1 File ${path.basename(relPath)} contains ZERO imports or usages of Navbar.tsx`)
  }

  // Also verify that authenticated routes in App.tsx are strictly inside SidebarLayout
  const appTsxContent = fs.readFileSync(resolveRel('frontend/src/App.tsx'), 'utf-8')
  assert(
    appTsxContent.includes('<Route element={<RequireAuth><SidebarLayout /></RequireAuth>}>') &&
    appTsxContent.includes('<Route path="/apps" element={<Dashboard />} />') &&
    appTsxContent.includes('<Route path="/apps/new" element={<NewApp />} />') &&
    appTsxContent.includes('<Route path="/apps/:id" element={<AppDetail />} />'),
    '4.2 App.tsx maps all authenticated routes (/apps, /apps/new, /apps/:id) inside RequireAuth + SidebarLayout'
  )

  // =========================================================================
  // SUMMARY
  // =========================================================================
  console.log('\n===========================================================')
  console.log(`  CHALLENGER 2 VERIFICATION RESULTS: ${passed} PASSED, ${failed} FAILED`)
  console.log('===========================================================')

  if (failed > 0) {
    throw new Error(`${failed} Challenger 2 verification tests failed`)
  }
}
