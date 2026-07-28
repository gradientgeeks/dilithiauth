import React from 'react'
import { renderToString } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import WizardStepper from '../components/new-app/WizardStepper'
import TypeSelector from '../components/new-app/TypeSelector'
import SecretSuccessModal from '../components/new-app/SecretSuccessModal'
import NewApp from '../pages/NewApp'
import type { OAuthApp } from '../api'

export default async function runM4TestSuite() {
  console.log('===========================================================')
  console.log('  EMPIRICAL VERIFICATION: MILESTONE 4 NEW APP WIZARD')
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

  // --- SECTION 1: WIZARD STEPPER COMPONENT ---
  console.log('--- SECTION 1: WizardStepper Component ---')
  const steps = [
    { id: 1, title: 'Basic Details', subtitle: 'Name & Description' },
    { id: 2, title: 'App Type', subtitle: 'Confidential vs Public' },
    { id: 3, title: 'URIs & Config', subtitle: 'Redirect URIs & Finalize' },
  ]

  const step1Html = renderToString(
    React.createElement(WizardStepper, { currentStep: 1, steps })
  )
  assert(step1Html.includes('Basic Details'), '1.1 Renders Step 1 title')
  assert(step1Html.includes('App Type'), '1.2 Renders Step 2 title')
  assert(step1Html.includes('URIs &amp; Config') || step1Html.includes('URIs & Config'), '1.3 Renders Step 3 title')

  const step2Html = renderToString(
    React.createElement(WizardStepper, { currentStep: 2, steps })
  )
  assert(step2Html.includes('bg-emerald-500'), '1.4 Completed Step 1 displays completed indicator')

  // --- SECTION 2: TYPE SELECTOR COMPONENT ---
  console.log('\n--- SECTION 2: TypeSelector Component ---')
  const confidentialTypeHtml = renderToString(
    React.createElement(TypeSelector, { isPublic: false, onChange: () => {} })
  )
  assert(confidentialTypeHtml.includes('Confidential Server-Side App'), '2.1 Renders Confidential option title')
  assert(confidentialTypeHtml.includes('Client Secret Flow'), '2.2 Renders Client Secret Flow badge')
  assert(confidentialTypeHtml.includes('Public SPA / Native App'), '2.3 Renders Public option title')
  assert(confidentialTypeHtml.includes('PKCE Authorization Flow'), '2.4 Renders PKCE Authorization Flow badge')

  const publicTypeHtml = renderToString(
    React.createElement(TypeSelector, { isPublic: true, onChange: () => {} })
  )
  assert(publicTypeHtml.includes('border-[var(--color-cyan)]') || publicTypeHtml.includes('bg-cyan-500'), '2.5 Displays active cyan selection styling for Public Client')

  // --- SECTION 3: SECRET SUCCESS MODAL COMPONENT ---
  console.log('\n--- SECTION 3: SecretSuccessModal Component ---')
  const mockConfidentialApp: OAuthApp = {
    ID: 'app_m4_1',
    DeveloperID: 'dev_m4',
    Name: 'Payment Backend API',
    Description: 'Confidential client application',
    HomepageURL: 'https://pay.example.com',
    ClientID: 'client_m4_confidential_123',
    RedirectURIs: ['https://pay.example.com/callback'],
    IsPublic: false,
    CreatedAt: '2026-07-28T20:00:00Z',
    SecretRotatedAt: null,
  }

  const confidentialModalHtml = renderToString(
    React.createElement(SecretSuccessModal, {
      app: mockConfidentialApp,
      secret: 'sec_quantum_dilithium_44_key_xyz987',
      onNavigateToApp: () => {},
      onNavigateToDashboard: () => {},
    })
  )

  assert(confidentialModalHtml.includes('Application Created Successfully!'), '3.1 Renders success modal heading')
  assert(confidentialModalHtml.includes('Payment Backend API'), '3.2 Renders created app name')
  assert(confidentialModalHtml.includes('client_m4_confidential_123'), '3.3 Displays Client ID')
  assert(confidentialModalHtml.includes('sec_quantum_dilithium_44_key_xyz987'), '3.4 Displays Client Secret')
  assert(confidentialModalHtml.includes('Save your Client Secret now'), '3.5 Displays Amber warning banner for secret')
  assert(confidentialModalHtml.includes('Continue to App Details'), '3.6 Renders "Continue to App Details" CTA')

  const mockPublicApp: OAuthApp = {
    ID: 'app_m4_2',
    DeveloperID: 'dev_m4',
    Name: 'Mobile React Native App',
    Description: 'Public client application',
    HomepageURL: 'https://mobile.example.com',
    ClientID: 'client_m4_public_456',
    RedirectURIs: ['https://mobile.example.com/oauth'],
    IsPublic: true,
    CreatedAt: '2026-07-28T20:00:00Z',
    SecretRotatedAt: null,
  }

  const publicModalHtml = renderToString(
    React.createElement(SecretSuccessModal, {
      app: mockPublicApp,
      onNavigateToApp: () => {},
      onNavigateToDashboard: () => {},
    })
  )

  assert(publicModalHtml.includes('Registered as a') && publicModalHtml.includes('Public Client (PKCE mode)'), '3.7 Renders PKCE public client info banner')
  assert(!publicModalHtml.includes('Save your Client Secret now'), '3.8 Does NOT render secret warning for public app without secret')

  // --- SECTION 4: NEW APP PAGE WIZARD ---
  console.log('\n--- SECTION 4: NewApp Page Component ---')
  const newAppHtml = renderToString(
    React.createElement(
      MemoryRouter,
      null,
      React.createElement(NewApp)
    )
  )

  assert(newAppHtml.includes('Register New Application'), '4.1 Renders main heading')
  assert(newAppHtml.includes('Application Name'), '4.2 Renders Step 1 Application Name input label')
  assert(newAppHtml.includes('Homepage URL'), '4.3 Renders Step 1 Homepage URL label')
  assert(newAppHtml.includes('Next: Application Type'), '4.4 Renders Step 1 Next navigation CTA')
  assert(!newAppHtml.toLowerCase().includes('nav-logo') && !newAppHtml.includes('DilithiAuth Identity Provider'), '4.5 Contains ZERO Navbar component rendering')

  // --- SUMMARY ---
  console.log('\n===========================================================')
  console.log(`  VERIFICATION RESULTS: ${passed} PASSED, ${failed} FAILED`)
  console.log('===========================================================')

  if (failed > 0) {
    throw new Error(`${failed} M4 verification tests failed`)
  }
}
