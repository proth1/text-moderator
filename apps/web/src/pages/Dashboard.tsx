import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import { Activity, Shield, AlertCircle, FileCheck } from "lucide-react";

// Mock data - replace with actual API calls
const stats = {
  total_moderations: 12453,
  blocked_percentage: 8.5,
  pending_reviews: 23,
  active_policies: 5,
  moderations_today: 342,
  reviews_today: 15,
};

export default function Dashboard() {
  return (
    <div>
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="mt-2 text-gray-600">Overview of moderation activities</p>
      </div>

      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">
              Total Moderations
            </CardTitle>
            <Activity className="h-4 w-4 text-gray-400" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total_moderations.toLocaleString()}</div>
            <p className="text-xs text-gray-600 mt-1">
              +{stats.moderations_today} today
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">Blocked Content</CardTitle>
            <Shield className="h-4 w-4 text-gray-400" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.blocked_percentage}%</div>
            <p className="text-xs text-gray-600 mt-1">Of all submissions</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">Pending Reviews</CardTitle>
            <AlertCircle className="h-4 w-4 text-gray-400" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.pending_reviews}</div>
            <p className="text-xs text-gray-600 mt-1">
              +{stats.reviews_today} today
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">Active Policies</CardTitle>
            <FileCheck className="h-4 w-4 text-gray-400" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.active_policies}</div>
            <p className="text-xs text-gray-600 mt-1">Published policies</p>
          </CardContent>
        </Card>
      </div>

      <div className="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-gray-600">
              <p className="mb-2">Activity feed coming soon...</p>
              <p>Recent moderation decisions and reviews will appear here.</p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Category Breakdown</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-gray-600">
              <p className="mb-2">Category statistics coming soon...</p>
              <p>Distribution of content by category will be displayed here.</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
