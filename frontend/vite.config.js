import { defineConfig } from 'vite';
import path from 'path';

export default defineConfig({
  resolve: {
    alias: {
      '/wailsjs': path.resolve(__dirname, '../wailsjs/wailsjs'),
    },
  },
});
