import { Button } from "@/components/ui/button";
import { Cloud } from "lucide-react";
import { Link, Outlet } from "react-router";

export default function LandingLayout() {
  console.log("LandingLayout rendered");
  return (
    <div className="flex flex-col min-h-screen bg-background text-foreground">
      {/* Navigation */}
      <nav className="flex items-center justify-between px-6 py-4 border-b border-border">
        <div className="flex items-center gap-2">
          <Cloud className="w-8 h-8 text-primary" />
          <h1 className="text-2xl font-bold text-foreground">Visory</h1>
        </div>
        <div className="flex items-center gap-4">
          <Link
            to="/auth/login"
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            Sign In
          </Link>
          <Link to="/auth/register">
            <Button className="bg-primary hover:bg-primary/90">
              Get Started
            </Button>
          </Link>
        </div>
      </nav>

      <Outlet />

      {/* Footer */}
      <footer className="px-6 py-8 border-t border-border">
        <div className="max-w-6xl mx-auto flex flex-col md:flex-row items-center justify-between">
          <div className="flex items-center gap-2 mb-4 md:mb-0">
            <Cloud className="w-6 h-6 text-primary" />
            <span className="text-foreground font-semibold">Visory</span>
          </div>
          <p className="text-muted-foreground text-sm">
            © 2025 Visory. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  );
}
