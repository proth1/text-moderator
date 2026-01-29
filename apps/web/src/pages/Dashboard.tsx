import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import {
  Activity,
  Shield,
  AlertCircle,
  FileCheck,
  TrendingUp,
  TrendingDown,
  Clock,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Zap,
  Users,
  ArrowUpRight,
  BarChart3,
} from "lucide-react";

// Mock data - replace with actual API calls
const stats = {
  total_moderations: 12453,
  blocked_percentage: 8.5,
  pending_reviews: 23,
  active_policies: 5,
  moderations_today: 342,
  reviews_today: 15,
  avg_response_time: 145, // ms
  accuracy_rate: 99.2,
};

const recentActivity = [
  { id: 1, action: "blocked", content: "Spam message detected", category: "spam", time: "2m ago", confidence: 0.94 },
  { id: 2, action: "allowed", content: "User comment approved", category: "safe", time: "5m ago", confidence: 0.98 },
  { id: 3, action: "review", content: "Flagged for human review", category: "borderline", time: "8m ago", confidence: 0.62 },
  { id: 4, action: "blocked", content: "Toxic content removed", category: "toxicity", time: "12m ago", confidence: 0.89 },
  { id: 5, action: "allowed", content: "Product review passed", category: "safe", time: "15m ago", confidence: 0.96 },
];

const categoryBreakdown = [
  { name: "Safe Content", count: 11245, percentage: 90.3, color: "bg-emerald-500" },
  { name: "Toxicity", count: 523, percentage: 4.2, color: "bg-red-500" },
  { name: "Spam", count: 412, percentage: 3.3, color: "bg-yellow-500" },
  { name: "Harassment", count: 189, percentage: 1.5, color: "bg-orange-500" },
  { name: "Other", count: 84, percentage: 0.7, color: "bg-purple-500" },
];

const ActionIcon = ({ action }: { action: string }) => {
  switch (action) {
    case "blocked":
      return <XCircle className="w-4 h-4 text-red-400" />;
    case "allowed":
      return <CheckCircle2 className="w-4 h-4 text-emerald-400" />;
    case "review":
      return <AlertTriangle className="w-4 h-4 text-yellow-400" />;
    default:
      return <Clock className="w-4 h-4 text-slate-400" />;
  }
};

