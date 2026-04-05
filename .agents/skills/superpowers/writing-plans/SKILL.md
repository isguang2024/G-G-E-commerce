---
name: writing-plans
description: Use when a task in this repository is multi-step, spans multiple files or modules, changes frontend-backend contracts, or includes migrations and needs a written implementation plan before coding
---

# Writing Plans

## Overview

Write implementation plans only when the work is large enough to benefit from explicit sequencing. In this repository, plans are for non-trivial work: cross-module changes, frontend-backend contract updates, migrations, or tasks that will take sustained effort across multiple steps.

The plan should make execution safer and easier, not turn every small task into process theater. Prefer clear decomposition, exact file paths, verification points, and dependency ordering over exhaustive prose.

**Announce at start:** "I'm using the writing-plans skill to create the implementation plan."

**Repository policy:** Do not require this skill for single-page tweaks, isolated component fixes, copy updates, style cleanup, or single-endpoint adjustments.

**Save plans to:** `docs/superpowers/plans/YYYY-MM-DD-<feature-name>.md`
- (User preferences for plan location override this default)

## When to Use

Use this skill when one or more of these are true:

- The task spans 3 or more files and crosses module boundaries.
- The change touches both `frontend/` and `backend/`.
- The task changes API contracts, data shape, migrations, or seed behavior.
- The task needs staged rollout or ordered checkpoints.
- The task is large enough that losing the execution order would create rework.

Do not use this skill when:

- The change is local and obvious.
- The work can be completed safely in one short implementation pass.
- A brief execution note in the conversation is enough.

## Scope Check

If the request actually contains multiple independent workstreams, split them into separate plans or plan sections with explicit boundaries. Each section should produce testable progress on its own.

## File Structure

Before defining tasks, map out which files will be created or modified and what each one is responsible for. This is where the decomposition gets locked in.

- Design units with clear boundaries and well-defined interfaces. Each file should have one clear responsibility.
- You reason best about code you can hold in context at once, and your edits are more reliable when files are focused. Prefer smaller, focused files over large ones that do too much.
- Files that change together should live together. Split by responsibility, not by technical layer.
- In existing codebases, follow established patterns. If the codebase uses large files, don't unilaterally restructure - but if a file you're modifying has grown unwieldy, including a split in the plan is reasonable.

This structure informs the task decomposition. Each task should produce self-contained changes that make sense independently.

## Task Granularity

Default to task-level decomposition, not ultra-granular ceremony. A task should usually be 10-30 minutes of coherent work, with explicit verification.

Break work down further when:

- The change is risky or easy to regress.
- The task has a hard dependency chain.
- The repository already has nearby tests worth preserving through smaller increments.

Avoid forcing every plan into "test, fail, code, pass, commit" micro-steps when the repository context does not support it.

## Plan Document Header

Every plan should start with this header:

```markdown
# [Feature Name] Implementation Plan

> **Execution note:** Use superpowers:subagent-driven-development when tasks are independent, or superpowers:executing-plans when this plan should be executed inline in a controlled sequence.

**Goal:** [One sentence describing what this builds]

**Architecture:** [2-3 sentences about approach]

**Tech Stack:** [Key technologies/libraries]

---
```

## Task Structure

````markdown
### Task N: [Component Name]

**Files:**
- Create: `exact/path/to/file.py`
- Modify: `exact/path/to/existing.py:123-145`
- Test: `tests/exact/path/to/test.py`

- [ ] **Implement**

```python
# Exact code or structured change description needed for this task
```

- [ ] **Verify**

Run: `exact command`
Expected: `specific success signal`

- [ ] **Notes**

- Risks, follow-up checks, or contract assumptions for this task
````

If TDD is appropriate for this repository context, explicitly say so inside the task and reference superpowers:test-driven-development. Do not assume TDD is mandatory for every task.

## No Placeholders

Every step must contain the actual content an engineer needs. These are **plan failures** — never write them:
- "TBD", "TODO", "implement later", "fill in details"
- "Add appropriate error handling" / "add validation" / "handle edge cases"
- "Write tests for the above" (without exact files or commands)
- "Similar to Task N" (repeat the code — the engineer may be reading tasks out of order)
- Steps that describe what to do without giving file paths, commands, or enough structure to execute safely
- References to types, functions, or methods not defined in any task

## Remember
- Exact file paths always
- Give enough detail to execute without rediscovering the design
- Include exact verification commands with expected success signals
- Keep plans pragmatic and proportionate to task size
- Do not silently assume worktrees, TDD, or PR flow unless the task actually needs them

## Self-Review

After writing the complete plan, look at the spec with fresh eyes and check the plan against it. This is a checklist you run yourself — not a subagent dispatch.

**1. Spec coverage:** Skim each section/requirement in the spec. Can you point to a task that implements it? List any gaps.

**2. Placeholder scan:** Search your plan for red flags — any of the patterns from the "No Placeholders" section above. Fix them.

**3. Type consistency:** Do the types, method signatures, and property names you used in later tasks match what you defined in earlier tasks? A function called `clearLayers()` in Task 3 but `clearFullLayers()` in Task 7 is a bug.

If you find issues, fix them inline. No need to re-review — just fix and move on. If you find a spec requirement with no task, add the task.

## Execution Handoff

After saving the plan, offer execution choice:

**"Plan complete and saved to `docs/superpowers/plans/<filename>.md`. Two execution options:**

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?"**

**If Subagent-Driven chosen:**
- **REQUIRED SUB-SKILL:** Use superpowers:subagent-driven-development
- Fresh subagent per task + two-stage review

**If Inline Execution chosen:**
- **REQUIRED SUB-SKILL:** Use superpowers:executing-plans
- Batch execution with checkpoints for review

If the task is still small after review, say so explicitly and skip the plan execution workflow.
