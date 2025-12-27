import path from "path";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import { reactRouter } from "@react-router/dev/vite";
import mdx from "fumadocs-mdx/vite";
import * as MdxConfig from "./source.config";

// https://vite.dev/config/
// https://reactrouter.com/tutorials/quickstart#vite-config
export default defineConfig({
  plugins: [mdx(MdxConfig), tailwindcss(), reactRouter()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  ssr: {
    // Don't externalize these packages during SSR
    noExternal: ["fumadocs-core", "fumadocs-ui", "@fumadocs/ui", "fumadocs-mdx"],
  },
});
