import { docs } from "fumadocs-mdx:collections/server";
import { type InferPageType, loader, multiple } from "fumadocs-core/source";
import { lucideIconsPlugin } from "fumadocs-core/source/lucide-icons";
import { openapiPlugin, openapiSource } from "fumadocs-openapi/server";
import { openapi } from "@/lib/openapi";
import { icons } from "lucide-react";
import { createElement } from "react";

const openapiFiles = await openapiSource(openapi, { groupBy: "tag" });
openapiFiles.files.push({
  type: "meta",
  path: "shelf/meta",
  data: {
    defaultOpen: true,
  },
} as any);

export const source = loader(
  multiple({
    docs: docs.toFumadocsSource(),
    openapi: openapiFiles,
  }),
  {
    baseUrl: "/",
    plugins: [lucideIconsPlugin(), openapiPlugin()],
  },
);

export function getPageImage(page: InferPageType<typeof source>) {
  const segments = [...page.slugs, "image.webp"];

  return {
    segments,
    url: `/og/${segments.join("/")}`,
  };
}
