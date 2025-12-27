import { source } from "@/lib/source";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { cn } from "@/lib/utils";
import { Cloud, Menu } from "lucide-react";
import { Link, Outlet, useLocation } from "react-router";
import { useState } from "react";

export default function DocsLayout() {
  const location = useLocation();
  const tree = source.pageTree;
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const sidebarContent = (
    <nav className="space-y-1">
      <DocsSidebarItems
        items={tree.children}
        currentPath={location.pathname}
        onNavigate={() => setSidebarOpen(false)}
      />
    </nav>
  );

  return (
    <div className="flex min-h-screen bg-background text-foreground">
      {/* Mobile Header */}
      <div className="md:hidden fixed top-0 left-0 right-0 z-40 flex items-center gap-2 p-4 border-b border-border bg-background">
        <Sheet open={sidebarOpen} onOpenChange={setSidebarOpen}>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" className="shrink-0">
              <Menu className="h-5 w-5" />
              <span className="sr-only">Toggle menu</span>
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-72 p-0">
            <SheetHeader className="border-b border-border">
              <SheetTitle className="flex items-center gap-2">
                <Cloud className="w-5 h-5 text-primary" />
                Visory Docs
              </SheetTitle>
            </SheetHeader>
            <ScrollArea className="flex-1 p-4">{sidebarContent}</ScrollArea>
          </SheetContent>
        </Sheet>
        <Link to="/" className="flex items-center gap-2">
          <Cloud className="w-5 h-5 text-primary" />
          <span className="font-bold">Visory Docs</span>
        </Link>
      </div>

      {/* Desktop Sidebar */}
      <aside className="hidden md:flex w-64 flex-col border-r border-border fixed inset-y-0 left-0">
        <div className="flex items-center gap-2 p-4 border-b border-border">
          <Link to="/" className="flex items-center gap-2">
            <Cloud className="w-6 h-6 text-primary" />
            <span className="text-lg font-bold">Visory Docs</span>
          </Link>
        </div>
        <ScrollArea className="flex-1 p-4">{sidebarContent}</ScrollArea>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-auto md:ml-64">
        <div className="max-w-4xl mx-auto px-6 py-8 pt-20 md:pt-8">
          <Outlet />
        </div>
      </main>
    </div>
  );
}

interface SidebarItemsProps {
  items: typeof source.pageTree.children;
  currentPath: string;
  onNavigate?: () => void;
}

function DocsSidebarItems({ items, currentPath, onNavigate }: SidebarItemsProps) {
  return (
    <>
      {items.map((item, index) => {
        if (item.type === "page") {
          const isActive = currentPath === item.url;
          return (
            <Link
              key={item.url}
              to={item.url!}
              onClick={onNavigate}
              className={cn(
                "block px-3 py-2 rounded-md text-sm transition-colors",
                isActive
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:bg-muted hover:text-foreground"
              )}
            >
              {item.name}
            </Link>
          );
        }
        if (item.type === "folder") {
          return (
            <div key={item.name + index} className="space-y-1">
              <div className="px-3 py-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                {item.name}
              </div>
              <div className="pl-2">
                <DocsSidebarItems
                  items={item.children!}
                  currentPath={currentPath}
                  onNavigate={onNavigate}
                />
              </div>
            </div>
          );
        }
        if (item.type === "separator") {
          return (
            <div key={`sep-${index}`} className="pt-4 pb-2">
              <div className="px-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                {item.name}
              </div>
            </div>
          );
        }
        return null;
      })}
    </>
  );
}
