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

