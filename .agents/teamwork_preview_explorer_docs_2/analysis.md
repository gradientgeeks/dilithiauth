# DilithiAuth Docs Page: Overflow & Code Block Protection Analysis (R2 Requirement)

**Target Files**:
- `frontend/src/pages/Docs.tsx`
- `frontend/src/index.css`

**Investigation Date**: 2026-07-28  
**Scope**: Mobile to Tablet/Desktop Viewports (320px to 1023px)  
**Role**: Read-only Explorer (`teamwork_preview_explorer_docs_2`)

---

## 1. Executive Summary

A comprehensive, line-by-line responsive layout and code block audit was conducted for the DilithiAuth Documentation page (`Docs.tsx`) and global stylesheet (`index.css`). The goal was to identify all layout bottlenecks, missing overflow protection rules, hardcoded fixed-pixel styles, inline monospace text breaks, copy-button overlay collisions, and table/code block rendering deficiencies across viewports from **320px to 1023px**.

### Key Findings
1. **Critical Window Horizontal Overflow Bug (Section 3 Grid)**: Line 848 of `Docs.tsx` applies an inline style `style={{ gridTemplateColumns: '260px 1fr', gap: '1.25rem' }}` to `.detail-grid`. Because inline styles override CSS media queries, the desktop 2-column layout (260px left sidebar + 1fr right inspector) is forced down to 320px screen width. This pushes total content width past ~500px, causing severe horizontal scrolling on mobile viewports (320px–767px).
2. **Interactive Parameter Inputs Overlap (Section 3)**: Line 921 of `Docs.tsx` uses inline style `style={{ gridTemplateColumns: '1fr 1fr' }}` for the 4 parameter input fields. On 320px viewports, each field is squeezed into ~120px width, causing text clipping in option labels like `"Thin Token (Ed25519 + pqc_ref)"`.
3. **Unprotected Long Endpoint URLs (Section 3 Inspector Header)**: The endpoint header container (`Docs.tsx:964-995`) is a flex container without `flex-wrap: wrap`. Long URLs (e.g. `/authorize?client_id=...` over 180 chars) lack `word-break: break-all` or `overflow-x: auto`, causing the flex item to expand to ~800px and push the window width out.
4. **Unwrapped CLI Command & Manifest Code Boxes (Section 1)**: In `Docs.tsx:644-716`, installation commands (`mvn dependency:add -Dartifact=com.dilithiauth:dilithiauth-sdk:1.0.0`) are rendered in `<code>` tags inside non-wrapping flex containers without `min-width: 0` or `overflow-x: auto`. On screens < 480px, long command strings (~420px wide) break out of their cards.
5. **Copy Button Overlay Collision (Section 3 Body Tab)**: In `Docs.tsx:1021-1045`, the "Copy Response" button is positioned absolutely (`top: 0.5rem; right: 0.5rem`) inside a relative container over `<pre class="code-block">`. Top-line JSON content (e.g., `"issuer": "https://..."`) renders directly underneath the copy button, obscuring code text and leading to misclicks on narrow screens.
6. **Unwrapped Inline Monospace Text (`.mono`)**: In `index.css:402-410`, `.mono` lacks `word-break: break-word` / `overflow-wrap: anywhere` and `max-width: 100%`. Long cryptographic hashes (e.g., `47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU`) inside list items and cards overflow their parent boxes on 320px viewports.
7. **Missing Global Table & Code Block Protection in `index.css`**: `index.css` lacks responsive table wrapper classes (`.table-container` with `overflow-x: auto`) and standardized code block header patterns.

---

## 2. Detailed Viewport Defect Catalog (320px to 1023px)

### Defect 1: Inline Style Grid Override (`Docs.tsx:848`)
- **Location**: `frontend/src/pages/Docs.tsx`, line 848
- **Current Code**:
  ```tsx
  <div className="detail-grid" style={{ gridTemplateColumns: '260px 1fr', gap: '1.25rem' }}>
  ```
