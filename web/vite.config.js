import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

// Each game session runs its own dev server on its own port, so each gets its own cache and they never trip over
// each other.
const port = process.argv.includes('--port') ? process.argv[process.argv.indexOf('--port') + 1] : null;

export default defineConfig({
  plugins: [svelte()],
  cacheDir: port ? `node_modules/.vite-${port}` : 'node_modules/.vite',
  build: {
    outDir: '../server/internal/webui/dist',
    emptyOutDir: true,
    assetsInlineLimit: 0,
    // The newest browser that still runs on Windows 7.
    target: 'chrome109',
  },
  // AMK_API points the dev server at a test copy, so it never touches the real till on 8417.
  server: { proxy: { '/api': process.env.AMK_API ?? 'http://127.0.0.1:8417' } },
});
