/// <reference types="vitest/config" />
import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";

const target = process.env.JIN_WEB ?? "http://127.0.0.1:7373";

export default defineConfig({
  plugins: [svelte()],
  build: { outDir: "../internal/web/dist", emptyOutDir: true, chunkSizeWarningLimit: 1500 },
  server: {
    proxy: {
      "/api": {
        target,
        changeOrigin: true,
        configure: (proxy) => proxy.on("proxyReq", (req) => req.removeHeader("origin")),
      },
    },
  },
  test: { environment: "node" },
});