- **Issue**: `index.css` defines `@media (max-width: 640px) { .detail-grid { grid-template-columns: 1fr; } }`. However, specificity rules dictate that inline HTML `style=""` overrides CSS media queries. Consequently, on viewports < 640px (down to 320px), the grid forces a 260px left column and a right column side-by-side.
- **Impact**: Minimum width required is 260px + 20px gap + 220px inspector = 500px. On a 320px screen (with 280px content area inside `.container`), this forces horizontal scrolling for the entire document window.

### Defect 2: Inline Parameter Inputs 2-Column Grid (`Docs.tsx:921`)
- **Location**: `frontend/src/pages/Docs.tsx`, line 921
- **Current Code**:
  ```tsx
  <div className="detail-grid" style={{ gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}>
  ```
- **Issue**: Hardcoded inline `1fr 1fr` grid forces 2 input columns on all screens, ignoring screen width.
- **Impact**: Squeezes inputs on 320px–480px mobile screens. Dropdown values like `"Thin Token (Ed25519 + pqc_ref)"` become unreadable and truncated.

### Defect 3: Non-Wrapping Endpoint URL Header (`Docs.tsx:964-995`)
- **Location**: `frontend/src/pages/Docs.tsx`, line 964
- **Current Code**:
  ```tsx
  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderBottom: '1px solid #1f2937', paddingBottom: '0.5rem', marginBottom: '0.75rem' }}>
    <div className="flex items-center gap-2">
      <span className={`badge ${...}`}>{endpointRes.method}</span>
      <span className="mono" style={{ fontSize: '0.85rem', color: '#f9fafb' }}>
        {endpointRes.url}
      </span>
      <span className="badge" style={{ ... }}>{endpointRes.status} {endpointRes.statusText}</span>
    </div>
    <div className="flex gap-2">...</div>
  </div>
  ```
- **Issue**: 
  1. Header container lacks `flex-wrap: wrap`.
  2. Outer `div` does not allow flex wrapping between the left route info and right tab buttons (`Body`, `Headers`, `Metrics`).
  3. `.mono` URL span containing single long URL string has no `word-break: break-all`, `max-width`, or `overflow-x: auto`.
- **Impact**: Long URLs expand the `.mono` span past 600px width on mobile, pushing right-side tab buttons off-screen and forcing window-level horizontal scroll.

### Defect 4: Unwrapped Installation & Manifest Code Containers (`Docs.tsx:644-716`)
- **Location**: `frontend/src/pages/Docs.tsx`, lines 644–677 & 682–716
- **Current Code**:
  ```tsx
  <div style={{ background: '#060911', border: '1px solid #1f2937', borderRadius: '6px', padding: '0.75rem 1rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '0.5rem', marginTop: '0.3rem' }}>
    <code style={{ fontFamily: 'var(--font-mono)', color: '#38bdf8', fontSize: '0.85rem' }}>
      {langInfo.installCmd}
    </code>
    <button className="btn btn-outline btn-sm" ...>Copy</button>
  </div>
  ```
- **Issue**: The flex row contains a raw `<code>` element and a Copy `<button>`. The `<code>` element lacks `min-width: 0`, `overflow-x: auto`, or `word-break: break-all`.
- **Impact**: On viewports < 480px, long command strings like `mvn dependency:add -Dartifact=com.dilithiauth:dilithiauth-sdk:1.0.0` or `go get github.com/Uttam-Mahata/dilithiauth-go` cannot shrink, pushing the copy button and card right edge off screen.

### Defect 5: Copy Button Collision in Code Block Body Tab (`Docs.tsx:1021-1045`)
- **Location**: `frontend/src/pages/Docs.tsx`, lines 1021–1045
- **Current Code**:
  ```tsx
  <div style={{ position: 'relative' }}>
    <button
      className="btn btn-outline btn-sm"
      onClick={...}
      style={{ position: 'absolute', top: '0.5rem', right: '0.5rem', zIndex: 10 }}
    >
      <Copy size={13} />
      <span>Copy Response</span>
    </button>
    <pre className="code-block" style={{ maxHeight: '340px' }}>
      {JSON.stringify(endpointRes.body, null, 2)}
    </pre>
  </div>
  ```
