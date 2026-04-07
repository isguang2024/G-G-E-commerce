// Package evaluator is the single entry point for permission decisions in GGE 5.0.
//
// Final permission = workspace feature package keys ∩ member role keys.
// Backed by casbin/v2 with a two-tier cache (local LRU + Redis).
//
// This package is a placeholder introduced in Phase 0 of the v5 refactor.
// The Evaluator interface and casbin model land in Phase 3.
package evaluator
