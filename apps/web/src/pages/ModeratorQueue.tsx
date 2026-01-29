import { useState } from "react";
import { Link } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Badge from "../components/ui/Badge";
import Button from "../components/ui/Button";
import { useReviewQueue } from "../hooks/useReviews";
import { formatDate, formatPercentage } from "../lib/utils";
import { Eye, Inbox, Clock, CheckCircle, Filter } from "lucide-react";

export default function ModeratorQueue() {
  const [filter, setFilter] = useState<"pending" | "reviewed" | undefined>("pending");
  const { data: reviews, isLoading, isError } = useReviewQueue({ status: filter });

  const getConfidenceColor = (confidence: number) => {
    if (confidence >= 0.8) return "from-red-500 to-red-400";
    if (confidence >= 0.5) return "from-yellow-500 to-yellow-400";
    return "from-emerald-500 to-emerald-400";
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Review Queue</h1>
          <p className="mt-1 text-slate-400">Moderate flagged content requiring human review</p>
        </div>
        <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-yellow-500/10 border border-yellow-500/20">
          <Clock className="w-4 h-4 text-yellow-400" />
          <span className="text-sm text-yellow-400 font-medium">
            {reviews?.filter(r => r.status === "pending").length || 0} Pending
          </span>
        </div>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Inbox className="w-5 h-5 text-blue-400" />
              Content Reviews
            </CardTitle>
            <div className="flex items-center gap-2">
              <Filter className="w-4 h-4 text-slate-500" />
              <div className="flex gap-2">
                <Button
                  variant={filter === "pending" ? "default" : "outline"}
                  size="sm"
                  onClick={() => setFilter("pending")}
                >
                  <Clock className="w-3 h-3 mr-1" />
                  Pending
                </Button>
                <Button
                  variant={filter === "reviewed" ? "default" : "outline"}
                  size="sm"
                  onClick={() => setFilter("reviewed")}
                >
                  <CheckCircle className="w-3 h-3 mr-1" />
                  Reviewed
                </Button>
                <Button
                  variant={filter === undefined ? "default" : "outline"}
                  size="sm"
                  onClick={() => setFilter(undefined)}
                >
                  All
                </Button>
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-12">
              <div className="w-8 h-8 border-2 border-blue-500/30 border-t-blue-500 rounded-full animate-spin mx-auto mb-4" />
              <p className="text-slate-500">Loading reviews...</p>
            </div>
          ) : isError ? (
            <div className="text-center py-12 text-red-400" data-testid="error-message">
              Failed to load reviews
            </div>
          ) : !reviews || reviews.length === 0 ? (
            <div className="text-center py-12">
              <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-slate-800 flex items-center justify-center">
                <Inbox className="w-8 h-8 text-slate-600" />
              </div>
              <p className="text-slate-500 mb-2">No reviews found</p>
              <p className="text-xs text-slate-600">
                {filter === "pending" ? "All content has been reviewed" : "No items match this filter"}
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto rounded-lg border border-slate-800">
              <table className="min-w-full divide-y divide-slate-800">
                <thead className="bg-slate-800/50">
                  <tr>
                    <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      ID
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Text Preview
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Category
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Confidence
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Created
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800">
                  {reviews.map((review) => {
                    const topCategory = Object.entries(review.category_scores).sort(
                      ([, a], [, b]) => (b || 0) - (a || 0)
                    )[0];

                    return (
                      <tr key={review.id} className="hover:bg-slate-800/30 transition-colors" data-testid="review-item">
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-slate-300">
                          <code className="px-2 py-1 rounded bg-slate-800 text-blue-400">
                            {review.id.slice(0, 8)}
                          </code>
                        </td>
                        <td className="px-6 py-4 text-sm text-slate-300 max-w-xs">
                          <div className="truncate" title={review.text}>
                            {review.text}
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <Badge variant="outline" className="capitalize">
                            {topCategory ? topCategory[0].replace(/_/g, " ") : "N/A"}
                          </Badge>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="flex items-center gap-2">
                            <div className="w-16 h-2 rounded-full bg-slate-800 overflow-hidden">
                              <div
                                className={`h-full rounded-full bg-gradient-to-r ${getConfidenceColor(review.confidence)}`}
                                style={{ width: `${review.confidence * 100}%` }}
                              />
                            </div>
                            <span className="text-sm text-slate-400">
                              {formatPercentage(review.confidence)}
                            </span>
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <Badge
                            variant={review.status === "pending" ? "warning" : "success"}
                            data-testid="review-status"
                          >
                            {review.status}
                          </Badge>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-500">
                          {formatDate(review.created_at)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm">
                          <Link to={`/reviews/${review.decision_id}`}>
                            <Button size="sm" variant="ghost">
                              <Eye className="h-4 w-4 mr-1" />
                              Review
                            </Button>
                          </Link>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
