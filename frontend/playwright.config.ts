import { defineConfig, devices } from '@playwright/test'
import { loadEnv } from 'vite'

const env = loadEnv('test', process.cwd(), '')
const serverPort = Number(env.E2E_PORT || env.VITE_PORT || 4174)
const baseURL = env.E2E_BASE_URL || `http://127.0.0.1:${serverPort}`

process.env.E2E_BASE_URL = process.env.E2E_BASE_URL || baseURL
process.env.E2E_USERNAME = process.env.E2E_USERNAME || env.E2E_USERNAME || ''
process.env.E2E_PASSWORD = process.env.E2E_PASSWORD || env.E2E_PASSWORD || ''

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 2 : undefined,
  timeout: 30_000,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL,
    headless: process.env.E2E_HEADED !== 'true',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure'
  },
  projects: [
    {
      name: 'setup',
      testMatch: /.*auth\.setup\.ts/
    },
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        storageState: 'e2e/.auth/user.json'
      },
      dependencies: ['setup']
    }
  ],
  webServer: {
    command: `pnpm exec vite --mode test --host 127.0.0.1 --port ${serverPort}`,
    url: baseURL,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
    stdout: 'pipe',
    stderr: 'pipe'
  }
})
