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

const registerSchema = z
  .object({
    email: z.email("Please enter a valid email"),
    username: z.string().min(2, "Username must be at least 2 characters"),
    password: z.string().min(8, "Password must be at least 8 characters"),
    confirmPassword: z.string(),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ["confirmPassword"],
  });

type RegisterFormData = z.infer<typeof registerSchema>;

export default function Register() {
  const navigate = useNavigate();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
  });

  const registerMutation = useMutation(
    orpc.auth.register.mutationOptions({
      onSuccess(data) {
        console.log("Registration successful:", data);
        toast.success("Registration successful!");
        navigate("/dashboard");
      },
      onError() {
        toast.error("Registration failed. Please try again.");
      },
    }),
  );

  const onSubmit = (data: RegisterFormData) => {
    registerMutation.mutate({
      username: data.username,
      email: data.email,
      password: data.password,
      role: "user",
    });
  };

  return (
    <div className="bg-background relative min-h-screen overflow-hidden">
      <div className="from-background absolute -top-10 left-0 h-1/2 w-full rounded-b-full bg-linear-to-b to-transparent blur"></div>
      <div className="from-primary/80 absolute -top-64 left-0 h-1/2 w-full rounded-full bg-linear-to-b to-transparent blur-3xl"></div>
      <div className="relative z-10 grid min-h-screen grid-cols-1 md:grid-cols-2">
        <div className="hidden flex-1 items-center justify-center space-y-8 p-8 text-center md:flex">
          <div className="space-y-6">
            <h1 className="text-4xl font-bold leading-tight md:text-5xl lg:text-6xl text-primary">
              Join Visory Today!
            </h1>
            <p className="text-muted-foreground text-lg">
              Create an account and start managing your virtual machines with
              ease.
            </p>
          </div>
        </div>

        {/* Right Side - Register Form */}
        <div className="flex flex-1 items-center justify-center p-8">
          <Card className="border-border/70 bg-card/20 w-full max-w-lg shadow-[0_10px_26px_#e0e0e0a1] backdrop-blur-lg dark:shadow-none">
            <CardContent className="space-y-6 p-8">
              {/* Logo and Header */}
              <div className="space-y-4 text-center">
                <div className="flex items-center justify-center space-x-2">
                  <span className="text-2xl font-bold tracking-tight md:text-4xl">
                    Register
                  </span>
                </div>
                <p className="text-muted-foreground text-sm">
                  Create an account into visory and start managing your virtual
                  machines today.
                </p>
              </div>

              <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
                {/* Email Input */}
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    {...register("email")}
                    id="email"
                    type="email"
                    placeholder="Enter your email"
                    autoComplete="off"
                  />
                  {errors.email && (
                    <p className="text-sm text-destructive">
                      {errors.email.message}
                    </p>
                  )}
                </div>

                {/* Username Input */}
                <div className="space-y-2">
                  <Label htmlFor="username">Username</Label>
                  <Input
                    {...register("username")}
                    id="username"
                    type="text"
                    placeholder="Enter your username"
                    autoComplete="off"
                  />
                  {errors.username && (
                    <p className="text-sm text-destructive">
                      {errors.username.message}
                    </p>
                  )}
                </div>

                {/* Password Input */}
                <div className="space-y-2">
                  <Label htmlFor="password">Password</Label>
                  <Input
                    {...register("password")}
                    id="password"
                    type="password"
                    placeholder="At least 8 characters"
                    className="border-border border"
                    autoComplete="off"
                  />
                  {errors.password && (
                    <p className="text-sm text-destructive">
                      {errors.password.message}
                    </p>
                  )}
                </div>

                {/* Confirm Password Input */}
                <div className="space-y-2">
                  <Label htmlFor="confirm-password">Confirm Password</Label>
                  <Input
                    {...register("confirmPassword")}
                    id="confirm-password"
                    type="password"
                    placeholder="Confirm your password"
                    className="border-border border"
                    autoComplete="off"
                  />
                  {errors.confirmPassword && (
                    <p className="text-sm text-destructive">
                      {errors.confirmPassword.message}
                    </p>
                  )}
                </div>

                {/* Register Button */}
                <Button
                  className="w-full"
                  type="submit"
                  disabled={registerMutation.isPending}
                >
                  {registerMutation.isPending
                    ? "Creating account..."
                    : "Create Account"}
                </Button>

                {/* Login Link */}
                <p className="text-center text-sm text-muted-foreground">
                  Already have an account?{" "}
                  <button
                    type="button"
                    onClick={() => navigate("/login")}
                    className="text-primary hover:underline font-medium"
                  >
                    Login here
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

              <Button
                variant="outline"
                className="text-primary hover:bg-primary-foreground/95 w-full"
                type="button"
              >
                <svg
                  className="h-5 w-5 text-foreground"
                  viewBox="0 0 24 24"
                  fill="currentColor"
                >
                  <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
                  <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
                  <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
                  <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
                </svg>

                <span className="ml-2 text-foreground">
                  Sign up with Google
                </span>
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
