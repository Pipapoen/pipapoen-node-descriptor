import {defineConfig} from 'vite';
import postcss from './postcss.config.mjs';
import {fileURLToPath} from 'node:url';
export default defineConfig({
  root:'demo', base:'./',
  css:{postcss},
  resolve:{alias:{'@pnd/transport':fileURLToPath(new URL('../demo/src/transport.ts',import.meta.url))}},
  build:{outDir:'../build/site',emptyOutDir:true},
});
