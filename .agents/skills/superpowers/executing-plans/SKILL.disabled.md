---
name: executing-plans
description: Use when a written implementation plan already exists and this repository needs that plan executed in a controlled sequence in the current or a separate session
---

# Executing Plans

## Overview

Load the plan, review it critically, execute it in order, and stop when reality diverges from the plan. This skill is for planned work that truly benefits from controlled execution, not for every normal coding task.

**Announce at start:** "I'm using the executing-plans skill to implement this plan."

**Repository policy:** If tasks are independent and subagents are available, prefer superpowers:subagent-driven-development. Use this skill when inline execution is the better fit.

## When to Use

Use this skill when:

- A real plan file already exists.
- The work has ordering constraints or shared state that make per-task subagents less suitable.
- The user wants the planned work executed in this session or in a tightly controlled batch.

Do not use this skill when:

- The task is small enough to implement directly.
- The plan is obsolete, vague, or clearly heavier than the work itself.
- The only remaining step is commit / PR / merge cleanup.

## The Process

### Step 1: Load and Review Plan
1. Read plan file
2. Review critically - identify any questions or concerns about the plan
3. If concerns: Raise them with your human partner before starting
4. If no concerns: Create TodoWrite and proceed

### Step 2: Execute Tasks

For each task:
1. Mark as in_progress
2. Follow each step exactly unless the codebase has materially changed
3. Run verifications as specified
4. Mark as completed

If a task reveals that the plan is wrong, stop and raise the discrepancy instead of freelancing a new plan in silence.

### Step 3: Complete Development

After all tasks complete and verified:
- Summarize what was executed and what was verified.
- If the user explicitly wants commit, PR, merge, branch cleanup, or integration handling:
  - Announce: "I'm using the finishing-a-development-branch skill to complete this work."
  - Use superpowers:finishing-a-development-branch
- Otherwise stop after verified execution and report results.

## When to Stop and Ask for Help

**STOP executing immediately when:**
- Hit a blocker (missing dependency, test fails, instruction unclear)
- Plan has critical gaps preventing starting
- You don't understand an instruction
- Verification fails repeatedly

**Ask for clarification rather than guessing.**

## When to Revisit Earlier Steps

**Return to Review (Step 1) when:**
- Partner updates the plan based on your feedback
- Fundamental approach needs rethinking

**Don't force through blockers** - stop and ask.

## Remember
- Review plan critically first
- Follow plan steps exactly
- Don't skip verifications
- Reference other skills only when the plan or repository context actually requires them
- Stop when blocked, don't guess
- Never start implementation on main/master branch without explicit user consent

## Integration

**Common companion skills:**
- **superpowers:writing-plans** - Creates the plan this skill executes
- **superpowers:subagent-driven-development** - Better default when tasks are independent
- **superpowers:verification-before-completion** - Required before making success claims
- **superpowers:finishing-a-development-branch** - Only when the user wants commit, PR, merge, or cleanup

**Not a default prerequisite:**
- **superpowers:using-git-worktrees** - Use only when isolation is actually needed
