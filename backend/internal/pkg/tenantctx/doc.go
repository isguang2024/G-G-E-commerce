// Package tenantctx carries the tenant_id through request context.
//
// GGE 5.0 reserves a tenant dimension across the whole stack (see doc ch.10).
// At this stage every request resolves to the built-in "default" tenant; the
// middleware, repository scopes and cache key prefixes that consume this
// package land in later phases.
package tenantctx
