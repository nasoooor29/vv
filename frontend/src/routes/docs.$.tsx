import { source, browserCollections } from "@/lib/source";
import { createClientLoader } from "fumadocs-mdx/runtime/browser";
import { useParams } from "react-router";
import { useMemo, Suspense } from "react";

// Create client loader for MDX content
const clientLoader = createClientLoader(browserCollections.docs.raw, {
  id: "docs",
  component: ({ default: MDXContent, frontmatter }) => {
    return (
      <article className="prose prose-neutral dark:prose-invert max-w-none">
        <h1 className="text-3xl font-bold mb-4">{frontmatter.title}</h1>
        {frontmatter.description && (
          <p className="text-lg text-muted-foreground mb-8">
            {frontmatter.description}
          </p>
        )}
        <MDXContent />
      </article>
    );
  },
});

export default function DocsPage() {
  const params = useParams();
  const slugs = params["*"]?.split("/").filter((v) => v.length > 0) ?? [];
  const page = source.getPage(slugs);

  if (!page) {
    return (
      <div className="text-center py-16">
        <h1 className="text-2xl font-bold mb-4">Page Not Found</h1>
        <p className="text-muted-foreground">
          The documentation page you're looking for doesn't exist.
        </p>
      </div>
    );
  }

  // The path for clientLoader - without ./ prefix as it's stripped by the loader
  const mdxPath = `${page.slug}.mdx`;
  const Content = useMemo(() => clientLoader.getComponent(mdxPath), [mdxPath]);

  return (
    <Suspense fallback={<div className="animate-pulse">Loading...</div>}>
      <Content />
    </Suspense>
  );
}
