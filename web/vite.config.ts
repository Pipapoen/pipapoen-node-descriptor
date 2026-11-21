import { defineConfig } from 'vite';
import postcss from './postcss.config.mjs';
import {fileURLToPath} from 'node:url';
export default defineConfig({
  root: 'web',
  css: {postcss},
  resolve:{alias:{'@pnd/transport':fileURLToPath(new URL('./src/transport.ts',import.meta.url))}},
  build: { outDir: '../internal/webui/assets', emptyOutDir: true },
  server: { proxy: { '/api': 'http://127.0.0.1:8080', '/auth': 'http://127.0.0.1:8080' } }
});
