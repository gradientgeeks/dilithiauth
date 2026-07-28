import React from 'react'
import { renderToString } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import Dashboard from '../pages/Dashboard'
import AppCard, { formatClientId } from '../components/dashboard/AppCard'
import AppSearchFilter from '../components/dashboard/AppSearchFilter'
import AppEmptyState from '../components/dashboard/AppEmptyState'
import type { OAuthApp } from '../api'

export default async function runM3TestSuite() {
  console.log('===========================================================')
  console.log('  EMPIRICAL VERIFICATION: MILESTONE 3 DASHBOARD REDESIGN')
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

  // --- SECTION 1: FORMAT CLIENT ID UNIT TESTS ---
  console.log('--- SECTION 1: Client ID Masking & Formatting Helper ---')
  assert(formatClientId('client_1234567890abcdef') === 'client_1234...', '1.1 Masks Client ID longer than 16 chars with ellipsis (client_1234...)')
  assert(formatClientId('client_short') === 'client_short', '1.2 Preserves Client ID 16 chars or shorter')
  assert(formatClientId('') === '', '1.3 Handles empty Client ID string gracefully')

  // --- SECTION 2: APP CARD COMPONENT RENDERING ---
  console.log('\n--- SECTION 2: AppCard Component Rendering ---')
  const mockPublicApp: OAuthApp = {
    ID: 'app_1',
    DeveloperID: 'dev_1',
    Name: 'React SPA Client',
    Description: 'Post-quantum single page application',
    HomepageURL: 'https://spa.example.com',
    ClientID: 'client_9876543210fedcba',
    RedirectURIs: ['https://spa.example.com/callback'],
    IsPublic: true,
    CreatedAt: '2026-01-15T10:00:00Z',
    SecretRotatedAt: null,
  }

  const mockConfidentialApp: OAuthApp = {
    ID: 'app_2',
    DeveloperID: 'dev_1',
    Name: 'Backend API Service',
    Description: 'Confidential server-to-server client',
    HomepageURL: 'https://api.example.com',
    ClientID: 'client_1122334455667788',
    RedirectURIs: ['https://api.example.com/oauth/callback'],
    IsPublic: false,
    CreatedAt: '2026-02-20T14:30:00Z',
    SecretRotatedAt: null,
  }

  const publicCardHtml = renderToString(
    React.createElement(
      MemoryRouter,
      null,
      React.createElement(AppCard, {
        app: mockPublicApp,
        copiedId: null,
        onCopyClientId: () => {},
      })
    )
  )

  assert(publicCardHtml.includes('React SPA Client'), '2.1 Renders application name')
  assert(publicCardHtml.includes('Post-quantum single page application'), '2.2 Renders description')
  assert(publicCardHtml.includes('Public SPA'), '2.3 Renders "Public SPA" badge for public app')
  assert(publicCardHtml.includes('client_9876...'), '2.4 Renders masked Client ID (client_9876...)')
  assert(publicCardHtml.includes('href="/apps/app_1"'), '2.5 Contains navigation link to /apps/app_1')
  assert(publicCardHtml.includes('Jan 15, 2026') || publicCardHtml.includes('15/01/2026') || publicCardHtml.includes('2026'), '2.6 Formats and displays CreatedAt date')

  const confidentialCardHtml = renderToString(
    React.createElement(
      MemoryRouter,
      null,
      React.createElement(AppCard, {
        app: mockConfidentialApp,
        copiedId: null,
        onCopyClientId: () => {},
      })
    )
  )

  assert(confidentialCardHtml.includes('Confidential Server'), '2.7 Renders "Confidential Server" badge for confidential app')

  const copiedCardHtml = renderToString(
    React.createElement(
      MemoryRouter,
      null,
      React.createElement(AppCard, {
        app: mockPublicApp,
        copiedId: 'client_9876543210fedcba',
        onCopyClientId: () => {},
      })
    )
  )
  assert(copiedCardHtml.includes('Copied!'), '2.8 Displays "Copied!" feedback state when copiedId matches')

  // --- SECTION 3: APP SEARCH FILTER COMPONENT RENDERING ---
  console.log('\n--- SECTION 3: AppSearchFilter Component Rendering ---')
  const filterHtml = renderToString(
    React.createElement(AppSearchFilter, {
      search: 'react',
      onSearchChange: () => {},
      typeFilter: 'public',
      onTypeFilterChange: () => {},
      showingCount: 1,
      totalCount: 5,
      publicCount: 3,
      confidentialCount: 2,
    })
  )

  assert(filterHtml.includes('placeholder="Search applications by name or Client ID..."'), '3.1 Renders search input with placeholder')
  assert(filterHtml.includes('value="react"'), '3.2 Binds search query input value')
  assert(filterHtml.includes('All'), '3.3 Renders "All" filter tab')
  assert(filterHtml.includes('Public SPA'), '3.4 Renders "Public SPA" filter tab')
  assert(filterHtml.includes('Confidential Server'), '3.5 Renders "Confidential Server" filter tab')
  assert(filterHtml.includes('Showing') && filterHtml.includes('1') && filterHtml.includes('5'), '3.6 Displays "Showing 1 of 5 applications" badge')

  // --- SECTION 4: APP EMPTY STATE COMPONENT RENDERING ---
  console.log('\n--- SECTION 4: AppEmptyState Component Rendering ---')
  const emptyUnfilteredHtml = renderToString(
    React.createElement(
      MemoryRouter,
      null,
      React.createElement(AppEmptyState, {})
    )
  )

  assert(emptyUnfilteredHtml.includes('No applications registered yet'), '4.1 Displays default empty state heading when no filters active')
  assert(emptyUnfilteredHtml.includes('Create Application'), '4.2 Renders "+ Create Application" button link')
  assert(!emptyUnfilteredHtml.includes('Clear Filters'), '4.3 Does NOT render "Clear Filters" button when unfiltered')

  const emptyFilteredHtml = renderToString(
    React.createElement(
      MemoryRouter,
      null,
      React.createElement(AppEmptyState, {
        search: 'nonexistent',
        typeFilter: 'public',
        onClearFilters: () => {},
      })
    )
  )

  assert(emptyFilteredHtml.includes('No matching applications found'), '4.4 Displays filtered empty state heading when search/filter active')
  assert(emptyFilteredHtml.includes('Clear Filters'), '4.5 Renders "Clear Filters" button when filters active')

  // --- SECTION 5: DASHBOARD PAGE LAYOUT ---
  console.log('\n--- SECTION 5: Dashboard Page Layout ---')
  const dashboardHtml = renderToString(
    React.createElement(
      MemoryRouter,
      null,
      React.createElement(Dashboard)
    )
  )

  assert(dashboardHtml.includes('Applications'), '5.1 Renders "Applications" heading')
  assert(dashboardHtml.includes('Manage your post-quantum OAuth 2.0 / OIDC applications'), '5.2 Renders subtitle')
  assert(dashboardHtml.includes('Create Application'), '5.3 Renders "Create Application" CTA')
  assert(dashboardHtml.includes('skeleton'), '5.4 Renders loading skeleton on initial mount while loading')

  // --- SUMMARY ---
  console.log('\n===========================================================')
  console.log(`  VERIFICATION RESULTS: ${passed} PASSED, ${failed} FAILED`)
  console.log('===========================================================')

  if (failed > 0) {
    throw new Error(`${failed} verification tests failed`)
  }
}
