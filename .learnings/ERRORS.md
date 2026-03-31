# Errors`r`n`r`nCommand failures and integration errors.`r`n`r`n---`r`n
`r`n## 2026-04-01 用户创建空邮箱唯一索引冲突`r`n- 现象：`users.email` 是唯一索引，但创建用户时空字符串 `'` 会被当作真实值写入，导致第二个空邮箱用户插入失败。`r`n- 处理：改为部分唯一索引，仅对非空邮箱生效，并在创建/更新前 trim 后做邮箱查重。`r`n---`r`n
