import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  build: {
    outDir: '../server/static',
    emptyOutDir: true,
    rollupOptions: {
      // Removes random generated hashes at the end
      output: {
        entryFileNames: `assets/[name].js`,
        chunkFileNames: 'assets/[name].js',
        assetFileNames: 'assets/[name].[ext]',
      },
    }
  },
})
