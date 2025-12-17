import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { toast } from "sonner";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useNavigate } from "react-router";
import OAuthMenu from "@/components/oauth";
import { useSession } from "@/stores/user";

const loginSchema = z.object({
  username: z.string(),
  password: z.string().min(1, "Password is required"),
});

type LoginFormData = z.infer<typeof loginSchema>;

export default function Login() {
  const navigate = useNavigate();
  const setSession = useSession((s) => s.setSession);
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  const login = useMutation(
    orpc.auth.login.mutationOptions({
      onSuccess(data) {
        console.log("Login successful:", data);
        toast.success("Login successful!");
        setSession(data);
        navigate("/app/dashboard");
      },
    }),
  );

  return (
    <div className="bg-background relative min-h-screen overflow-hidden">
      <div className="from-background absolute -top-10 left-0 h-1/2 w-full rounded-b-full bg-linear-to-b to-transparent blur"></div>
      <div className="from-primary/80 absolute -top-64 left-0 h-1/2 w-full rounded-full bg-linear-to-b to-transparent blur-3xl"></div>
      <div className="relative z-10 grid min-h-screen grid-cols-1 md:grid-cols-2">
        <div className="hidden flex-1 items-center justify-center space-y-8 p-8 text-center md:flex">
          <div className="space-y-6">
            <h1 className="text-4xl font-bold leading-tight md:text-5xl lg:text-6xl text-primary">
              Welcome Back to Visory!
            </h1>
          </div>
        </div>

        {/* Right Side - Login Form */}
        <div className="flex flex-1 items-center justify-center md:p-8">
          <Card className="md:border-border/70 md:bg-card/20 w-full md:max-w-md md:shadow-[0_10px_26px_#e0e0e0a1] md:backdrop-blur-lg dark:shadow-none border-none bg-transparent">
            <CardContent className="space-y-6 p-8">
              {/* Logo and Header */}
              <div className="space-y-4 text-center">
                <div className="flex items-center justify-center space-x-2">
                  <span className="text-2xl font-bold tracking-tight md:text-4xl">
                    Login
                  </span>
                </div>
                <p className="text-muted-foreground text-sm">
                  Create an account into visory and start managing your virtual
                  machines today.
                </p>
              </div>

              <form
                onSubmit={handleSubmit((data) => login.mutate(data))}
                className="space-y-6"
              >
                {/* Email Input */}
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    {...register("username")}
                    id="email"
                    placeholder="Enter your email or username"
                    autoComplete="off"
                  />
                  {errors.username && (
                    <p className="text-sm text-destructive">
                      {errors.username.message}
                    </p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="password">Password</Label>
                  <Input
                    {...register("password")}
                    id="password"
                    type="password"
                    placeholder="Enter your password"
                    className="border-border border"
                    autoComplete="off"
                  />
                  {errors.password && (
                    <p className="text-sm text-destructive">
                      {errors.password.message}
                    </p>
                  )}
                </div>

                {/* Continue Button */}
                <Button
                  className="w-full"
                  type="submit"
                  disabled={login.isPending}
                >
                  {login.isPending ? "Logging in..." : "Continue"}
                </Button>

                {/* Register Link */}
                <p className="text-center text-sm text-muted-foreground">
                  Don't have an account?{" "}
                  <button
                    type="button"
                    onClick={() => navigate("/auth/register")}
                    className="text-primary hover:underline font-medium"
                  >
                    Register here
                  </button>
                </p>
              </form>

              {/* Divider */}
              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="border-border w-full border-t"></div>
                </div>
                <div className="relative flex justify-center text-sm">
                  <span className="bg-card text-muted-foreground px-2">OR</span>
                </div>
              </div>

              <OAuthMenu Text="Sign in with" />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
