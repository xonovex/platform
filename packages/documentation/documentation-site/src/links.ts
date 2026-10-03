import {dirname, extname, resolve} from "node:path";
import {exists, inside, repositoryPath} from "./files.js";
import type {Catalog} from "./model.js";

export interface ResolvedLink {
  readonly href: string;
  readonly asset: string | undefined;
  readonly dark: string | undefined;
}

export const resolveLink = (
  root: string,
  catalog: Catalog,
  source: string,
  href: string,
): ResolvedLink => {
  const untouched = {href, asset: undefined, dark: undefined};
  if (
    href.startsWith("#") ||
    href.startsWith("//") ||
    /^[a-z][a-z\d+.-]*:/iu.test(href)
  )
    return untouched;
  const hash = href.indexOf("#");
  const fragment = hash === -1 ? "" : href.slice(hash);
  const withoutFragment = hash === -1 ? href : href.slice(0, hash);
  const query = withoutFragment.indexOf("?");
  const pathPart =
    query === -1 ? withoutFragment : withoutFragment.slice(0, query);
  const path = repositoryPath(
    root,
    inside(
      root,
      href.startsWith("/")
        ? pathPart.slice(1)
        : resolve(root, dirname(source), decodeURI(pathPart)),
    ),
  );
  const document = catalog.documents.find(
    (item) =>
      item.source === path ||
      item.source === path.replace(/\/$/u, "") + "/README.md",
  );
  if (document !== undefined)
    return {
      href: document.route + ".html" + fragment,
      asset: undefined,
      dark: undefined,
    };
  if (
    exists(root, path) &&
    [".png", ".jpg", ".jpeg", ".svg", ".gif", ".webp", ".dot", ".pdf"].includes(
      extname(path),
    )
  ) {
    const url = "/repository-assets/" + path;
    const themed =
      path.includes("plugin-workflow/diagrams/") && path.endsWith(".svg");
    return {
      href: url + fragment,
      asset: path,
      dark: themed ? url.replace(/\.svg$/u, ".dark.svg") : undefined,
    };
  }
  return {
    href: catalog.repository + "/blob/main/" + path + fragment,
    asset: undefined,
    dark: undefined,
  };
};
