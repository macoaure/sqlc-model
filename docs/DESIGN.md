---
name: sqlc-model docs
description: Precise, disciplined reference documentation for the sqlc-model Go library
colors:
  structural-indigo: "#3F51B5"
  signal-teal: "#009688"
typography:
  body:
    fontFamily: "Roboto, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"
    fontWeight: 400
    lineHeight: 1.6
  label:
    fontFamily: "'Roboto Mono', 'SF Mono', Consolas, monospace"
    fontWeight: 400
---

# Design System: sqlc-model docs

## 1. Overview

**Creative North Star: "The Reference Desk"**

This site is a technical reference desk, not a showroom. It runs on Material for MkDocs, essentially unmodified beyond a two-color palette choice (indigo primary, teal accent) and the extension set that enables tabs, annotated code blocks, and admonitions. Every structural decision — flat surfaces, a single functional shadow, no custom animation, no instant-loading JS layer — comes from the framework's own restraint, and this project has not fought it. That restraint is deliberate, not incidental: PRODUCT.md's positioning is precision over persuasion, and a documentation shell that adds visual noise on top of Material's defaults would undercut that.

Structural Indigo frames navigation, headers, and links — it establishes hierarchy, not mood. Signal Teal is reserved for functional emphasis: active nav state, search highlight, hover feedback. Neither color decorates; both signpost.

**Key Characteristics:**
- Flat by default; the only shadow in the system is functional (sticky header separation on scroll).
- Two-color palette used structurally (indigo = frame, teal = signal), not decoratively.
- Typography, code samples, and admonitions carry the credibility — no illustration, no hero imagery, no gradient.
- No customization beyond Material for MkDocs' defaults: consistent with a "precise, disciplined, direct" personality that has no interest in a distinctive visual identity for its own sake.

## 2. Colors

A restrained, two-role palette layered onto Material's neutral light/dark surfaces — nothing decorative, nothing beyond what wayfinding requires.

### Primary
- **Structural Indigo** (`#3F51B5`, Material Design Indigo 500): navigation bar, section headers, and inline links. Used for structure and orientation, not accent.

### Secondary
- **Signal Teal** (`#009688`, Material Design Teal 500): the toggle icons, active/selected states, and search-match highlighting. Reserved for "this is interactive or currently active," never used as a background wash.

### Neutral
- **Light surface** (Material default, inherited — white `#ffffff` background, near-black `rgba(0,0,0,.87)` text): the default (light) color scheme.
- **Slate surface** (Material default, inherited — deep navy-black `≈#14141a` background, pale lavender-white `≈#d2d7f9` text): the `slate` dark scheme, auto-switched via `prefers-color-scheme`.

These neutrals are Material for MkDocs' built-in variables; this project has not overridden them.

### Named Rules
**The Two-Color Rule.** Only indigo and teal carry meaning. If a third color shows up anywhere in this site, it's either a syntax-highlighting token (governed separately by the Pygments theme) or a mistake — never a new accent introduced for decoration.

## 3. Typography

**Body Font:** Roboto (system sans-serif fallback)
**Label/Mono Font:** Roboto Mono (system monospace fallback)

**Character:** Material's default pairing — a humanist grotesque body paired with its own monospace sibling. Neither font is customized; the pairing does the work of feeling neutral and technical rather than editorial.

### Hierarchy
- **Headline** (Material default weight/size scale): page and top-level section titles; sets the document's scan structure via `navigation.path` breadcrumbs and `toc.follow`.
- **Title** (Material default): H2/H3 section headers within a page.
- **Body** (400 weight, 1.6 line-height): prose content. No explicit max-width override is set beyond Material's own content column, which already keeps measure in a readable range.
- **Label** (Roboto Mono, Material default size): inline code, code blocks, and config keys — anywhere a literal token from the API needs to read as exact, not prose.