- **Issue**: The copy button is overlaid directly inside the top-right corner of the `<pre>` element.
- **Impact**: In JSON output, line 1 of formatted JSON (`{\n  "issuer": "https://..."`) renders directly underneath the copy button. On viewports 320px–768px, text lines under the copy button are obscured and unreadable.

### Defect 6: Inline Monospace Hash Overflow (`index.css:402-410` & `Docs.tsx`)
- **Location**: `frontend/src/index.css`, lines 402–410 (`.mono`)
- **Current Code**:
  ```css
  .mono {
    font-family: var(--font-mono);
    font-size: 0.825rem;
    background: #060911;
    border: 1px solid #1f2937;
    color: #60a5fa;
    padding: 0.15rem 0.4rem;
    border-radius: 4px;
  }
  ```
- **Issue**: Missing `word-break: break-all` / `overflow-wrap: anywhere` and `max-width: 100%`.
- **Impact**: Inline `.mono` spans used for SHA-256 PQC reference hashes (e.g. `47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU`) or proxy directives (`large_client_header_buffers 4 16k;`) cause text line breaking failures, forcing card containers to overflow on 320px screens.

### Defect 7: Missing Table Responsive Wrapper in `index.css`
- **Location**: `frontend/src/index.css`, lines 413–439 (`table`)
- **Current Code**: Standard `table`, `th`, `td` styles without container overflow rules.
- **Issue**: Any table rendered in documentation without a scroll wrapper (`overflow-x: auto`) will break mobile layouts.

---

## 3. Code Block Protection & Enterprise Documentation Standards Assessment

To meet enterprise-grade documentation standards (e.g., Stripe, AWS, GitHub Docs), code blocks and table interfaces must satisfy four core rules:

1. **Horizontal Containment Guarantee**: No `<pre>`, `<code>`, `<table>`, or parameter list may ever force the `window` or `body` element to display a horizontal scrollbar. Horizontal scrolling MUST be isolated exclusively to the inner container of the code block or table (`overflow-x: auto`).
2. **Copy Button Non-Obscurity**: Copy buttons MUST reside in a dedicated header bar or top control strip (or have dedicated padding reserved) so code text never renders beneath human interactive controls.
3. **Mobile Readability & Font Scale**: 
   - Monospace font size: `0.8125rem` (13px) or `0.85rem` with `line-height: 1.5–1.6`.
   - Code block padding: `0.85rem 1rem`.
   - Scrollbar styling: thin custom scrollbars (`::-webkit-scrollbar` height/width `6px`) with smooth momentum touch scrolling (`-webkit-overflow-scrolling: touch`).
4. **Responsive Grid Breakpoints**:
   - `< 640px` (Mobile): 1-column stacked layout for API sidebar + inspector, parameter inputs, and code cards.
   - `640px – 1023px` (Tablet): Adaptable single/two column grids with fluid input widths.
   - `≥ 1024px` (Desktop): Full side-by-side split view.

---

## 4. Proposed Remediation Plan & Exact CSS Updates

### A. Updates to `frontend/src/index.css`

Add and refine the following CSS classes and rules in `index.css`:

