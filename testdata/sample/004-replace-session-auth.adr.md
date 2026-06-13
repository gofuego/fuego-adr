---
title: Replace session-based auth with JWT
status: superseded
date_proposed: 2026-01-05
date_superseded: 2026-02-10
author: bob
superseded_by: 2
tags: [auth]
---

## Context

Initial authentication was session-based with server-side cookie storage.

## Decision

We decided to use session-based authentication with Redis for session storage.

## Consequences

- Simple implementation
- Required Redis dependency
- Not suitable for mobile clients (cookie handling issues)
