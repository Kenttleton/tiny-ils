import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import { join } from "path";

// https://vite.dev/config/
/** @type {import('vite').UserConfig} */
export default defineConfig({
  plugins: [react()],
  build: {
    target: "modules",
    outDir: join(__dirname, "dist"),
    sourcemap: true,
    emptyOutDir: true,
  },
});
