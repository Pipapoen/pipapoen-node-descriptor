import { defineConfig } from 'vite';
export default defineConfig({
  root: 'web',
  build: { outDir: '../internal/webui/assets', emptyOutDir: true },
  server: { proxy: { '/api': 'http://127.0.0.1:8080', '/auth': 'http://127.0.0.1:8080' } }
});
