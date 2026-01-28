import { useState } from "react";
import { Link } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Badge from "../components/ui/Badge";
import Button from "../components/ui/Button";
import { useReviewQueue } from "../hooks/useReviews";
import { formatDate, formatPercentage } from "../lib/utils";
import { Eye } from "lucide-react";

export default function ModeratorQueue() {
  const [filter, setFilter] = useState<"pending" | "reviewed" | undefined>("pending");
  const { data: reviews, isLoading, isError } = useReviewQueue({ status: filter });

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Review Queue</h1>
        <p className="mt-2 text-gray-600">Moderate flagged content</p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Pending Reviews</CardTitle>
            <div className="flex gap-2">
              <Button
                variant={filter === "pending" ? "default" : "outline"}
                size="sm"
                onClick={() => setFilter("pending")}
              >
                Pending
              </Button>
              <Button
                variant={filter === "reviewed" ? "default" : "outline"}
                size="sm"
                onClick={() => setFilter("reviewed")}
              >
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
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-gray-500">Loading reviews...</div>
          ) : isError ? (
            <div className="text-center py-8 text-red-600">Failed to load reviews</div>
          ) : !reviews || reviews.length === 0 ? (
            <div className="text-center py-8 text-gray-500">No reviews found</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      ID
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Text Preview
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Category
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Confidence
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Created
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {reviews.map((review) => {
                    const topCategory = Object.entries(review.category_scores).sort(
                      ([, a], [, b]) => (b || 0) - (a || 0)
                    )[0];

                    return (
                      <tr key={review.id}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                          {review.id.slice(0, 8)}...
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-900 max-w-xs truncate">
                          {review.text}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 capitalize">
                          {topCategory ? topCategory[0].replace(/_/g, " ") : "N/A"}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {formatPercentage(review.confidence)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <Badge
                            variant={review.status === "pending" ? "warning" : "success"}
                          >
                            {review.status}
                          </Badge>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
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
