import {basename} from "node:path";
import type MarkdownIt from "markdown-it";
import type StateBlock from "markdown-it/lib/rules_block/state_block.mjs";
import {z} from "zod";
import {resolveLink} from "./links.js";
import type {Catalog} from "./model.js";

const environmentSchema = z.looseObject({
  frontmatter: z.looseObject({sourcePath: z.string().optional()}).optional(),
});

type MarkerState = Pick<
  StateBlock,
  "sCount" | "blkIndent" | "getLines" | "line"
>;

export const repositoryMarkdown = (
  md: {
    readonly block: {
      readonly ruler: {
        readonly before: (
          before: string,
          name: string,
          rule: (
            state: MarkerState,
            startLine: number,
            endLine: number,
            silent: boolean,
          ) => boolean,
          options: {readonly alt: string[]},
        ) => void;
      };
    };
    readonly renderer: MarkdownIt["renderer"];
    readonly utils: Pick<MarkdownIt["utils"], "escapeHtml">;
  },
  root: string,
  catalog: Catalog,
  base = "/",
): void => {
  md.block.ruler.before(
    "paragraph",
    "installation_markers",
    (state, startLine, _endLine, silent) => {
      if ((state.sCount[startLine] ?? 0) - state.blkIndent >= 4) return false;
      const line = state.getLines(
        startLine,
        startLine + 1,
        state.blkIndent,
        false,
      );
      if (!/^<!-- xonovex:installation:(?:start|end) -->$/u.test(line.trim()))
        return false;
      if (!silent) state.line = startLine + 1;
      return true;
    },
    {alt: ["paragraph", "reference", "blockquote"]},
  );
  const linkOpen = md.renderer.rules.link_open;
  md.renderer.rules.link_open = (
    tokens,
    index,
    options,
    environment: unknown,
    renderer,
  ) => {
    const source = environmentSchema.parse(environment).frontmatter?.sourcePath;
    const token = tokens[index];
    const href = token?.attrGet("href");
    if (
      source !== undefined &&
      href !== undefined &&
      href !== null &&
      token !== undefined
    ) {
      const resolved = resolveLink(root, catalog, source, href);
      token.attrSet("href", resolved.href);
      if (resolved.asset?.endsWith(".dot")) {
        token.attrSet("download", basename(resolved.asset));
        token.attrSet("href", base.slice(0, -1) + resolved.href);
      }
    }
    return linkOpen === undefined
      ? renderer.renderToken(tokens, index, options)
      : linkOpen(tokens, index, options, environment, renderer);
  };
  const image = md.renderer.rules.image;
  md.renderer.rules.image = (
    tokens,
    index,
    options,
    environment: unknown,
    renderer,
  ) => {
    const source = environmentSchema.parse(environment).frontmatter?.sourcePath;
    const token = tokens[index];
    const href = token?.attrGet("src");
    if (
      source !== undefined &&
      href !== undefined &&
      href !== null &&
      token !== undefined
    ) {
      const resolved = resolveLink(root, catalog, source, href);
      token.attrSet("src", resolved.href);
      if (resolved.dark !== undefined)
        return (
          '<ThemeDiagram light="' +
          md.utils.escapeHtml(resolved.href) +
          '" dark="' +
          md.utils.escapeHtml(resolved.dark) +
          '" alt="' +
          md.utils.escapeHtml(token.content) +
          '" />'
        );
    }
    return image === undefined
      ? renderer.renderToken(tokens, index, options)
      : image(tokens, index, options, environment, renderer);
  };
  const text = md.renderer.rules.text;
  md.renderer.rules.text = (
    tokens,
    index,
    options,
    environment: unknown,
    renderer,
  ) => {
    const rendered =
      text === undefined
        ? md.utils.escapeHtml(tokens[index]?.content ?? "")
        : text(tokens, index, options, environment, renderer);
    return rendered
      .replaceAll("{{", "&#123;&#123;")
      .replaceAll("}}", "&#125;&#125;");
  };
  const code = md.renderer.rules.code_inline;
  md.renderer.rules.code_inline = (
    tokens,
    index,
    options,
    environment: unknown,
    renderer,
  ) => {
    tokens[index]?.attrSet("v-pre", "");
    return code === undefined
      ? renderer.renderToken(tokens, index, options)
      : code(tokens, index, options, environment, renderer);
  };
  const fence = md.renderer.rules.fence;
  md.renderer.rules.fence = (
    tokens,
    index,
    options,
    environment: unknown,
    renderer,
  ) => {
    const token = tokens[index];
    if (token?.info.trim().startsWith("{")) token.info = "text";
    const rendered =
      fence === undefined
        ? renderer.renderToken(tokens, index, options)
        : fence(tokens, index, options, environment, renderer);
    return rendered.replaceAll(/<pre\b[^>]*>/gu, (tag) =>
      tag.includes("v-pre") ? tag : tag.replace("<pre", "<pre v-pre"),
    );
  };
};