### Named Rules
**The Literal Token Rule.** Anything that is a real API name, config key, or error identifier renders in Roboto Mono, never in body prose styling — so a reader can visually distinguish "the exact thing to type" from "the explanation of it."

## 4. Elevation

Flat by default. The only shadow in the system is Material's built-in sticky-header separator, which appears solely when the page is scrolled and the header detaches from the content — a functional wayfinding cue, not a decorative lift. Admonitions, tabs, and code blocks are distinguished by background tint and border, never by shadow.

### Named Rules
**The Flat-By-Default Rule.** Surfaces are flat at rest. The single exception — header separation on scroll — exists purely to answer "am I still at the top of the page," not to add depth for its own sake. No card, admonition, or code block should ever pick up a shadow to look "elevated."

## 5. Components

Every component here is Material for MkDocs' stock implementation with the two-color palette applied; none are custom-built. The description below documents what's rendered, as a guardrail against future customization drifting toward decoration.

### Navigation
- **Style:** top-level tabs (`navigation.tabs`) for the five Diataxis-plus-project sections, with `navigation.sections` expanding each into a sidebar tree, `navigation.path` breadcrumbs, and `navigation.footer` for prev/next paging.
- **Active state:** Signal Teal indicates the current section/page; everything else stays indigo-on-neutral.
- **Search:** `search.suggest` + `search.highlight` — teal highlights the matched term in results, no other treatment.

### Admonitions
- **Style:** `admonition` + `pymdownx.details` (collapsible variants). Type-specific accent colors (note/tip/warning/danger) come from Material's own admonition palette, independent of the site's indigo/teal choice — this project has not remapped them.
- **Character:** used sparingly, for genuine caveats and warnings in reference/how-to content — not as decorative callout boxes.

### Code Blocks
- **Style:** `pymdownx.highlight` with `pygments_lang_class`, `content.code.copy` (copy button), `content.code.annotate` (inline numbered annotations).
- **Character:** the primary trust signal of the site — exact, copyable, annotated where a line needs explanation. No custom chrome beyond Material's default code-block frame.

### Tabs
- **Style:** `pymdownx.tabbed` with `alternate_style` and `content.tabs.link` (linked tab state across the page), used for showing config/API variants side by side (e.g. different query contract shapes).

### Signature Component: Diataxis Section Cards
The four documentation-type entry points (Tutorials / How-to / Reference / Explanation) plus Project are presented as plain linked lists in prose (see `content/index.md`), not as a card grid. This is a deliberate choice consistent with the Anti-references below — resist the temptation to turn this into an icon-plus-card grid on redesign.

## 6. Do's and Don'ts

### Do:
- **Do** keep indigo structural (nav/headers/links) and teal functional (active/highlight/interactive) — never swap their roles or introduce a third accent.
- **Do** let code blocks and literal API tokens (Roboto Mono) be the primary visual signal of trustworthiness — precision over decoration.
- **Do** keep surfaces flat; the sticky-header shadow is the only elevation in the system.
- **Do** route new content through Diataxis's four types plus Project — don't blend tutorial narrative into reference material or vice versa (PRODUCT.md: "one authoritative answer per question").

### Don't:
- **Don't** make this look like a generic ORM marketing site — no hero gradients, no "blazing fast" copy, no feature-grid-as-decoration (PRODUCT.md anti-reference, verbatim).
- **Don't** add heavy client-side interactivity or animation typical of some modern doc frameworks (scroll-driven reveals, choreographed transitions, `navigation.instant` prefetch flourishes) — keep it fast and text-first (PRODUCT.md anti-reference, verbatim).
- **Don't** turn the Diataxis section list into an icon-plus-card grid; it's intentionally plain-linked prose.
- **Don't** use a colored `border-left`/`border-right` stripe as a callout accent if admonition styling is ever customized — use Material's full background/title-bar treatment instead.
- **Don't** introduce a third named accent color. If something needs to stand out, reuse teal; if it needs structure, reuse indigo.
