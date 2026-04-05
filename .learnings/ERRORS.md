# Errors`r`n`r`nCommand failures and integration errors.`r`n`r`n---`r`n
`r`n## 2026-04-01 用户创建空邮箱唯一索引冲突`r`n- 现象：`users.email` 是唯一索引，但创建用户时空字符串 `'` 会被当作真实值写入，导致第二个空邮箱用户插入失败。`r`n- 处理：改为部分唯一索引，仅对非空邮箱生效，并在创建/更新前 trim 后做邮箱查重。`r`n---`r`n
## [ERR-20260405-001] shadcn-skill-move

**Logged**: 2026-04-05T00:00:00+08:00
**Priority**: medium
**Status**: pending
**Area**: docs

### Summary
Tried to copy and delete the shadcn skill directory in parallel, and deletion completed before copy.

### Error
`	ext
Copy-Item : Cannot find path 'C:\Users\Administrator\Documents\GitHub\G-G-E-commerce\.agents\skills\shadcn' because it does not exist.
` 

### Context
- Operation attempted: move project-local shadcn skill to global Codex skills directory
- Cause: copy and delete were executed in parallel against the same source path
- Recovery: rebuild global copy from git-tracked HEAD contents, keep local deletion as intended

### Suggested Fix
Never parallelize a move operation that reads and deletes the same directory; copy first, verify, then delete.

### Metadata
- Reproducible: yes
- Related Files: .agents/skills/shadcn

---

## [ERR-20260406-001] rg-access-denied-backend-audit

**Logged**: 2026-04-06T00:00:00+08:00
**Priority**: medium
**Status**: pending
**Area**: docs

### Summary
`rg.exe` again failed with `Access is denied` during backend audit, so source scanning had to fall back to PowerShell primitives.

### Error
```text
Program 'rg.exe' failed to run: Access is denied
```

### Context
- Operation attempted: recursive source search under `backend/` for `menus`, `ui_pages.space_key`, `visibility_scope`, and page-space binding references
- Impact: default fast search path is unreliable in this Windows workspace session
- Workaround used: `Get-ChildItem -Recurse -File` plus `Select-String -Encoding UTF8`, with later narrowing to source extensions only

### Suggested Fix
Treat `rg.exe` as optional in this environment and prefer a documented PowerShell fallback for Windows audits unless executable permissions are fixed.

### Metadata
- Reproducible: yes
- Related Files: backend
- See Also: ERR-20260405-002

---

## [ERR-20260405-002] rg-access-denied

**Logged**: 2026-04-05T00:00:00+08:00
**Priority**: medium
**Status**: pending
**Area**: docs

### Summary
`rg.exe` is present in the environment but cannot be executed because PowerShell returns `Access is denied`.

### Error
```text
Program 'rg.exe' failed to run: Access is denied
```

### Context
- Operation attempted: recursive code search under `frontend/src`
- Impact: normal `rg`-based search workflow is unavailable in this workspace session
- Workaround used: PowerShell `Get-ChildItem -Recurse` plus `Select-String -Encoding UTF8`
- Observed again on 2026-04-06 while verifying `docs/superpowers-integration.md` and `docs/change-log.md`

### Suggested Fix
If this environment is expected to support `rg`, verify execution policy or binary permissions; otherwise prefer documenting PowerShell search as the fallback path for Windows sessions.

### Metadata
- Reproducible: yes
- Related Files: frontend/src

---
