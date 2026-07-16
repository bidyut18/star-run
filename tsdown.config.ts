import { defineConfig } from 'tsdown'

export default defineConfig({
  entry: ['./index.ts'],
  outDir: 'dist',
  format: 'esm',
  platform: 'node',
  target: 'node22',
  clean: true,
  dts: false,    
  minify: true,   
})