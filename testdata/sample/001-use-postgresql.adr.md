---
title: Use PostgreSQL for persistence
status: accepted
date_proposed: 2026-01-10
date_accepted: 2026-01-15
author: fabio
approvers: [alice, bob]
tags: [database, infrastructure]
affects:
  - src/database/**
  - docker-compose.yaml
---

## Context

We need a relational database for our application. The main options are PostgreSQL, MySQL, and SQLite. Our team has strong PostgreSQL experience, and we need advanced features like JSONB columns and full-text search.

## Decision

We will use **PostgreSQL 16** as our primary data store.

- All persistent state goes through PostgreSQL
- We will use connection pooling via PgBouncer
- Schema migrations managed by `golang-migrate`

## Consequences

- Strong ecosystem and community support
- JSONB columns allow flexible schema evolution
- Operational overhead for running a database server in production
- Team training needed for PgBouncer configuration
