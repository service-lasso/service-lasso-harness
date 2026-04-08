# Service Lasso - OpenSpec Draft Tracker

_Status: ref-only draft tracker_

Purpose:
- track the OpenSpec-style draft files being prepared in `ref/`
- keep focus on the main parts currently worth formalizing:
  - core
  - ui
  - service template
- record source docs, migration targets, and open questions before promotion into `.governance/specs/`

Important rule:
- these are **drafts in ref only** for analysis/design stabilization
- do not treat them as final governed specs until they are reviewed and promoted into the tracked governance area

## Draft Spec Register

| Draft Spec | Area | Status | Main Source Docs | Likely Promotion Target | Intended Repo Target | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `SPEC-CORE-SERVICE-RUNTIME.md` | Core | `draft` | `QUESTION-LIST-AND-CODE-VALIDATION.md`, `ARCHITECTURE-DECISIONS.md`, `SERVICE-MANAGER-BEHAVIOR.md`, `RUNTIME-API-INDEX.md` | future `.governance/specs/SPEC-CORE-SERVICE-RUNTIME.md` | `service-lasso` | Main runtime/service contract spec. |
| `SPEC-UI-ADMIN-SERVICE.md` | UI | `draft` (expanded first-pass) | `SERVICEADMIN-NAV-AND-API.md`, `UI-STATE-REVIEW.md`, `SHADCN-ADMIN-*`, `REFERENCE-APP-*` | future `.governance/specs/SPEC-UI-ADMIN-SERVICE.md` | `lasso-@serviceadmin` | Optional admin UI/service contract; first concrete pass written from donor UI docs. |
| `SPEC-SERVICE-TEMPLATE-REPO.md` | Service Template | `draft` (expanded first-pass) | `SERVICE-TEMPLATE-REPO.md`, `SERVICE-STRUCTURE-REVIEW.md`, representative `services/*/service.json` | future `.governance/specs/SPEC-SERVICE-TEMPLATE-REPO.md` | `service-template` | Service-author/template/release contract; first concrete pass written from donor template/structure docs. |
| `SPEC-SERVICE-LASSO-HARNESS.md` | Harness | `draft` | `README.md`, `docs/usage-flow.md`, `service-template` planning docs | future `.governance/specs/SPEC-SERVICE-LASSO-HARNESS.md` or repo-local governed spec | `service-lasso-harness` | Dedicated harness spec covering Go implementation direction, release-binary distribution, consumer-repo interaction, and minimum v1 runner behavior. |

## Explicit current repo mapping

The current intended repo mapping for the three ref OpenSpec drafts is:

1. Core -> `service-lasso`
2. UI -> `lasso-@serviceadmin`
3. Service Template -> `service-template`

## Promotion Checklist

Before moving any draft from `ref/openspec-drafts/` into `.governance/specs/`:

1. confirm the main intent is stable enough
2. check the draft against transcript-backed decisions
3. check the draft against donor code evidence
4. remove speculation that is not yet justified
5. make open questions explicit instead of hiding them in vague language
6. ensure acceptance criteria are concrete enough to govern future implementation

## Current Priority Order

1. Service Template
2. Core
3. UI

Reasoning:
- Max explicitly asked to do the template OpenSpec first
- template work is now concrete enough to advance immediately using the reconciled contract docs plus the donor service-structure review
- core still remains foundational, but current execution order is template first, then core tightening, then UI

## Current Open Questions To Review Later

### Core
- Should `config` remain a first-class action or only a lighter-weight regeneration path when concrete need appears?
- What exact normalized top-level `service.json` sections should replace donor `execconfig`?
- What exact `.state/` JSON shape should be written on start and updated on stop?
- Which lifecycle events stay log-only versus writing structured state?

### UI
- What exact minimum service-admin page contract should the optional admin UI support?
- Which state/log/runtime fields must the API expose versus compute client-side?
- How much UI should be generic versus service-specific plugin/extension driven?

### Service Template
- What exact template folder layout should become canonical?
- What exact sample service should ship first in the template repo?
- What exact release artifact layout and packaging scripts should be mandatory?
- How much generated/example state should the template include versus describe only?
