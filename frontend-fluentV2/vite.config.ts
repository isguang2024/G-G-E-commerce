import path from 'node:path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          const normalizedId = id.replace(/\\/g, '/')

          if (normalizedId.includes('node_modules')) {
            if (normalizedId.includes('@fluentui')) return 'vendor-fluent'
            if (normalizedId.includes('@tanstack') || normalizedId.includes('react-router') || normalizedId.includes('zustand')) {
              return 'vendor-react'
            }
            return 'vendor'
          }

          if (normalizedId.includes('/src/pages/auth/') || normalizedId.includes('/src/pages/status/')) return 'auth'
          if (normalizedId.includes('/src/pages/dashboard/') || normalizedId.includes('/src/features/dashboard/')) return 'dashboard'
          if (normalizedId.includes('/src/pages/workspace/') || normalizedId.includes('/src/features/inbox/')) return 'workspace'
          if (normalizedId.includes('/src/pages/message/') || normalizedId.includes('/src/features/message/')) return 'message'
          if (
            normalizedId.includes('/src/pages/system/') ||
            normalizedId.includes('/src/features/system/') ||
            normalizedId.includes('/src/features/access/')
          ) {
            return 'system'
          }
          if (normalizedId.includes('/src/pages/team/') || normalizedId.includes('/src/features/team/')) return 'team'

          return undefined
        },
      },
    },
  },
  server: {
    host: '127.0.0.1',
    port: 9030,
    proxy: {
      '/api': {
        target: process.env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
    },
  },
})
