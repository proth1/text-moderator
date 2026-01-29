import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import { BarChart3, TrendingUp, AlertTriangle, Activity, PieChart, Clock, Sparkles } from "lucide-react";

export default function Analytics() {
  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Analytics</h1>
          <p className="mt-1 text-slate-400">Moderation insights and performance trends</p>
        </div>
        <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-purple-500/10 border border-purple-500/20">
          <Sparkles className="w-4 h-4 text-purple-400" />
          <span className="text-sm text-purple-400 font-medium">Coming Soon</span>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Card className="group hover:border-blue-500/50 transition-all">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-slate-400">
              Moderation Volume
            </CardTitle>
            <div className="p-2 rounded-lg bg-blue-500/10 group-hover:bg-blue-500/20 transition-colors">
              <BarChart3 className="h-4 w-4 text-blue-400" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-white">—</div>
            <p className="text-xs text-slate-500 mt-1">Daily moderation statistics</p>
            <div className="mt-4 h-2 rounded-full bg-slate-800 overflow-hidden">
              <div className="h-full w-0 rounded-full bg-gradient-to-r from-blue-500 to-blue-400 animate-pulse" />
            </div>
          </CardContent>
        </Card>

        <Card className="group hover:border-emerald-500/50 transition-all">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-slate-400">
              Model Performance
            </CardTitle>
            <div className="p-2 rounded-lg bg-emerald-500/10 group-hover:bg-emerald-500/20 transition-colors">
              <TrendingUp className="h-4 w-4 text-emerald-400" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-white">—</div>
            <p className="text-xs text-slate-500 mt-1">Accuracy and confidence trends</p>
            <div className="mt-4 h-2 rounded-full bg-slate-800 overflow-hidden">
              <div className="h-full w-0 rounded-full bg-gradient-to-r from-emerald-500 to-emerald-400 animate-pulse" />
            </div>
          </CardContent>
        </Card>

        <Card className="group hover:border-yellow-500/50 transition-all">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-slate-400">
              False Positives
            </CardTitle>
            <div className="p-2 rounded-lg bg-yellow-500/10 group-hover:bg-yellow-500/20 transition-colors">
              <AlertTriangle className="h-4 w-4 text-yellow-400" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-white">—</div>
            <p className="text-xs text-slate-500 mt-1">Override rate analysis</p>
            <div className="mt-4 h-2 rounded-full bg-slate-800 overflow-hidden">
              <div className="h-full w-0 rounded-full bg-gradient-to-r from-yellow-500 to-yellow-400 animate-pulse" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Charts Grid */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="w-5 h-5 text-blue-400" />
              Moderation Trends
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-64 flex items-center justify-center rounded-lg bg-slate-800/30 border-2 border-dashed border-slate-700">
              <div className="text-center">
                <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-slate-800 flex items-center justify-center">
                  <BarChart3 className="h-8 w-8 text-slate-600" />
                </div>
                <p className="text-slate-500 font-medium">Time Series Chart</p>
                <p className="text-xs text-slate-600 mt-1">Moderation volume over time</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <PieChart className="w-5 h-5 text-purple-400" />
              Category Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-64 flex items-center justify-center rounded-lg bg-slate-800/30 border-2 border-dashed border-slate-700">
              <div className="text-center">
                <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-slate-800 flex items-center justify-center">
                  <PieChart className="h-8 w-8 text-slate-600" />
                </div>
                <p className="text-slate-500 font-medium">Distribution Chart</p>
                <p className="text-xs text-slate-600 mt-1">Content by category breakdown</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Clock className="w-5 h-5 text-emerald-400" />
              Response Times
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-64 flex items-center justify-center rounded-lg bg-slate-800/30 border-2 border-dashed border-slate-700">
              <div className="text-center">
                <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-slate-800 flex items-center justify-center">
                  <Clock className="h-8 w-8 text-slate-600" />
                </div>
                <p className="text-slate-500 font-medium">Latency Metrics</p>
                <p className="text-xs text-slate-600 mt-1">API response time percentiles</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="w-5 h-5 text-yellow-400" />
              Model Accuracy
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-64 flex items-center justify-center rounded-lg bg-slate-800/30 border-2 border-dashed border-slate-700">
              <div className="text-center">
                <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-slate-800 flex items-center justify-center">
                  <TrendingUp className="h-8 w-8 text-slate-600" />
                </div>
                <p className="text-slate-500 font-medium">Accuracy Trends</p>
                <p className="text-xs text-slate-600 mt-1">Model performance over time</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Coming Soon Banner */}
      <Card className="border-purple-500/30 bg-gradient-to-r from-purple-500/5 to-blue-500/5">
        <CardContent className="py-8">
          <div className="text-center">
            <Sparkles className="w-12 h-12 text-purple-400 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-white mb-2">Advanced Analytics Coming Soon</h3>
            <p className="text-slate-400 max-w-md mx-auto">
              We're building powerful analytics dashboards with real-time metrics,
              trend analysis, and actionable insights for your moderation workflow.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
