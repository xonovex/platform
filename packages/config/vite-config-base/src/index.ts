import tailwind from "@tailwindcss/vite";
import {defineConfig} from "vite";

export default defineConfig({
  resolve: {conditions: ["source", "development"], tsconfigPaths: true},
  plugins: [tailwind()],
});
