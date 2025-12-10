import { Button } from "@/components/ui/button";
import { useSession } from "@/stores/user";
import {
  Cloud,
  Container,
  Zap,
  Shield,
  Activity,
  Gauge,
  ChevronRight,
  Check,
  Home,
} from "lucide-react";
import { Link } from "react-router";

export default function LandingPage() {
  const s = useSession((state) => state.session);
  console.log("LandingLayout rendered");
  return (
    <div className="flex flex-col min-h-screen bg-background text-foreground">
      {/* Hero Section */}
      <section className="flex-1 flex flex-col items-center justify-center px-6 py-20 text-center max-w-5xl mx-auto">
        <div className="space-y-6">
          <div className="inline-block px-4 py-2 bg-primary/10 border border-primary/30 rounded-full">
            <p className="text-sm text-primary">
              Unified VM & Container Management Platform
            </p>
          </div>

          <h2 className="text-5xl md:text-6xl font-bold text-foreground leading-tight">
            Manage VMs and Containers with
            <span className="text-primary"> Ease</span>
          </h2>

          <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
            Take control of your virtual infrastructure. Monitor, deploy, and
            scale both virtual machines and containerized applications from a
            single, intuitive dashboard.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center pt-4">
            {s ? (
              <>
                <Link to="/app">
                  <Button variant={"default"} size="lg" className="w-full">
                    Go to Dashboard
                    <Home />
                  </Button>
                </Link>
              </>
            ) : (
              <>
                <Link to="/auth/register">
                  <Button variant={"default"} size="lg" className="w-full">
                    Start Your Journey
                    <ChevronRight className="w-5 h-5 ml-2" />
                  </Button>
                </Link>
              </>
            )}
          </div>
        </div>
      </section>

      {/* Features Grid */}
      <section className="px-6 py-20 bg-card/50">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-16">
            <h3 className="text-3xl md:text-4xl font-bold text-foreground mb-4">
              Powerful Features
            </h3>
            <p className="text-muted-foreground text-lg">
              Everything you need to manage your virtual infrastructure
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {/* Feature 1 */}
            <div className="p-6 bg-card border border-border rounded-lg hover:border-primary/50 transition-colors">
              <Cloud className="w-8 h-8 text-primary mb-4" />
              <h4 className="text-xl font-semibold text-foreground mb-2">
                VM Management
              </h4>
              <p className="text-muted-foreground">
                Create, manage, and monitor virtual machines across your
                infrastructure with complete control
              </p>
            </div>

            {/* Feature 2 */}
            <div className="p-6 bg-card border border-border rounded-lg hover:border-primary/50 transition-colors">
              <Container className="w-8 h-8 text-accent mb-4" />
              <h4 className="text-xl font-semibold text-foreground mb-2">
                Container Orchestration
              </h4>
              <p className="text-muted-foreground">
                Deploy and manage containerized applications with ease,
                leveraging Docker and other container runtimes
              </p>
            </div>

            {/* Feature 3 */}
            <div className="p-6 bg-card border border-border rounded-lg hover:border-primary/50 transition-colors">
              <Activity className="w-8 h-8 text-chart-2 mb-4" />
              <h4 className="text-xl font-semibold text-foreground mb-2">
                Real-time Monitoring
              </h4>
              <p className="text-muted-foreground">
                Get instant insights into resource usage, performance metrics,
                and system health across all your infrastructure
              </p>
            </div>

            {/* Feature 4 */}
            <div className="p-6 bg-card border border-border rounded-lg hover:border-primary/50 transition-colors">
              <Shield className="w-8 h-8 text-chart-3 mb-4" />
              <h4 className="text-xl font-semibold text-foreground mb-2">
                Security & Compliance
              </h4>
              <p className="text-muted-foreground">
                Built-in security controls, access management, and audit logs to
                keep your infrastructure secure
              </p>
            </div>

            {/* Feature 5 */}
            <div className="p-6 bg-card border border-border rounded-lg hover:border-primary/50 transition-colors">
              <Zap className="w-8 h-8 text-chart-4 mb-4" />
              <h4 className="text-xl font-semibold text-foreground mb-2">
                Quick Provisioning
              </h4>
              <p className="text-muted-foreground">
                Spin up new VMs and containers in seconds with pre-configured
                templates and one-click deployment
              </p>
            </div>

            {/* Feature 6 */}
            <div className="p-6 bg-card border border-border rounded-lg hover:border-primary/50 transition-colors">
              <Gauge className="w-8 h-8 text-chart-5 mb-4" />
              <h4 className="text-xl font-semibold text-foreground mb-2">
                Resource Optimization
              </h4>
              <p className="text-muted-foreground">
                Intelligent resource allocation and recommendations to optimize
                performance and reduce costs
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Benefits Section */}
      <section className="px-6 py-20 bg-background">
        <div className="max-w-4xl mx-auto">
          <div className="text-center mb-16">
            <h3 className="text-3xl md:text-4xl font-bold text-foreground mb-4">
              Why Choose Visory?
            </h3>
            <p className="text-muted-foreground text-lg">
              Built for teams who want simplicity without sacrificing power
            </p>
          </div>

          <div className="space-y-4">
            {[
              "Unified dashboard for VMs and containers",
              "Intuitive UI with minimal learning curve",
              "Comprehensive API for automation",
              "Multi-user support with granular permissions",
              "Real-time alerts and notifications",
              "Detailed analytics and reporting",
            ].map((benefit, idx) => (
              <div
                key={idx}
                className="flex items-center gap-4 p-4 bg-card border border-border rounded-lg"
              >
                <Check className="w-6 h-6 text-primary flex-shrink-0" />
                <span className="text-foreground">{benefit}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="px-6 py-20 bg-card border-t border-border">
        <div className="max-w-3xl mx-auto text-center space-y-6">
          <h3 className="text-3xl md:text-4xl font-bold text-foreground">
            Ready to Simplify Your Infrastructure?
          </h3>
          <p className="text-lg text-muted-foreground">
            Join teams that are streamlining their VM and container management
            with Visory
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center pt-4">
            <Link to="/auth/register">
              <Button
                size="lg"
                className="bg-primary hover:bg-primary/90 text-lg px-8"
              >
                Get Started Now
                <ChevronRight className="w-5 h-5 ml-2" />
              </Button>
            </Link>
          </div>
        </div>
      </section>

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