```css
/* Responsive Grid Classes for Docs & API Inspector */
.api-ref-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.25rem;
}
@media (min-width: 768px) {
  .api-ref-grid {
    grid-template-columns: 260px 1fr;
  }
}

.param-inputs-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;
}
@media (min-width: 640px) {
  .param-inputs-grid {
    grid-template-columns: 1fr 1fr;
  }
}

/* Enhanced Monospace & Code Block Protection */
.mono {
  font-family: var(--font-mono);
  font-size: 0.825rem;
  background: #060911;
  border: 1px solid #1f2937;
  color: #60a5fa;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  word-break: break-word;
  overflow-wrap: anywhere;
  max-width: 100%;
}

.code-block {
  background: #060911;
  color: #e5e7eb;
  font-family: var(--font-mono);
  font-size: 0.825rem;
  padding: 0.85rem 1rem;
  border-radius: var(--radius-sm);
  border: 1px solid #1f2937;
  overflow-x: auto;
  max-width: 100%;
  white-space: pre;
  line-height: 1.5;
  -webkit-overflow-scrolling: touch;
}

/* Dedicated Code Card Header Strip */
.code-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #0b0f19;
  padding: 0.5rem 0.85rem;
  border-top-left-radius: 6px;
  border-top-right-radius: 6px;
  border: 1px solid #1f2937;
  border-bottom: none;
  flex-wrap: wrap;
  gap: 0.5rem;
}

/* Code Snippet / Command Box Container */
.cmd-box {
  background: #060911;
  border: 1px solid #1f2937;
  border-radius: 6px;
  padding: 0.65rem 0.85rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  margin-top: 0.3rem;
  max-width: 100%;
  overflow: hidden;
}

.cmd-box code {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  color: #38bdf8;
  overflow-x: auto;
  white-space: nowrap;
  min-width: 0;
  flex: 1;
  -webkit-overflow-scrolling: touch;
}

/* Responsive Table Wrapper Protection */
.table-container {
  width: 100%;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  margin-bottom: 1rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
}
.table-container table {
  margin-bottom: 0;
}
```

---

### B. Updates to `frontend/src/pages/Docs.tsx`

#### Change 1: Section 1 Installation & Manifest Code Boxes (`Docs.tsx:641-717`)
Replace the rigid inline flex divs with the `.cmd-box` pattern:

```tsx
{/* CLI Package Installation */}
<div>
  <div className="detail-label">CLI Package Installation</div>
  <div className="cmd-box">
    <code>{langInfo.installCmd}</code>
    <button
      className="btn btn-outline btn-sm"
      onClick={() => copyToClipboard(`install-${langInfo.id}`, langInfo.installCmd)}
      style={{ flexShrink: 0 }}
    >
      {copiedMap[`install-${langInfo.id}`] ? (
        <>
          <Check size={14} color="#10b981" />
          <span style={{ color: '#10b981', fontWeight: 600 }}>Copied!</span>
        </>
      ) : (
        <>
          <Copy size={14} />
          <span>Copy</span>
        </>
      )}
    </button>
  </div>
</div>

{/* Manifest File */}
<div>
  <div className="detail-label">Manifest File ({langInfo.manifestFile})</div>
  <div className="cmd-box">
    <code style={{ color: '#9ca3af', whiteSpace: 'pre' }}>
      {langInfo.manifestContent}
    </code>
    <button
      className="btn btn-outline btn-sm"
      onClick={() => copyToClipboard(`manifest-${langInfo.id}`, langInfo.manifestContent)}
      style={{ flexShrink: 0 }}
    >
      {copiedMap[`manifest-${langInfo.id}`] ? (
        <>
          <Check size={14} color="#10b981" />
          <span style={{ color: '#10b981', fontWeight: 600 }}>Copied!</span>
        </>
      ) : (
        <>
          <Copy size={14} />
          <span>Copy</span>
        </>
      )}
    </button>
  </div>
</div>
```

#### Change 2: Section 3 API Reference Grid (`Docs.tsx:848`)
Replace inline `style={{ gridTemplateColumns: '260px 1fr', gap: '1.25rem' }}` with CSS class `.api-ref-grid`:

```tsx
<div className="api-ref-grid">
  {/* Endpoint List Sidebar */}
  ...
  {/* Endpoint Tester & Inspector */}
  ...
</div>
```

#### Change 3: Section 3 Parameter Inputs Grid (`Docs.tsx:921`)
Replace inline `style={{ gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}` with CSS class `.param-inputs-grid`:

```tsx
<div className="param-inputs-grid">
  <div>
    <label className="label">Client ID</label>
    <input ... />
  </div>
  ...
</div>
```

#### Change 4: Section 3 Endpoint Header & Long URL (`Docs.tsx:964-1017`)
Refactor the inspector response header container to support flex-wrapping, text wrapping/truncation, and clear tab button spacing:

