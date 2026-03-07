import defaultMdxComponents from "fumadocs-ui/mdx";
import * as TabsComponents from "fumadocs-ui/components/tabs";
import type { MDXComponents } from "mdx/types";
import { ApiDemoWithCode } from "@/components/api-demo-server";

export function getMDXComponents(components?: MDXComponents): MDXComponents {
  return {
    ...defaultMdxComponents,
    ...TabsComponents,
    ApiDemo: ApiDemoWithCode,
    ...components,
  };
}
