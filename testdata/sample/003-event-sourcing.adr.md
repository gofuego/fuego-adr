---
title: Adopt event sourcing for audit trail
status: tbd
date_proposed: 2026-06-01
author: fabio
deadline: 2026-06-20
tags: [architecture, database]
affects:
  - src/events/**
  - src/database/migrations/**
---

## Context

We need a complete audit trail for compliance. The current approach of after-the-fact logging is fragile and has gaps.

Event sourcing would make the audit trail a first-class citizen rather than a side effect.

## Decision

*Pending team discussion.*

## Consequences

*To be determined after decision is made.*
