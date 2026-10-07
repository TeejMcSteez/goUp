import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  build: {
    cssMinify: "esbuild",
    target: "esnext",
    sourcemap: false,
    // Skip polyfill injection modern browsers handle natively
    modulePreload: { polyfill: false },
    // Want to be performant as possible under 200kb chunks if possible
    chunkSizeWarningLimit: 220,
    outDir: "../server/static",
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (
            id.includes("node_modules/react") ||
            id.includes("node_modules/react-dom")
          ) {
            return "react-vendor";
          }
        },
      },
    },
  },
});