```tsx
<div
  style={{
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    borderBottom: '1px solid #1f2937',
    paddingBottom: '0.5rem',
    marginBottom: '0.75rem',
    flexWrap: 'wrap',
    gap: '0.75rem'
  }}
>
  <div className="flex items-center gap-2" style={{ minWidth: 0, flex: 1, flexWrap: 'wrap' }}>
    <span
      className={`badge ${
        endpointRes.method === 'GET' ? 'badge-public' : 'badge-confidential'
      }`}
      style={{ flexShrink: 0 }}
    >
      {endpointRes.method}
    </span>
    <span
      className="mono"
      style={{
        fontSize: '0.85rem',
        color: '#f9fafb',
        wordBreak: 'break-all',
        maxWidth: '100%'
      }}
    >
      {endpointRes.url}
    </span>
    <span
      className="badge"
      style={{
        background: 'rgba(5, 150, 105, 0.15)',
        color: '#6ee7b7',
        border: '1px solid rgba(5, 150, 105, 0.3)',
        flexShrink: 0
      }}
    >
      {endpointRes.status} {endpointRes.statusText}
    </span>
  </div>

  <div className="flex gap-2" style={{ flexShrink: 0 }}>
    <button className={`btn btn-sm ${endpointTab === 'body' ? 'btn-primary' : 'btn-outline'}`} onClick={() => setEndpointTab('body')}>
      Body
    </button>
    <button className={`btn btn-sm ${endpointTab === 'headers' ? 'btn-primary' : 'btn-outline'}`} onClick={() => setEndpointTab('headers')}>
      Headers
    </button>
    <button className={`btn btn-sm ${endpointTab === 'metrics' ? 'btn-primary' : 'btn-outline'}`} onClick={() => setEndpointTab('metrics')}>
      PQC Metrics
    </button>
  </div>
</div>
```

#### Change 5: Section 3 Body Tab Copy Button Header (`Docs.tsx:1021-1045`)
Replace absolute button overlay with top control header bar `.code-header`:

```tsx
{endpointTab === 'body' && (
  <div>
    <div className="code-header">
      <span className="mono" style={{ fontSize: '0.75rem' }}>
        Response Body (JSON)
      </span>
      <button
        className="btn btn-outline btn-sm"
        onClick={() =>
          copyToClipboard(`res-body-${selectedEndpoint}-${tokenStrategy}`, JSON.stringify(endpointRes.body, null, 2))
        }
      >
        {copiedMap[`res-body-${selectedEndpoint}-${tokenStrategy}`] ? (
          <>
            <Check size={13} color="#10b981" />
            <span style={{ color: '#10b981' }}>Copied</span>
          </>
        ) : (
          <>
            <Copy size={13} />
            <span>Copy Response</span>
          </>
        )}
      </button>
    </div>
    <pre
      className="code-block"
      style={{
        borderTopLeftRadius: 0,
        borderTopRightRadius: 0,
        maxHeight: '340px'
      }}
    >
      {JSON.stringify(endpointRes.body, null, 2)}
    </pre>
  </div>
)}
```

---

## 5. Verification Checklist

To verify the proposed changes after implementation:

1. **320px Mobile Viewport Test (Chrome DevTools / Firefox Responsive View)**:
   - Navigate to `http://localhost:5173/docs`.
   - Set viewport width to 320px.
   - Verify that horizontal window scroll is **0px** (no body horizontal scrollbar).
   - Check Section 3: verify endpoint list stacks above parameter inputs and inspector panel.
   - Check parameter inputs: verify Client ID, Redirect URI, Subject, and Token Strategy stack vertically in single columns and option labels are fully visible.
   - Check code blocks in Section 1 and Section 2: verify horizontal scrollbar appears inside `<pre>` code container without extending the card width.
2. **Copy Button Non-Collision Check**:
   - Select `/token` endpoint with `full` strategy.
   - Click "Body" tab in Section 3 inspector.
   - Verify top lines of JSON (`{ "access_token": "..."`) are cleanly visible with no button overlay.
3. **Table & List Container Check**:
   - Inspect Section 3 PQC Metrics tab and Section 4 feature cards on 360px and 768px viewports.
   - Verify cryptographic hash strings wrap or scroll within container bounds.
