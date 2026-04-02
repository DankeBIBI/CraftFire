import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import AutoImport from "unplugin-auto-import/vite";
import { resolve } from "path";

/**
 * CraftFire 前端 Vite 配置。
 * 配置 Vue 3 插件、路径别名和构建优化。
 */
export default defineConfig({
  plugins: [
    // 自动导入 Vue / VueUse / Pinia / TresJS / 项目 composables
    AutoImport({
      imports: [
        // Vue 核心
        "vue",
        // VueUse
        "@vueuse/core",
        // Pinia
        "pinia",
        // TresJS 手动列出（auto-detect 有限制）
        {
          "@tresjs/core": [
            "useRenderLoop",
            "useTresContext",
          ],
        },
      ],
      // 项目的 composables 自动导入
      dirs: ["src/composables/**", "src/stores/**"],
      // 生成 TypeScript 类型定义文件
      dts: "src/auto-imports.d.ts",
      // eslintrc 生成（方便 IDE 识别）
      eslintrc: {
        enabled: true,
      },
      resolver: (name) => {
        if (name.startsWith("@/")) {
          return {
            name,
            from: resolve(__dirname, "src", name.slice(2)),
          };
        }
      },
    }),

    vue({
      template: {
        compilerOptions: {
          isCustomElement: (tag) =>
            tag.startsWith("Tres") && tag !== "TresCanvas",
        },
      },
    }),
  ],

  resolve: {
    alias: {
      "@": resolve(__dirname, "src"),
    },
  },

  server: {
    port: 5173,
    strictPort: true,
  },

  build: {
    outDir: "../backend/frontend/dist",
    emptyOutDir: true,
    sourcemap: false,
    minify: "esbuild",
    rollupOptions: {
      output: {
        manualChunks: {
          three: ["three"],
          vue: ["vue", "vue-router", "pinia"],
        },
      },
    },
  },
});
