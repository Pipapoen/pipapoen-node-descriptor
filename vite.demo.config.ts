import {defineConfig} from 'vite';
export default defineConfig({
  root:'demo', base:'./',
  css:{postcss:{plugins:[]}},
  build:{outDir:'../dist/demo',emptyOutDir:true},
});
