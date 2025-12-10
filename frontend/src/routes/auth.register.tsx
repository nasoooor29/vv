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
        navigate("/app/dashboard");
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
        <div className="flex flex-1 items-center justify-center md:p-8">
          <Card className="md:border-border/70 md:bg-card/20 w-full md:max-w-md md:shadow-[0_10px_26px_#e0e0e0a1] md:backdrop-blur-lg dark:shadow-none border-none bg-transparent">
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
                    onClick={() => navigate("/auth/login")}
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

              <OAuthMenu Text="Continue with" />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
