import { test as setup } from '@playwright/test'
import fs from 'node:fs'
import path from 'node:path'

const authFile = path.resolve(process.cwd(), 'e2e/.auth/user.json')

setup('prepare auth storage state', async () => {
  fs.mkdirSync(path.dirname(authFile), { recursive: true })

  // P3A-1 先只保证共享 storageState 文件存在。
  // 真实登录态复用在后续场景用例里按 .env.test 账号补齐。
  fs.writeFileSync(authFile, JSON.stringify({ cookies: [], origins: [] }, null, 2), 'utf8')
})
