import { Outlet } from "react-router";

export default function AuthLayout() {
  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-sm text-gray-500 p-2 bg-gray-100 absolute top-0 left-0">auth layout</div>
      <div className="w-full max-w-md">
        <Outlet />
      </div>
    </div>
  );
}
