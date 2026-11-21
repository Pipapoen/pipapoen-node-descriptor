import {defineConfig} from 'vite';
import {fileURLToPath} from 'node:url';
export default defineConfig({
  root:'demo', base:'./',
  resolve:{alias:{'@pnd/transport':fileURLToPath(new URL('./demo/src/transport.ts',import.meta.url))}},
  build:{outDir:'../dist/demo',emptyOutDir:true},
});
