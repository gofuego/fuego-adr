---
title: Use JWT for API authentication
status: accepted
date_proposed: 2026-02-01
date_accepted: 2026-02-10
author: alice
approvers: [fabio]
supersedes: [4]
tags: [auth, security]
affects:
  - src/auth/**
  - src/middleware/auth.go
---

## Context

We need an authentication mechanism for our REST API. Options considered: session-based auth with cookies, JWT tokens, and OAuth2.

Our API is consumed by mobile clients and third-party services, making stateless authentication preferable.

## Decision

We will use **JWT (JSON Web Tokens)** for API authentication.

- Access tokens expire after 15 minutes
- Refresh tokens stored in the database, 30-day expiry
- RS256 signing with rotatable key pairs

## Consequences

- Stateless verification reduces database load per request
- Mobile clients can store tokens locally
- Token revocation requires a blocklist check (added complexity)
- Key rotation must be planned and automated
