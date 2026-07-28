import React from 'react'
import { renderToString } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import AppCard, { AppCard as AppCardNamed, formatClientId } from '../components/dashboard/AppCard'
import AppSearchFilter, { AppSearchFilter as AppSearchFilterNamed, type ClientTypeFilter } from '../components/dashboard/AppSearchFilter'
import AppEmptyState, { AppEmptyState as AppEmptyStateNamed } from '../components/dashboard/AppEmptyState'
import Dashboard from '../pages/Dashboard'
import type { OAuthApp } from '../api'

export default async function runM3ChallengeSuite() {
  console.log('===========================================================')
  console.log('  ADVERSARIAL CHALLENGE SUITE: MILESTONE 3 (R2 APPS PAGE)')
  console.log('===========================================================\n')

  let totalTests = 0
  let passed = 0
  let failed = 0

  function testResult(name: string, isPass: boolean, detail: string = '') {
    totalTests++
    if (isPass) {
      console.log(`[PASS] Test ${totalTests}: ${name}`)
      passed++
    } else {
      console.log(`[FAIL] Test ${totalTests}: ${name}`)
      if (detail) console.log(`       Detail: ${detail}`)
      failed++
    }
  }

  // --------------------------------------------------------------------------
  // TEST GROUP 1: Component Exports and Types
  // --------------------------------------------------------------------------
  console.log('--- GROUP 1: Component Exports and Module Structure ---')

  testResult(
    'AppCard module provides default and named export',
    typeof AppCard === 'function' && typeof AppCardNamed === 'function' && AppCard === AppCardNamed,
    'AppCard default and named exports must match'
  )

  testResult(
    'AppSearchFilter module provides default and named export',
    typeof AppSearchFilter === 'function' && typeof AppSearchFilterNamed === 'function' && AppSearchFilter === AppSearchFilterNamed,
    'AppSearchFilter default and named exports must match'
  )

  testResult(
    'AppEmptyState module provides default and named export',
    typeof AppEmptyState === 'function' && typeof AppEmptyStateNamed === 'function' && AppEmptyState === AppEmptyStateNamed,
    'AppEmptyState default and named exports must match'
  )

  testResult(
    'Dashboard page module provides default export',
    typeof Dashboard === 'function',
    'Dashboard default export must be a React component function'
  )

  testResult(
    'formatClientId helper function is exported',
    typeof formatClientId === 'function',
    'formatClientId must be exported from AppCard'
  )

  // --------------------------------------------------------------------------
  // TEST GROUP 2: formatClientId Unit & Edge Case Tests
  // --------------------------------------------------------------------------
  console.log('\n--- GROUP 2: formatClientId Helper Function ---')

  testResult(
    'formatClientId("") returns empty string',
    formatClientId('') === '',
    `Received: "${formatClientId('')}"`
  )

  testResult(
    'formatClientId handles short Client ID (<=16 chars) without truncating',
    formatClientId('client_12345') === 'client_12345',
    `Received: "${formatClientId('client_12345')}"`
  )

  testResult(
    'formatClientId handles exact 16-character Client ID without truncating',
    formatClientId('1234567890123456') === '1234567890123456',
    `Received: "${formatClientId('1234567890123456')}"`
  )

  testResult(
    'formatClientId truncates Client ID > 16 chars to first 11 chars + "..."',
    formatClientId('12345678901234567') === '12345678901...',
    `Received: "${formatClientId('12345678901234567')}"`
  )

  testResult(
    'formatClientId truncates standard UUID (36 chars) correctly',
    formatClientId('f47ac10b-58cc-4372-a567-0e02b2c3d479') === 'f47ac10b-58...',
    `Received: "${formatClientId('f47ac10b-58cc-4372-a567-0e02b2c3d479')}"`
  )

  // Edge case check: null or undefined input handling
  let safeNullCheck = true
  try {
    const res = formatClientId(null as any)
    if (res !== '') safeNullCheck = false
  } catch (e) {
    safeNullCheck = false
  }
  testResult(
    'formatClientId safely handles null input without throwing exception',
    safeNullCheck,
    'BUG DETECTED: formatClientId(null) throws exception if null is passed'
  )

  // --------------------------------------------------------------------------
  // TEST GROUP 3: Search and Type Filter Logic (Pure Unit Logic)
  // --------------------------------------------------------------------------
  console.log('\n--- GROUP 3: Search and Type Filter Matching Logic ---')

  const sampleApps: OAuthApp[] = [
    {
      ID: '1',
      DeveloperID: 'dev1',
      ClientID: 'pub_client_001_web_app',
      Name: 'Portal Web App',
      Description: 'Customer facing SPA frontend',
      HomepageURL: 'https://app.example.com',
      RedirectURIs: ['http://localhost:3000/callback'],
      IsPublic: true,
      CreatedAt: '2026-01-10T12:00:00Z',
      SecretRotatedAt: null,
    },
    {
      ID: '2',
      DeveloperID: 'dev2',
      ClientID: 'conf_client_002_backend',
      Name: 'Backend Service',
      Description: 'Node.js confidential API client',
      HomepageURL: 'https://backend.example.com',
      RedirectURIs: ['http://api.internal/oauth/callback'],
      IsPublic: false,
      CreatedAt: '2026-02-01T15:30:00Z',
      SecretRotatedAt: null,
    },
    {
      ID: '3',
      DeveloperID: 'dev3',
      ClientID: 'pub_client_003_mobile',
      Name: 'Mobile Native App',
      Description: 'iOS and Android client',
      HomepageURL: 'https://mobile.example.com',
      RedirectURIs: ['myapp://oauth/callback'],
      IsPublic: true,
      CreatedAt: '2026-03-01T09:00:00Z',
      SecretRotatedAt: null,
    },
  ]

  const filterFn = (appList: OAuthApp[], search: string, typeFilter: ClientTypeFilter) => {
    return appList.filter((app) => {
      const matchesSearch =
        app.Name.toLowerCase().includes(search.toLowerCase()) ||
        app.ClientID.toLowerCase().includes(search.toLowerCase())

      const matchesType =
        typeFilter === 'all'
          ? true
          : typeFilter === 'public'
          ? app.IsPublic
          : !app.IsPublic

      return matchesSearch && matchesType
    })
  }

  testResult(
    'Filter with search="" and typeFilter="all" returns all apps',
    filterFn(sampleApps, '', 'all').length === 3,
    'Expected 3 apps'
  )

  testResult(
    'Search by name case-insensitively ("portal" -> 1 app)',
    filterFn(sampleApps, 'portal', 'all').length === 1 && filterFn(sampleApps, 'portal', 'all')[0].ID === '1',
    'Expected Portal Web App'
  )

  testResult(
    'Search by Client ID ("conf_client" -> 1 app)',
    filterFn(sampleApps, 'conf_client', 'all').length === 1 && filterFn(sampleApps, 'conf_client', 'all')[0].ID === '2',
    'Expected Backend Service'
  )

  testResult(
    'Filter by type "public" returns only Public SPAs (2 apps)',
    filterFn(sampleApps, '', 'public').length === 2 && filterFn(sampleApps, '', 'public').every((a) => a.IsPublic),
    'Expected 2 public apps'
  )

  testResult(
    'Filter by type "confidential" returns only Confidential Servers (1 app)',
    filterFn(sampleApps, '', 'confidential').length === 1 && !filterFn(sampleApps, '', 'confidential')[0].IsPublic,
    'Expected 1 confidential app'
  )

  testResult(
    'Combined search ("app") + typeFilter ("public") returns matching public app',
    filterFn(sampleApps, 'app', 'public').length === 2,
    'Expected 2 public apps containing "app"'
  )

  testResult(
    'Combined search ("backend") + typeFilter ("public") returns 0 apps',
    filterFn(sampleApps, 'backend', 'public').length === 0,
    'Expected 0 apps'
  )

  testResult(
    'Search query with special characters ("<script>") does not crash filter logic',
    filterFn(sampleApps, '<script>', 'all').length === 0,
    'Special characters should return 0 results without throwing error'
  )

  // Adversarial check: Null/undefined ClientID or Name in app record
  const badApp: OAuthApp = {
    ID: '4',
    DeveloperID: 'dev4',
    ClientID: undefined as any,
    Name: 'Broken App',
    Description: '',
    HomepageURL: '',
    RedirectURIs: [],
    IsPublic: true,
    CreatedAt: '2026-01-01',
    SecretRotatedAt: null,
  }
  let safeBadAppSearch = true
  try {
    filterFn([badApp], 'test', 'all')
  } catch (e) {
    safeBadAppSearch = false
  }
  testResult(
    'Search logic handles missing/undefined ClientID without crashing Dashboard',
    safeBadAppSearch,
    'BUG DETECTED: If app.ClientID is undefined, app.ClientID.toLowerCase() throws TypeError in Dashboard filter logic!'
  )

  // --------------------------------------------------------------------------
  // TEST GROUP 4: AppCard SSR Component Rendering & Edge Cases
  // --------------------------------------------------------------------------
  console.log('\n--- GROUP 4: AppCard Rendering & Visual Edge Cases ---')

  const publicApp = sampleApps[0]
  const confApp = sampleApps[1]

  const htmlPublicCard = renderToString(
    React.createElement(MemoryRouter, null,
      React.createElement(AppCard, {
        app: publicApp,
        copiedId: null,
        onCopyClientId: () => {},
      })
    )
  )

  testResult(
    'Public SPA AppCard renders "Public SPA" badge text',
    htmlPublicCard.includes('Public SPA'),
    'Badge text must be rendered'
  )

  testResult(
    'Public SPA AppCard applies cyan color styling classes',
    htmlPublicCard.includes('bg-cyan-500/10') && htmlPublicCard.includes('text-cyan-400'),
    'Cyan badge styling must be applied'
  )

  testResult(
    'AppCard renders masked Client ID',
    htmlPublicCard.includes(formatClientId(publicApp.ClientID)),
    'Masked Client ID must be rendered in <code> element'
  )

  testResult(
    'AppCard renders full Client ID in title attribute for accessibility hover',
    htmlPublicCard.includes(`title="${publicApp.ClientID}"`),
    'title attribute must contain full Client ID'
  )

  const htmlConfCard = renderToString(
    React.createElement(MemoryRouter, null,
      React.createElement(AppCard, {
        app: confApp,
        copiedId: null,
        onCopyClientId: () => {},
      })
    )
  )

  testResult(
    'Confidential Server AppCard renders "Confidential Server" badge text',
    htmlConfCard.includes('Confidential Server'),
    'Badge text must be rendered'
  )

  testResult(
    'Confidential Server AppCard applies purple color styling classes',
    htmlConfCard.includes('bg-purple-500/10') && htmlConfCard.includes('text-purple-400'),
    'Purple badge styling must be applied'
  )

  // Test copied state rendering
  const htmlCopiedCard = renderToString(
    React.createElement(MemoryRouter, null,
      React.createElement(AppCard, {
        app: publicApp,
        copiedId: publicApp.ClientID,
        onCopyClientId: () => {},
      })
    )
  )

  testResult(
    'AppCard renders "Copied!" title tooltip when copiedId matches app.ClientID',
    htmlCopiedCard.includes('title="Copied!"'),
    'Copied tooltip must be rendered when copiedId matches'
  )

  testResult(
    'AppCard applies emerald styling when copiedId matches app.ClientID',
    htmlCopiedCard.includes('text-emerald-400'),
    'Emerald success styling must be rendered'
  )

  // Edge case: AppCard with empty description fallback
  const appNoDesc: OAuthApp = { ...publicApp, Description: '' }
  const htmlNoDesc = renderToString(
    React.createElement(MemoryRouter, null,
      React.createElement(AppCard, {
        app: appNoDesc,
        copiedId: null,
        onCopyClientId: () => {},
      })
    )
  )

  testResult(
    'AppCard falls back to "No description provided." when description is empty',
    htmlNoDesc.includes('No description provided.'),
    'Fallback description must render'
  )

  // Edge case: AppCard with invalid date string
  const appInvalidDate: OAuthApp = { ...publicApp, CreatedAt: 'invalid-date-string' }
  const htmlInvalidDate = renderToString(
    React.createElement(MemoryRouter, null,
      React.createElement(AppCard, {
        app: appInvalidDate,
        copiedId: null,
        onCopyClientId: () => {},
      })
    )
  )
  const rendersInvalidDateStr = htmlInvalidDate.includes('Invalid Date')
  testResult(
    'AppCard handles invalid CreatedAt string gracefully without displaying "Invalid Date"',
    !rendersInvalidDateStr,
    'BUG DETECTED: `new Date("invalid-date-string").toLocaleDateString()` produces literal string "Invalid Date" in UI!'
  )

  // Edge case: Emoji / Multi-byte unicode character in app Name initial extraction
  const appEmojiName: OAuthApp = { ...publicApp, Name: '🚀 Rocket Auth' }
  const htmlEmojiCard = renderToString(
    React.createElement(MemoryRouter, null,
      React.createElement(AppCard, {
        app: appEmojiName,
        copiedId: null,
        onCopyClientId: () => {},
      })
    )
  )
  const hasBrokenEmojiSurrogate = htmlEmojiCard.includes('>\uD83D<')
  testResult(
    'AppCard initial extraction handles multi-byte emoji character without producing broken surrogate pair',
    !hasBrokenEmojiSurrogate,
    'BUG DETECTED: `app.Name.charAt(0)` on emoji "🚀" extracts high surrogate unit "\\uD83D", rendering a broken character block!'
  )

  // --------------------------------------------------------------------------
  // TEST GROUP 5: AppSearchFilter Component Rendering
  // --------------------------------------------------------------------------
  console.log('\n--- GROUP 5: AppSearchFilter Component Rendering ---')

  const htmlSearchFilter = renderToString(
    React.createElement(AppSearchFilter, {
      search: 'portal',
      onSearchChange: () => {},
      typeFilter: 'public',
      onTypeFilterChange: () => {},
      showingCount: 1,
      totalCount: 3,
      publicCount: 2,
      confidentialCount: 1,
    })
  )

  testResult(
    'AppSearchFilter renders search input with binding value "portal"',
    htmlSearchFilter.includes('value="portal"'),
    'Search input value attribute must match prop'
  )

  testResult(
    'AppSearchFilter renders search placeholder text',
    htmlSearchFilter.includes('placeholder="Search applications by name or Client ID..."'),
    'Placeholder text must be present'
  )

  testResult(
    'AppSearchFilter renders filter tabs: All (3), Public SPA (2), Confidential Server (1)',
    htmlSearchFilter.includes('All') && htmlSearchFilter.includes('>3<') &&
    htmlSearchFilter.includes('Public SPA') && htmlSearchFilter.includes('>2<') &&
    htmlSearchFilter.includes('Confidential Server') && htmlSearchFilter.includes('>1<'),
    'Tabs and counter badges must be rendered'
  )

  testResult(
    'AppSearchFilter renders showing count badge ("Showing 1 of 3 applications")',
    htmlSearchFilter.includes('Showing') && htmlSearchFilter.includes('>1<') && htmlSearchFilter.includes('>3<'),
    'Showing count badge must be rendered'
  )

  testResult(
    'Filter buttons specify type="button" to prevent unwanted form submission',
    (htmlSearchFilter.match(/type="button"/g) || []).length >= 3,
    'Filter tab buttons must have type="button"'
  )

  // --------------------------------------------------------------------------
  // TEST GROUP 6: AppEmptyState Component Rendering & Condition Logic
  // --------------------------------------------------------------------------
  console.log('\n--- GROUP 6: AppEmptyState Rendering & Conditions ---')

  // Unfiltered empty state
  const htmlUnfilteredEmpty = renderToString(
    React.createElement(MemoryRouter, null,
      React.createElement(AppEmptyState, {
        search: '',
        typeFilter: 'all',
        onClearFilters: () => {},
      })
    )
  )

  testResult(
    'Unfiltered empty state displays heading "No applications registered yet"',
    htmlUnfilteredEmpty.includes('No applications registered yet'),
    'Heading must match unfiltered empty state'
  )

  testResult(
    'Unfiltered empty state does NOT render "Clear Filters" button',
    !htmlUnfilteredEmpty.includes('Clear Filters'),
    'Clear Filters button should not be present when not filtered'
  )

  testResult(
    'Unfiltered empty state renders "Create Application" link to "/apps/new"',
    htmlUnfilteredEmpty.includes('Create Application') && htmlUnfilteredEmpty.includes('href="/apps/new"'),
    'Create Application link must point to /apps/new'
  )

  // Filtered empty state (search)
  const htmlFilteredEmpty = renderToString(
    React.createElement(MemoryRouter, null,
      React.createElement(AppEmptyState, {
        search: 'nonexistent_app',
        typeFilter: 'all',
        onClearFilters: () => {},
      })
    )
  )

  testResult(
    'Filtered empty state displays heading "No matching applications found"',
    htmlFilteredEmpty.includes('No matching applications found'),
    'Heading must match filtered empty state'
  )

  testResult(
    'Filtered empty state DOES render "Clear Filters" button when onClearFilters is provided',
    htmlFilteredEmpty.includes('Clear Filters'),
    'Clear Filters button must be present when filtered'
  )

  // Filtered empty state (typeFilter public)
  const htmlTypeFilteredEmpty = renderToString(
    React.createElement(MemoryRouter, null,
      React.createElement(AppEmptyState, {
        search: '',
        typeFilter: 'public',
        onClearFilters: () => {},
      })
    )
  )

  testResult(
    'Empty state treats non-"all" typeFilter as filtered state',
    htmlTypeFilteredEmpty.includes('No matching applications found') && htmlTypeFilteredEmpty.includes('Clear Filters'),
    'typeFilter !== "all" must trigger filtered empty state'
  )

  // --------------------------------------------------------------------------
  // TEST GROUP 7: Dashboard Page SSR Integration
  // --------------------------------------------------------------------------
  console.log('\n--- GROUP 7: Dashboard Page Layout & Integration ---')

  const htmlDashboard = renderToString(
    React.createElement(MemoryRouter, null,
      React.createElement(Dashboard)
    )
  )

  testResult(
    'Dashboard page renders primary heading "Applications"',
    htmlDashboard.includes('Applications'),
    'Page header title must be present'
  )

  testResult(
    'Dashboard page renders top header "Create Application" button',
    htmlDashboard.includes('Create Application'),
    'Header action button must be present'
  )

  // --------------------------------------------------------------------------
  // SUMMARY
  // --------------------------------------------------------------------------
  console.log('\n===========================================================')
  console.log(`  EMPIRICAL CHALLENGE SUMMARY: ${passed} PASSED, ${failed} FAILED (Out of ${totalTests} checks)`)
  console.log('===========================================================')

  return { totalTests, passed, failed }
}
