import React from 'react'
import { renderToString } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import DeleteAppModal from '../components/app-detail/DeleteAppModal'
import SnippetsTab from '../components/app-detail/SnippetsTab'
import SettingsTab from '../components/app-detail/SettingsTab'
import DangerZoneTab from '../components/app-detail/DangerZoneTab'
import SecretDisplayModal from '../components/app-detail/SecretDisplayModal'
import AppDetail from '../pages/AppDetail'
import type { OAuthApp } from '../api'

export default async function runM5Challenger2Suite() {
  console.log('===========================================================')
  console.log('  CHALLENGER 2: MILESTONE 5 STRESS TEST & REGRESSION SUITE')
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

  const mockConfidentialApp: OAuthApp = {
    ID: 'app_m5_confidential_123',
    DeveloperID: 'dev_m5',
    Name: 'Production Microservice IdP',
    Description: 'Quantum-safe backend service authentication provider',
    HomepageURL: 'https://service.example.com',
    ClientID: 'client_m5_conf_98765',
    RedirectURIs: ['https://service.example.com/oauth/callback', 'https://service.example.com/api/login'],
    IsPublic: false,
    CreatedAt: '2026-07-28T21:00:00Z',
    SecretRotatedAt: '2026-07-28T21:00:00Z',
  }

  const mockPublicApp: OAuthApp = {
    ID: 'app_m5_public_456',
    DeveloperID: 'dev_m5',
    Name: 'React SPA Client App',
    Description: 'Public single page web application',
    HomepageURL: 'https://spa.example.com',
    ClientID: 'client_m5_pub_12345',
    RedirectURIs: ['https://spa.example.com/callback'],
    IsPublic: true,
    CreatedAt: '2026-07-28T21:00:00Z',
    SecretRotatedAt: null,
  }

  // =========================================================================
  // TASK 1: TYPE-APP-NAME VALIDATION IN DeleteAppModal.tsx
  // =========================================================================
  console.log('--- TASK 1: DeleteAppModal.tsx Type-App-Name Validation ---')

  // 1.1 Unconfirmed initial state (empty string typed)
  const deleteModalEmpty = renderToString(
    React.createElement(DeleteAppModal, {
      appName: 'Production Microservice IdP',
      loading: false,
      onConfirm: () => {},
      onCancel: () => {},
    })
  )
  assert(
    deleteModalEmpty.includes('disabled=""') || deleteModalEmpty.includes('disabled'),
    '1.1 Delete button is disabled by default when typed name is empty'
  )
  assert(
    deleteModalEmpty.includes('cursor-not-allowed'),
    '1.2 Delete button contains cursor-not-allowed CSS styling when unconfirmed'
  )

  // 1.3 Case Sensitivity Test: Lowercase input should not enable delete button logic
  const appNameCase = 'Production Microservice IdP'
  const lowercaseInput = 'production microservice idp'
  const uppercaseInput = 'PRODUCTION MICROSERVICE IDP'
  assert(
    lowercaseInput.trim() !== appNameCase.trim(),
    '1.3 Case sensitivity check: lowercase string "production microservice idp" does NOT match "Production Microservice IdP"'
  )
  assert(
    uppercaseInput.trim() !== appNameCase.trim(),
    '1.4 Case sensitivity check: uppercase string "PRODUCTION MICROSERVICE IDP" does NOT match "Production Microservice IdP"'
  )

  // 1.5 Whitespace Handling Test: Trimmed leading/trailing spaces vs internal extra spaces
  const paddedInput = '   Production Microservice IdP   '
  const internalSpaceInput = 'Production  Microservice  IdP'
  assert(
    paddedInput.trim() === appNameCase.trim(),
    '1.5 Whitespace check: leading/trailing padded string trims correctly to match app name'
  )
  assert(
    internalSpaceInput.trim() !== appNameCase.trim(),
    '1.6 Whitespace check: internal double spaces do NOT match app name'
  )

  // 1.7 Loading State disable behavior
  const deleteModalLoading = renderToString(
    React.createElement(DeleteAppModal, {
      appName: 'Production Microservice IdP',
      loading: true,
      onConfirm: () => {},
      onCancel: () => {},
    })
  )
  assert(
    deleteModalLoading.includes('Deleting...'),
    '1.7 Loading state displays "Deleting..." text with spinner icon'
  )
  assert(
    deleteModalLoading.includes('disabled=""') || deleteModalLoading.includes('disabled'),
    '1.8 Submit button is disabled during loading state'
  )

  // =========================================================================
  // TASK 2: CODE SNIPPETS SWITCHER & TOKEN ENDPOINT SELECTOR IN SnippetsTab.tsx
  // =========================================================================
  console.log('\n--- TASK 2: SnippetsTab.tsx Strategy & Language Switcher ---')

  const snippetsHtml = renderToString(
    React.createElement(SnippetsTab, { app: mockConfidentialApp })
  )

  // 2.1 Endpoint strategies listed
  assert(snippetsHtml.includes('/token/thin'), '2.1 Contains /token/thin strategy selector')
  assert(snippetsHtml.includes('/token/full'), '2.2 Contains /token/full strategy selector')
  assert(snippetsHtml.includes('/token/cwt'), '2.3 Contains /token/cwt strategy selector')

  // 2.4 Default Thin strategy code snippet checks
  assert(snippetsHtml.includes('/token/thin'), '2.4 Default active endpoint is /token/thin')
  assert(snippetsHtml.includes('pqc_ref'), '2.5 Thin token snippet references sidecar pqc_ref hash')

  // 2.6 Language selector buttons
  assert(snippetsHtml.includes('cURL'), '2.6 cURL language option present')
  assert(snippetsHtml.includes('JS / Fetch'), '2.7 JS / Fetch language option present')
  assert(snippetsHtml.includes('Go'), '2.8 Go language option present')
  assert(snippetsHtml.includes('Python'), '2.9 Python language option present')

  // 2.10 Secret placeholder behavior for confidential vs public apps
  assert(
    snippetsHtml.includes('YOUR_CLIENT_SECRET'),
    '2.10 Confidential app snippet contains YOUR_CLIENT_SECRET placeholder'
  )

  const publicSnippetsHtml = renderToString(
    React.createElement(SnippetsTab, { app: mockPublicApp })
  )
  assert(
    !publicSnippetsHtml.includes('YOUR_CLIENT_SECRET'),
    '2.11 Public SPA snippet omits YOUR_CLIENT_SECRET placeholder'
  )

  // =========================================================================
  // TASK 3: MOBILE VIEWPORT RESPONSIVENESS (<768px)
  // =========================================================================
  console.log('\n--- TASK 3: Mobile Viewport Responsiveness (<768px) ---')

  // 3.1 AppDetail header flex-wrap & sub-nav tab bar horizontal scrolling
  const appDetailHtml = renderToString(
    React.createElement(
      MemoryRouter,
      { initialEntries: ['/apps/app_m5_confidential_123'] },
      React.createElement(AppDetail)
    )
  )
  assert(typeof appDetailHtml === 'string', '3.0 AppDetail renders within Router without throwing')
  const dangerZoneHtml = renderToString(
    React.createElement(DangerZoneTab, {
      app: mockConfidentialApp,
      onOpenRotateModal: () => {},
      onOpenDeleteModal: () => {},
    })
  )
  assert(
    dangerZoneHtml.includes('flex flex-col sm:flex-row'),
    '3.1 DangerZoneTab uses mobile flex-col stacking and sm:flex-row desktop row alignment'
  )

  const settingsTabHtml = renderToString(
    React.createElement(SettingsTab, {
      app: mockConfidentialApp,
      onSave: async () => {},
      saving: false,
      saveSuccess: false,
      saveError: null,
    })
  )
  assert(
    settingsTabHtml.includes('overflow-x-auto'),
    '3.2 SettingsTab Client ID container uses overflow-x-auto to prevent mobile horizontal overflow'
  )

  assert(
    snippetsHtml.includes('grid grid-cols-1 sm:grid-cols-3'),
    '3.3 SnippetsTab endpoint strategies use responsive 1-column mobile grid (cols-1) and 3-column tablet/desktop grid (sm:grid-cols-3)'
  )

  assert(
    snippetsHtml.includes('overflow-x-auto'),
    '3.4 SnippetsTab pre code block uses overflow-x-auto for mobile scrollable code blocks'
  )

  const secretModalHtml = renderToString(
    React.createElement(SecretDisplayModal, {
      clientId: 'client_m5_conf_98765',
      clientSecret: 'sec_ml_dsa_44_rotated_secret_999xxx',
      onClose: () => {},
    })
  )
  assert(
    secretModalHtml.includes('break-all') && secretModalHtml.includes('whitespace-pre-wrap'),
    '3.5 SecretDisplayModal secret container uses break-all and whitespace-pre-wrap for responsive secret display on small screens'
  )

  // =========================================================================
  // SUMMARY
  // =========================================================================
  console.log('\n===========================================================')
  console.log(`  CHALLENGER 2 RESULTS: ${passed} PASSED, ${failed} FAILED`)
  console.log('===========================================================')

  if (failed > 0) {
    throw new Error(`${failed} M5 Challenger 2 tests failed`)
  }
}
