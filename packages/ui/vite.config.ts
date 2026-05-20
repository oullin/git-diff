import tailwindcss from "@tailwindcss/vite";
import vue from "@vitejs/plugin-vue";
import { resolve } from "node:path";
import { defineConfig } from "vite";

export default defineConfig({
  base: "./",
  plugins: [vue(), tailwindcss()],
  cacheDir: resolve(__dirname, "../../.turbo/vite/ui"),
  resolve: {
    alias: {
      "@entry": resolve(__dirname, "./src"),
      "@": resolve(__dirname, "./src"),
      "@rich-text-editor": resolve(__dirname, "./src/components/ui/rich-text-editor"),
      "@app": resolve(__dirname, "./src/components/app"),
      "@components": resolve(__dirname, "./src/components"),
      "@ui": resolve(__dirname, "./src/components/ui"),
      "@composables": resolve(__dirname, "./src/composables"),
      "@lib": resolve(__dirname, "./src/lib"),
    },
  },
  test: {
    environment: "happy-dom",
    globals: true,
    alias: {
      "#electron-src/": resolve(__dirname, "./electron") + "/",
    },
  },
  server: {
    host: "127.0.0.1",
    allowedHosts: ["git-diff-ui.localhost"],
    hmr: {
      protocol: "wss",
      host: "git-diff-ui.localhost",
      clientPort: 1355,
    },
  },
});
