# Product

## Register

product

## Platform

web

## Users

Go developers who already use sqlc, or are actively evaluating it, and want a richer model layer on top of it. Most visits are mid-integration lookups: a developer with sqlc-model already wired into their codebase needs the exact config key, API shape, or lifecycle contract for the thing they're currently building. A secondary but real mode is evaluation: a developer deciding whether to adopt sqlc-model reads the explanation and architecture material to understand the boundary before committing.

## Product Purpose

These docs exist to make `sqlc-model` — an Eloquent-inspired rich model layer generated over sqlc — usable and trustworthy. They cover the full Diataxis split: tutorials to learn it, how-to guides to solve specific problems, reference for exact contracts, and explanation for the architectural reasoning. Success has two faces: an integrating developer finds the authoritative answer fast without wading through unrelated sections, and an evaluating developer reads enough to decide to adopt.

## Positioning

Generates fluent Active Record models, typed relationships, lifecycle state, and persistence adapters over statically declared sqlc queries — sqlc stays the source of truth for SQL, query signatures, and driver integration; the rich-model layer owns behavior and lifecycle on top of it.

## Brand Personality

Precise, disciplined, direct. The docs read like the README: terse architectural statements, explicit boundaries stated as fact rather than pitched, no marketing hedge language. Confidence comes from exactness, not enthusiasm.

## Anti-references

Not a generic ORM landing page. No hero gradients, no "blazing fast" vague performance copy, no feature-grid-as-decoration. The prose and structure should look like it was written by the people who built the boundary, not by someone selling it.

## Design Principles

The boundary is the argument: sqlc owns SQL and access, the rich-model layer owns behavior — every page should reinforce this split rather than blur it. One authoritative answer per question: don't let tutorial narrative and reference precision bleed into each other; a reader should never have to guess which section is the source of truth. Clarity over decoration: typography, code samples, and structure carry the credibility; visual flourish doesn't. Show exact contracts, not vibes: prefer real signatures, real config, real error names over paraphrase. Respect the reader's context: someone landing on a reference page is usually mid-task, not browsing — get them the answer with minimal navigation friction.

## Accessibility & Inclusion

WCAG AA baseline via MkDocs Material's built-in support (contrast, keyboard navigation, semantic structure). No additional known user needs beyond that baseline.
