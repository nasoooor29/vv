import browserCollections from "../../.source/browser";
import meta from "../../content/docs/meta.json";

// Page info type
interface PageInfo {
  slug: string;
  title: string;
  description?: string;
  url: string;
  path: string;
}

// Build pages map from meta.json
const pages: Map<string, PageInfo> = new Map();

// Known page titles (we'll extract from frontmatter when loaded)
const pageTitles: Record<string, string> = {
  index: "Documentation",
  "getting-started": "Getting Started",
  authentication: "Authentication",
  "oauth-setup": "OAuth Setup",
  rbac: "RBAC",
  docker: "Docker",
  "virtual-machines": "Virtual Machines",
  storage: "Storage",
  users: "Users",
  logs: "Logs",
  monitoring: "Monitoring",
  polling: "Polling",
};

// Build page tree from meta.json
interface PageTreeItem {
  type: "page" | "separator" | "folder";
  name: string;
  url?: string;
  children?: PageTreeItem[];
}

function buildPageTree(): { children: PageTreeItem[] } {
  const children: PageTreeItem[] = [];

  for (const item of meta.pages) {
    if (item.startsWith("---") && item.endsWith("---")) {
      // It's a separator/section header
      const name = item.slice(3, -3);
      children.push({ type: "separator", name });
    } else {
      // It's a page
      const url = item === "index" ? "/docs" : `/docs/${item}`;
      const title = pageTitles[item] || item.replace(/-/g, " ").replace(/\b\w/g, (l) => l.toUpperCase());
      
      pages.set(item, {
        slug: item,
        title,
        url,
        path: `/${item}`,
      });

      children.push({
        type: "page",
        name: title,
        url,
      });
    }
  }

  return { children };
}

const pageTree = buildPageTree();

// Get page by slugs
function getPage(slugs: string[]): PageInfo | undefined {
  const slug = slugs.length === 0 ? "index" : slugs.join("/");
  return pages.get(slug);
}

// Get all pages
function getPages(): PageInfo[] {
  return Array.from(pages.values());
}

export const source = {
  pageTree,
  getPage,
  getPages,
};

// Re-export browser collections for MDX loading
export { browserCollections };
