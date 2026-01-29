import { Link, Outlet, useLocation } from "react-router-dom";
import {
  LayoutDashboard,
  MessageSquare,
  CheckCircle,
  FileText,
  ScrollText,
  BarChart3,
  LogOut,
  Shield,
  Sparkles,
} from "lucide-react";
import { useAuth } from "../../hooks/useAuth";
import { cn } from "../../lib/utils";

const navigation = [
  { name: "Dashboard", href: "/", icon: LayoutDashboard },
  { name: "Moderation Demo", href: "/demo", icon: MessageSquare },
  { name: "Review Queue", href: "/reviews", icon: CheckCircle },
  { name: "Policies", href: "/policies", icon: FileText },
  { name: "Audit Log", href: "/audit", icon: ScrollText },
  { name: "Analytics", href: "/analytics", icon: BarChart3 },
];

export default function MainLayout() {
  const location = useLocation();
  const { user, logout } = useAuth();

  return (
    <div className="flex h-screen bg-slate-900">
      {/* Sidebar */}
      <div className="flex w-64 flex-col bg-slate-900 border-r border-slate-800">
        {/* Logo */}
        <div className="flex h-16 items-center px-6 border-b border-slate-800">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-lg shadow-blue-500/25">
              <Shield className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="text-lg font-bold text-white">Civitas AI</h1>
              <p className="text-[10px] text-slate-500 -mt-0.5">Trust & Safety</p>
            </div>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 space-y-1 px-3 py-4">
          {navigation.map((item) => {
            const isActive = location.pathname === item.href;
            return (
              <Link
                key={item.name}
                to={item.href}
                className={cn(
                  "flex items-center px-3 py-2.5 text-sm font-medium rounded-lg transition-all duration-200",
                  isActive
                    ? "bg-gradient-to-r from-blue-500/20 to-purple-500/20 text-blue-400 border border-blue-500/30"
                    : "text-slate-400 hover:bg-slate-800 hover:text-slate-200"
                )}
              >
                <item.icon className={cn("mr-3 h-5 w-5", isActive && "text-blue-400")} />
                {item.name}
                {isActive && (
                  <div className="ml-auto w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse" />
                )}
              </Link>
            );
          })}
        </nav>

        {/* Built with AI Badge */}
        <div className="px-4 py-3 mx-3 mb-3 rounded-lg bg-gradient-to-r from-purple-500/10 to-pink-500/10 border border-purple-500/20">
          <div className="flex items-center gap-2 text-xs">
            <Sparkles className="w-4 h-4 text-purple-400" />
            <span className="text-purple-300 font-medium">Built with AI</span>
          </div>
        </div>

        {/* User section */}
        <div className="border-t border-slate-800 p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 text-white text-sm font-semibold shadow-lg shadow-blue-500/25">
                {user?.email?.charAt(0).toUpperCase()}
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-slate-200">{user?.email}</p>
                <p className="text-xs text-slate-500 capitalize">{user?.role}</p>
              </div>
            </div>
            <button
              onClick={logout}
              className="p-2 rounded-lg text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition-colors"
              title="Logout"
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="flex flex-1 flex-col overflow-hidden bg-slate-950">
        <main className="flex-1 overflow-y-auto">
          <div className="py-8 px-8">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}