export default function Dashboard() {
  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Dashboard</h1>
          <p className="mt-1 text-slate-400">Real-time moderation insights and analytics</p>
        </div>
        <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20">
          <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
          <span className="text-sm text-emerald-400 font-medium">System Online</span>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
        {/* Total Moderations */}
        <Card className="relative overflow-hidden group">
          <div className="absolute inset-0 bg-gradient-to-br from-blue-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-slate-400">
              Total Moderations
            </CardTitle>
            <div className="p-2 rounded-lg bg-blue-500/10">
              <Activity className="h-4 w-4 text-blue-400" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-white">{stats.total_moderations.toLocaleString()}</div>
            <div className="flex items-center gap-1 mt-2">
              <TrendingUp className="w-4 h-4 text-emerald-400" />
              <span className="text-sm text-emerald-400">+{stats.moderations_today}</span>
              <span className="text-sm text-slate-500">today</span>
            </div>
          </CardContent>
        </Card>

        {/* Blocked Rate */}
        <Card className="relative overflow-hidden group">
          <div className="absolute inset-0 bg-gradient-to-br from-red-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-slate-400">Block Rate</CardTitle>
            <div className="p-2 rounded-lg bg-red-500/10">
              <Shield className="h-4 w-4 text-red-400" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-white">{stats.blocked_percentage}%</div>
            <div className="flex items-center gap-1 mt-2">
              <TrendingDown className="w-4 h-4 text-emerald-400" />
              <span className="text-sm text-emerald-400">-0.3%</span>
              <span className="text-sm text-slate-500">from last week</span>
            </div>
          </CardContent>
        </Card>

        {/* Pending Reviews */}
        <Card className="relative overflow-hidden group">
          <div className="absolute inset-0 bg-gradient-to-br from-yellow-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-slate-400">Pending Reviews</CardTitle>
            <div className="p-2 rounded-lg bg-yellow-500/10">
              <AlertCircle className="h-4 w-4 text-yellow-400" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-white">{stats.pending_reviews}</div>
            <div className="flex items-center gap-1 mt-2">
              <Users className="w-4 h-4 text-slate-400" />
              <span className="text-sm text-slate-400">+{stats.reviews_today} today</span>
            </div>
          </CardContent>
        </Card>

        {/* Active Policies */}
        <Card className="relative overflow-hidden group">
          <div className="absolute inset-0 bg-gradient-to-br from-purple-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-slate-400">Active Policies</CardTitle>
            <div className="p-2 rounded-lg bg-purple-500/10">
              <FileCheck className="h-4 w-4 text-purple-400" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-white">{stats.active_policies}</div>
            <div className="flex items-center gap-1 mt-2">
              <CheckCircle2 className="w-4 h-4 text-emerald-400" />
              <span className="text-sm text-emerald-400">All active</span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Performance Metrics */}
      <div className="grid grid-cols-1 gap-5 lg:grid-cols-3">
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Zap className="w-5 h-5 text-yellow-400" />
              Performance
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-slate-400">Response Time</span>
                <span className="text-sm font-semibold text-white">{stats.avg_response_time}ms</span>
              </div>
              <div className="h-2 rounded-full bg-slate-800 overflow-hidden">
                <div className="h-full w-1/4 rounded-full bg-gradient-to-r from-emerald-500 to-emerald-400" />
              </div>
              <p className="text-xs text-slate-500 mt-1">Excellent - Target: &lt;500ms</p>
            </div>
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-slate-400">Accuracy Rate</span>
                <span className="text-sm font-semibold text-white">{stats.accuracy_rate}%</span>
              </div>
              <div className="h-2 rounded-full bg-slate-800 overflow-hidden">
                <div className="h-full w-[99%] rounded-full bg-gradient-to-r from-blue-500 to-purple-500" />
              </div>
              <p className="text-xs text-slate-500 mt-1">Above target of 95%</p>
            </div>
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-slate-400">System Load</span>
                <span className="text-sm font-semibold text-white">23%</span>
              </div>
              <div className="h-2 rounded-full bg-slate-800 overflow-hidden">
                <div className="h-full w-1/4 rounded-full bg-gradient-to-r from-cyan-500 to-cyan-400" />
              </div>
              <p className="text-xs text-slate-500 mt-1">Healthy - Capacity available</p>
            </div>
          </CardContent>
        </Card>

        {/* Recent Activity */}
        <Card className="lg:col-span-2">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Activity className="w-5 h-5 text-blue-400" />
              Recent Activity
            </CardTitle>
            <button className="flex items-center gap-1 text-sm text-blue-400 hover:text-blue-300 transition-colors">
              View all
              <ArrowUpRight className="w-4 h-4" />
            </button>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {recentActivity.map((item) => (
                <div
                  key={item.id}
                  className="flex items-center gap-4 p-3 rounded-lg bg-slate-800/50 hover:bg-slate-800 transition-colors"
                >
                  <ActionIcon action={item.action} />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-white truncate">{item.content}</p>
                    <div className="flex items-center gap-2 mt-1">
                      <span className="text-xs px-2 py-0.5 rounded-full bg-slate-700 text-slate-300">{item.category}</span>
                      <span className="text-xs text-slate-500">{item.time}</span>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm font-medium text-white">{(item.confidence * 100).toFixed(0)}%</div>
                    <div className="text-xs text-slate-500">confidence</div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Category Breakdown */}
      <div className="grid grid-cols-1 gap-5 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="w-5 h-5 text-purple-400" />
              Category Breakdown
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {categoryBreakdown.map((category) => (
                <div key={category.name}>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-slate-300">{category.name}</span>
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-white">{category.count.toLocaleString()}</span>
                      <span className="text-xs text-slate-500">({category.percentage}%)</span>
                    </div>
                  </div>
                  <div className="h-2 rounded-full bg-slate-800 overflow-hidden">
                    <div
                      className={`h-full rounded-full ${category.color} transition-all duration-500`}
                      style={{ width: `${category.percentage}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Quick Actions */}
        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-3">
              <a
                href="/demo"
                className="flex flex-col items-center justify-center p-4 rounded-xl bg-gradient-to-br from-blue-500/20 to-blue-600/10 border border-blue-500/30 hover:border-blue-400/50 transition-all group"
              >
                <Activity className="w-8 h-8 text-blue-400 mb-2 group-hover:scale-110 transition-transform" />
                <span className="text-sm font-medium text-white">Test Moderation</span>
                <span className="text-xs text-slate-500">Try the API</span>
              </a>
              <a
                href="/reviews"
                className="flex flex-col items-center justify-center p-4 rounded-xl bg-gradient-to-br from-yellow-500/20 to-yellow-600/10 border border-yellow-500/30 hover:border-yellow-400/50 transition-all group"
              >
                <AlertCircle className="w-8 h-8 text-yellow-400 mb-2 group-hover:scale-110 transition-transform" />
                <span className="text-sm font-medium text-white">Review Queue</span>
                <span className="text-xs text-slate-500">{stats.pending_reviews} pending</span>
              </a>
              <a
                href="/policies"
                className="flex flex-col items-center justify-center p-4 rounded-xl bg-gradient-to-br from-purple-500/20 to-purple-600/10 border border-purple-500/30 hover:border-purple-400/50 transition-all group"
              >
                <FileCheck className="w-8 h-8 text-purple-400 mb-2 group-hover:scale-110 transition-transform" />
                <span className="text-sm font-medium text-white">Manage Policies</span>
                <span className="text-xs text-slate-500">{stats.active_policies} active</span>
              </a>
              <a
                href="/audit"
                className="flex flex-col items-center justify-center p-4 rounded-xl bg-gradient-to-br from-emerald-500/20 to-emerald-600/10 border border-emerald-500/30 hover:border-emerald-400/50 transition-all group"
              >
                <Shield className="w-8 h-8 text-emerald-400 mb-2 group-hover:scale-110 transition-transform" />
                <span className="text-sm font-medium text-white">Audit Log</span>
                <span className="text-xs text-slate-500">View evidence</span>
              </a>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
