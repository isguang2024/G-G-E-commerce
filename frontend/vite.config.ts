import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import { fileURLToPath } from 'url'
import { dirname } from 'path'
import vueDevTools from 'vite-plugin-vue-devtools'
import viteCompression from 'vite-plugin-compression'
import Components from 'unplugin-vue-components/vite'
import AutoImport from 'unplugin-auto-import/vite'
import ElementPlus from 'unplugin-element-plus/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import tailwindcss from '@tailwindcss/vite'
import { visualizer } from 'rollup-plugin-visualizer'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

export default ({ mode }: { mode: string }) => {
  const root = process.cwd()
  const env = loadEnv(mode, root)
  const { VITE_VERSION, VITE_PORT, VITE_BASE_URL, VITE_API_PROXY_URL } = env
  const apiProxyTarget = VITE_API_PROXY_URL || 'http://localhost:8080'

  return defineConfig({
    define: {
      __APP_VERSION__: JSON.stringify(VITE_VERSION)
    },
    base: VITE_BASE_URL,
    server: {
      port: Number(VITE_PORT),
      proxy: {
        '/api': {
          target: apiProxyTarget,
          changeOrigin: true
        }
      },
      host: true
    },
    // 路径别名
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        '@views': resolvePath('src/views'),
        '@imgs': resolvePath('src/assets/images'),
        '@icons': resolvePath('src/assets/icons'),
        '@utils': resolvePath('src/utils'),
        '@stores': resolvePath('src/store'),
        '@styles': resolvePath('src/assets/styles')
      }
    },
    build: {
      target: 'es2015',
      outDir: 'dist',
      chunkSizeWarningLimit: 2000,
      minify: 'terser',
      terserOptions: {
        compress: {
          // 生产环境去除 console
          drop_console: true,
          // 生产环境去除 debugger
          drop_debugger: true
        }
      },
      dynamicImportVarsOptions: {
        warnOnError: true,
        exclude: [],
        include: ['src/views/**/*.vue']
      },
      rollupOptions: {
        output: {
          // 重型依赖按 vendor chunk 拆分，避免主 chunk 膨胀
          manualChunks(id) {
            const normalizedId = id.replace(/\\/g, '/')
            // APP 维度分包：保持 account-portal / platform-admin / demo-app 代码边界稳定
            if (normalizedId.includes('/src/views/account-portal/')) return 'app-account-portal'
            if (normalizedId.includes('/src/views/system/') || normalizedId.includes('/src/views/dashboard/')) {
              return 'app-platform-admin'
            }
            if (normalizedId.includes('/src/views/demo/')) return 'app-demo'

            if (!id.includes('node_modules')) return
            if (normalizedId.includes('echarts') || normalizedId.includes('zrender')) return 'vendor-echarts'
            if (normalizedId.includes('xlsx')) return 'vendor-xlsx'
            if (normalizedId.includes('xgplayer')) return 'vendor-xgplayer'
            if (normalizedId.includes('element-plus') || normalizedId.includes('@element-plus')) return 'vendor-element-plus'
            if (normalizedId.includes('vue-img-cutter')) return 'vendor-image-cutter'
            if (normalizedId.includes('crypto-js')) return 'vendor-crypto'
            if (normalizedId.includes('@iconify')) return 'vendor-iconify'
            if (normalizedId.includes('@vue') || normalizedId.includes('vue-router') || normalizedId.includes('pinia')) return 'vendor-vue'
          }
        }
      }
    },
    plugins: [
      vue(),
      tailwindcss(),
      // 自动按需导入 API
      AutoImport({
        imports: ['vue', 'vue-router', 'pinia', '@vueuse/core'],
        dts: 'src/types/import/auto-imports.d.ts',
        resolvers: [ElementPlusResolver()],
        eslintrc: {
          enabled: true,
          filepath: './.auto-import.json',
          globalsPropValue: true
        }
      }),
      // 自动按需导入组件
      Components({
        dts: 'src/types/import/components.d.ts',
        resolvers: [ElementPlusResolver()]
      }),
      // 按需定制主题配置（useSource:true 会导入全部 scss 源码，改为默认按需 css）
      ElementPlus({}),
      // gzip 压缩
      viteCompression({
        verbose: false,
        disable: false,
        algorithm: 'gzip',
        ext: '.gz',
        threshold: 10240,
        deleteOriginFile: false
      }),
      // brotli 压缩（现代浏览器支持，比 gzip 再省 ~30%）
      viteCompression({
        verbose: false,
        disable: false,
        algorithm: 'brotliCompress',
        ext: '.br',
        threshold: 10240,
        deleteOriginFile: false
      }),
      vueDevTools(),
      // 打包分析：通过 `vite build --mode analyze` 触发，输出 dist/stats.html
      mode === 'analyze' &&
        visualizer({
          open: true,
          gzipSize: true,
          brotliSize: true,
          filename: 'dist/stats.html'
        })
    ].filter(Boolean),
    // 依赖预构建：避免运行时重复请求与转换，提升首次加载速度
    optimizeDeps: {
      entries: [
        'index.html',
        'src/main.ts',
        'src/components/**/*.vue',
        'src/views/**/*.vue',
        'src/router/**/*.ts',
        'src/store/**/*.ts',
        'src/hooks/**/*.ts',
        'src/composables/**/*.ts'
      ],
      // 开发环境首跳动态路由页时，避免运行中才补扫依赖导致整页 reload。
      ignoreOutdatedRequests: true,
      include: [
        'echarts/core',
        'echarts/charts',
        'echarts/components',
        'echarts/renderers',
        'xlsx',
        'xgplayer',
        'crypto-js',
        'file-saver',
        'vue-img-cutter',
        'vue-draggable-plus',
        'element-plus/es/components/row/style/css',
        'element-plus/es/components/col/style/css',
        'element-plus/es/components/radio/style/css',
        'element-plus/es/components/radio-group/style/css',
        'element-plus/es/components/radio-button/style/css',
        'element-plus/es/components/cascader/style/css',
        'element-plus/es/components/checkbox-group/style/css',
        'element-plus/es/components/date-picker/style/css',
        'element-plus/es/components/input-tag/style/css',
        'element-plus/es/components/rate/style/css',
        'element-plus/es/components/slider/style/css',
        'element-plus/es/components/time-picker/style/css',
        'element-plus/es/components/time-select/style/css',
        'element-plus/es/components/tree-select/style/css',
        'element-plus/es/components/table/style/css',
        'element-plus/es/components/table-column/style/css',
        'element-plus/es/components/pagination/style/css',
        'element-plus/es/components/empty/style/css',
        'element-plus/es/components/tag/style/css',
        'element-plus/es/components/watermark/style/css',
        'element-plus/es/components/dialog/style/css',
        'element-plus/es/components/form/style/css',
        'element-plus/es/components/tree/style/css',
        'element-plus/es/components/input/style/css',
        'element-plus/es/components/icon/style/css',
        'element-plus/es/components/switch/style/css',
        'element-plus/es/components/form-item/style/css',
        'element-plus/es/components/select/style/css',
        'element-plus/es/components/option/style/css',
        'element-plus/es/components/button/style/css',
        'element-plus/es/components/card/style/css',
        'element-plus/es/components/alert/style/css',
        'element-plus/es/components/checkbox/style/css',
        'element-plus/es/components/config-provider/style/css',
        'element-plus/es/components/menu/style/css',
        'element-plus/es/components/scrollbar/style/css',
        'element-plus/es/components/tooltip/style/css',
        'element-plus/es/components/dropdown/style/css',
        'element-plus/es/components/dropdown-menu/style/css',
        'element-plus/es/components/dropdown-item/style/css',
        'element-plus/es/components/result/style/css',
        'element-plus/es/components/option-group/style/css',
        'element-plus/es/components/popover/style/css',
        'element-plus/es/components/menu-item/style/css',
        'element-plus/es/components/sub-menu/style/css',
        'element-plus/es/components/message-box/style/css',
        'element-plus/es/components/drawer/style/css',
        'element-plus/es/components/input-number/style/css',
        '@iconify/vue'
      ]
    },
    css: {
      preprocessorOptions: {
        // sass variable and mixin
        scss: {
          additionalData: `
            @use "@styles/core/el-light.scss" as *; 
            @use "@styles/core/mixin.scss" as *;
          `
        }
      },
      postcss: {
        plugins: [
          {
            postcssPlugin: 'internal:charset-removal',
            AtRule: {
              charset: (atRule) => {
                if (atRule.name === 'charset') {
                  atRule.remove()
                }
              }
            }
          }
        ]
      }
    }
  })
}

function resolvePath(paths: string) {
  return path.resolve(__dirname, paths)
}
