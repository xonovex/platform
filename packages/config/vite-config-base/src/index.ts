import tailwind from "@tailwindcss/vite";
import {defineConfig} from "vite";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({
  resolve: {conditions: ["source", "development"]},
  plugins: [tsconfigPaths({projects: ["./tsconfig.json"]}), tailwind()],
});
