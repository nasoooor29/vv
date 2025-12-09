import { Outlet } from "react-router";

export default function AuthLayout() {
  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="w-full">
        <Outlet />
      </div>
    </div>
  );
}
