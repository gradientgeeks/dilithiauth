import React from 'react'
import { renderToString } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import Navbar from '../components/Navbar'
import App from '../App'
import { AuthProvider } from '../AuthContext'

import Docs from '../pages/Docs'

declare global {
  var __MOCK_DEVELOPER__: any
}

export default async function runTestSuite() {
  console.log('===========================================================')
  console.log('  EMPIRICAL VERIFICATION: ROUTE NAVIGATION & NAVBAR BEHAVIOR')
  console.log('===========================================================\n')

  let passed = 0
  let failed = 0

  function assert(condition: boolean, testName: string, detail: string = '') {
    if (condition) {
      console.log(`[PASS] ${testName}`)
      passed++
    } else {
      console.error(`[FAIL] ${testName} - ${detail}`)
      if (detail) console.error(`       Detail: ${detail}`)
      failed++
    }
  }

  // --- SECTION 1: ROUTE STRUCTURE & CODE INSPECTION ---
  console.log('--- SECTION 1: Route Structure & Auth Guard Inspection ---')

  assert(typeof App === 'function', '1.1 App component is exported as a function component')
  assert(typeof Docs === 'function', '1.2 Docs component is exported as a function component')
  assert(typeof Navbar === 'function', '1.3 Navbar component is exported as a function component')

  // --- SECTION 2: GUEST ACCESS & NAVBAR HIGHLIGHTING ON /docs ---
  console.log('\n--- SECTION 2: Guest User Access & Navbar Active Highlighting on /docs ---')

  // Set Guest State
  globalThis.__MOCK_DEVELOPER__ = null

  const guestNavbarHtml = renderToString(
    React.createElement(
      AuthProvider,
      null,
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/docs'] },
        React.createElement(Navbar)
      )
    )
  )

  const hasActiveDocsGuest = guestNavbarHtml.includes('class="active"') && guestNavbarHtml.includes('SDK &amp; Docs')
  assert(hasActiveDocsGuest, '2.1 Navbar highlights "/docs" link with class="active" for Guest user')

  const hasGuestLinks = guestNavbarHtml.includes('Overview') &&
                        guestNavbarHtml.includes('Sign in') &&
                        guestNavbarHtml.includes('Get Started')
  assert(hasGuestLinks, '2.2 Navbar displays Guest links (Overview, Sign in, Get Started)')

  const hasAuthLinksInGuest = guestNavbarHtml.includes('Applications') || guestNavbarHtml.includes('Sign out')
  assert(!hasAuthLinksInGuest, '2.3 Navbar does NOT display Authenticated links for Guest user')

  const docsPageGuestHtml = renderToString(
    React.createElement(
      AuthProvider,
      null,
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/docs'] },
        React.createElement(Docs)
      )
    )
  )

  const rendersDocsContentGuest = docsPageGuestHtml.includes('DilithiAuth SDK &amp; Interactive Documentation') ||
                                   docsPageGuestHtml.includes('DilithiAuth SDK')
  assert(rendersDocsContentGuest, '2.4 /docs route renders Docs page content for Guest user (accessible without auth)')


  // --- SECTION 3: ACTIVE CLASS HIGHLIGHTING LOGIC STRESS TESTING ---
  console.log('\n--- SECTION 3: Active Class Highlighting Logic Stress Testing ---')

  // Test route 1: /docs (Exact match)
  const navAtDocs = renderToString(
    React.createElement(
      AuthProvider,
      null,
      React.createElement(MemoryRouter, { initialEntries: ['/docs'] }, React.createElement(Navbar))
    )
  )
  const docsLinkIsActive = navAtDocs.includes('href="/docs" class="active"') ||
                           navAtDocs.includes('class="active" href="/docs"')
  assert(docsLinkIsActive, '3.1 Navbar /docs link has active class when current pathname === "/docs"')

  // Test route 2: / (Landing route)
  const navAtRoot = renderToString(
    React.createElement(
      AuthProvider,
      null,
      React.createElement(MemoryRouter, { initialEntries: ['/'] }, React.createElement(Navbar))
    )
  )
  const overviewLinkIsActive = navAtRoot.includes('href="/" class="active"') ||
                              navAtRoot.includes('class="active" href="/"')
  const docsActiveAtRoot = navAtRoot.includes('href="/docs" class="active"') ||
                           navAtRoot.includes('class="active" href="/docs"')
  assert(overviewLinkIsActive, '3.2 Navbar "/" link has active class when current pathname === "/"')
  assert(!docsActiveAtRoot, '3.3 Navbar /docs link does NOT have active class when current pathname === "/"')

  // Test route 3: /login
  const navAtLogin = renderToString(
    React.createElement(
      AuthProvider,
      null,
      React.createElement(MemoryRouter, { initialEntries: ['/login'] }, React.createElement(Navbar))
    )
  )
  const loginLinkIsActive = navAtLogin.includes('href="/login" class="active"') ||
                            navAtLogin.includes('class="active" href="/login"')
  const docsActiveAtLogin = navAtLogin.includes('href="/docs" class="active"') ||
                            navAtLogin.includes('class="active" href="/docs"')
  assert(loginLinkIsActive, '3.4 Navbar /login link has active class when current pathname === "/login"')
  assert(!docsActiveAtLogin, '3.5 Navbar /docs link does NOT have active class when current pathname === "/login"')


  // --- SECTION 4: AUTHENTICATED USER ROUTE ACCESSIBILITY & HIGHLIGHTING ---
  console.log('\n--- SECTION 4: Authenticated User Access & Navbar Active Highlighting ---')

  // Set Authenticated State
  globalThis.__MOCK_DEVELOPER__ = {
    id: 'dev_999',
    email: 'alice.quantum@example.com',
    name: 'Alice Quantum'
  }

  // Test 4.1: Render Navbar at /docs for Authenticated User
  const authNavAtDocs = renderToString(
    React.createElement(
      AuthProvider,
      null,
      React.createElement(MemoryRouter, { initialEntries: ['/docs'] }, React.createElement(Navbar))
    )
  )

  const hasAuthDocsActive = authNavAtDocs.includes('href="/docs" class="active"') ||
                            authNavAtDocs.includes('class="active" href="/docs"')
  assert(hasAuthDocsActive, '4.1 Navbar highlights "/docs" with class="active" for Authenticated user at /docs')

  const hasUserEmail = authNavAtDocs.includes('alice.quantum@example.com')
  assert(hasUserEmail, '4.2 Navbar renders user email pill (alice.quantum@example.com) for Authenticated user')

  const hasApplicationsLink = authNavAtDocs.includes('Applications')
  assert(hasApplicationsLink, '4.3 Navbar renders "Applications" link for Authenticated user')

  const hasSignOutButton = authNavAtDocs.includes('Sign out')
  assert(hasSignOutButton, '4.4 Navbar renders "Sign out" button for Authenticated user')

  const hasGuestOverviewLinkInAuth = authNavAtDocs.includes('Overview') || authNavAtDocs.includes('Sign in')
  assert(!hasGuestOverviewLinkInAuth, '4.5 Navbar does NOT render guest links (Overview, Sign in) for Authenticated user')

  // Test 4.2: Render Navbar at /apps for Authenticated User
  const authNavAtApps = renderToString(
    React.createElement(
      AuthProvider,
      null,
      React.createElement(MemoryRouter, { initialEntries: ['/apps'] }, React.createElement(Navbar))
    )
  )

  const hasAuthAppsActive = authNavAtApps.includes('href="/apps" class="active"') ||
                            authNavAtApps.includes('class="active" href="/apps"')
  assert(hasAuthAppsActive, '4.6 Navbar highlights "/apps" with class="active" for Authenticated user at /apps')

  const docsActiveAtApps = authNavAtApps.includes('href="/docs" class="active"') ||
                           authNavAtApps.includes('class="active" href="/docs"')
  assert(!docsActiveAtApps, '4.7 Navbar /docs link is NOT active when Authenticated user is at /apps')

  // Test 4.3: Render Docs page for Authenticated User
  const docsPageAuthHtml = renderToString(
    React.createElement(
      AuthProvider,
      null,
      React.createElement(MemoryRouter, { initialEntries: ['/docs'] }, React.createElement(Docs))
    )
  )

  const rendersDocsContentAuth = docsPageAuthHtml.includes('DilithiAuth SDK &amp; Interactive Documentation') ||
                                  docsPageAuthHtml.includes('DilithiAuth SDK')
  assert(rendersDocsContentAuth, '4.8 /docs route renders Docs page content for Authenticated user without redirect')

  // --- SUMMARY ---
  console.log('\n===========================================================')
  console.log(`  VERIFICATION RESULTS: ${passed} PASSED, ${failed} FAILED`)
  console.log('===========================================================')

  if (failed > 0) {
    throw new Error(`${failed} verification tests failed`)
  }
}
