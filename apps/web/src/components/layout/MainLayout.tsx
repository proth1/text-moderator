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

        {/* User section */}
        <div className="border-t border-slate-800 p-4">
          <div className="flex items-center gap-3">
            {/* Avatar */}
            <div className="flex-shrink-0 h-10 w-10 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-semibold shadow-lg shadow-blue-500/25">
              {user?.email?.charAt(0).toUpperCase() || "?"}
            </div>

            {/* User info */}
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-slate-200 truncate">
                {user?.email?.split("@")[0] || "User"}
              </p>
              <span className={cn(
                "inline-flex items-center px-2 py-0.5 rounded text-[10px] font-medium uppercase tracking-wide",
                user?.role === "admin" && "bg-purple-500/20 text-purple-300 border border-purple-500/30",
                user?.role === "moderator" && "bg-blue-500/20 text-blue-300 border border-blue-500/30",
                user?.role === "viewer" && "bg-slate-500/20 text-slate-300 border border-slate-500/30"
              )}>
                {user?.role || "user"}
              </span>
            </div>

            {/* Logout button */}
            <button
              onClick={logout}
              className="flex-shrink-0 p-2 rounded-lg text-slate-500 hover:text-red-400 hover:bg-red-500/10 transition-colors"
              title="Sign out"
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>

        {/* Built with AI Badge */}
        <div className="px-4 pb-4">
          <div className="flex items-center justify-center gap-2 py-2 text-[10px] text-slate-500">
            <Sparkles className="w-3 h-3" />
            <span>Built with AI</span>
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
