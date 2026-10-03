import {z} from "zod";
import type {PaletteColor} from "./model.js";

const colorSchema = z.string().regex(/^#[\da-f]{6}$/iu);
const requiredTokens = [
  "canvas",
  "surface",
  "text",
  "secondary",
  "muted",
  "border",
  "strong",
  "blue",
  "blue-strong",
  "blue-soft",
  "green",
  "green-strong",
  "green-soft",
];

export const designPalette = (document: string): readonly PaletteColor[] => {
  const colors = [
    ...document.matchAll(
      /^\|[^\n|]+\|\s*`([a-z-]+)`\s*\|\s*`(#[\da-f]{6})`\s*\|\s*`(#[\da-f]{6})`\s*\|/gimu,
    ),
  ].map((match) => ({
    token: match[1] ?? "",
    light: colorSchema.parse(match[2]),
    dark: colorSchema.parse(match[3]),
  }));
  const missing = requiredTokens.filter((token) =>
    colors.every((color) => color.token !== token),
  );
  if (
    missing.length > 0 ||
    new Set(colors.map((color) => color.token)).size !== colors.length
  )
    throw new Error(
      "DESIGN.md must define one light and dark color for every token: " +
        missing.join(", "),
    );
  return colors;
};

export const paletteCss = (palette: readonly PaletteColor[]): string =>
  ["light", "dark"]
    .map((mode) => {
      const declarations = palette.map(
        (color) =>
          "  --xonovex-" +
          color.token +
          ": " +
          (mode === "light" ? color.light : color.dark) +
          ";",
      );
      return (
        (mode === "light" ? ":root" : ".dark") +
        " {\n" +
        declarations.join("\n") +
        "\n}"
      );
    })
    .join("\n\n") + "\n";

export const darkSvg = (
  svg: string,
  palette: readonly PaletteColor[],
): string => {
  const mapping = new Map(
    palette.map((color) => [color.light.toLowerCase(), color.dark]),
  );
  return svg.replaceAll(
    /\b(fill|stroke)="(#[\da-f]{6}|white|black)"/giu,
    (match: string, attribute: string, color: string) => {
      const named = new Map([
        ["white", "#ffffff"],
        ["black", "#000000"],
      ]);
      const key = named.get(color.toLowerCase()) ?? color.toLowerCase();
      const dark = mapping.get(key);
      return dark === undefined ? match : attribute + '="' + dark + '"';
    },
  );
};
