import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  build: {
    cssMinify: "esbuild",
    sourcemap: false,
    // Want to be performant as possible under 200kb chunks if possible
    chunkSizeWarningLimit: 200,
    outDir: "../server/static",
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes("chart.js") || id.includes("react-chartjs-2")) {
            return "chart-vendor";
          }
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
