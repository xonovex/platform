import {z} from "zod";

const modulePrefix = "xonovex-search-part:";
const chunkSize = 300_000;

export const splitSearchIndex = (
  code: string,
  id: string,
):
  | {readonly code: string; readonly modules: ReadonlyMap<string, string>}
  | undefined => {
  if (
    id === "/@localSearchIndex" ||
    code.length < chunkSize ||
    !id.startsWith("/@localSearchIndex")
  )
    return undefined;
  const json: unknown = JSON.parse(code.slice("export default ".length));
  const index = z.string().parse(json);
  const count = Math.ceil(index.length / chunkSize);
  const parts = Array.from({length: count}, (_, position) => ({
    id: modulePrefix + id.slice(1) + ":" + String(position),
    value: index.slice(position * chunkSize, (position + 1) * chunkSize),
  }));
  return {
    code:
      parts
        .map(
          (part, position) =>
            "import part" +
            String(position) +
            " from " +
            JSON.stringify(part.id) +
            ";",
        )
        .join("\n") +
      "\nexport default [" +
      parts.map((_, position) => "part" + String(position)).join(",") +
      "].join('');",
    modules: new Map(
      parts.map((part) => [
        part.id,
        "export default " + JSON.stringify(part.value),
      ]),
    ),
  };
};

export const searchChunkName = (id: string): string | undefined =>
  id.startsWith(modulePrefix) ? id.replaceAll(/[^\w-]/gu, "-") : undefined;

export const searchChunks = (): {
  readonly name: string;
  readonly apply: "build";
  readonly transform: (code: string, id: string) => string | undefined;
  readonly resolveId: (id: string) => string | undefined;
  readonly load: (id: string) => string | undefined;
} => {
  const modules = new Map<string, string>();
  return {
    name: "xonovex-search-chunks",
    apply: "build",
    transform(code, id) {
      const result = splitSearchIndex(code, id);
      if (result === undefined) return;
      for (const [part, content] of result.modules) modules.set(part, content);
      return result.code;
    },
    resolveId: (id) => (modules.has(id) ? id : undefined),
    load: (id) => modules.get(id),
  };
};
