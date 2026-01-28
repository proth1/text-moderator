import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Badge from "../components/ui/Badge";
import Button from "../components/ui/Button";
import Textarea from "../components/ui/Textarea";
import { useReviewDetail, useSubmitReviewAction } from "../hooks/useReviews";
import { ReviewActionType } from "../types";
import { formatDate, formatPercentage } from "../lib/utils";
import { ArrowLeft } from "lucide-react";

export default function ReviewDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [rationale, setRationale] = useState("");

  const { data: review, isLoading, isError } = useReviewDetail(id!);
  const submitAction = useSubmitReviewAction();

  const handleAction = async (actionType: ReviewActionType) => {
    if (!id) return;

    await submitAction.mutateAsync({
      id,
      action: actionType,
      rationale: rationale.trim() || undefined,
    });

    navigate("/reviews");
  };

  if (isLoading) {
    return <div className="text-center py-8">Loading review...</div>;
  }

  if (isError || !review) {
    return <div className="text-center py-8 text-red-600">Failed to load review</div>;
  }

  return (
    <div>
      <div className="mb-6">
        <Button variant="ghost" onClick={() => navigate("/reviews")}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Queue
        </Button>
      </div>

      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Review Detail</h1>
        <p className="mt-2 text-gray-600">Decision ID: {review.decision_id}</p>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Original Content</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="bg-gray-50 p-4 rounded-md">
                <p className="text-sm text-gray-900 whitespace-pre-wrap">{review.text}</p>
              </div>
              <div className="mt-4 flex gap-2">
                <Badge variant="secondary">Submission ID: {review.submission_id.slice(0, 12)}...</Badge>
                <Badge variant="secondary">{formatDate(review.created_at)}</Badge>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Model Analysis</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <div className="text-sm font-medium text-gray-700 mb-2">
                    Suggested Action
                  </div>
                  <Badge variant="warning">{review.suggested_action.toUpperCase()}</Badge>
                </div>

                <div>
                  <div className="text-sm font-medium text-gray-700 mb-2">
                    Confidence Score
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="flex-1 bg-gray-200 rounded-full h-2">
                      <div
                        className="bg-blue-600 h-2 rounded-full"
                        style={{ width: `${review.confidence * 100}%` }}
                      />
                    </div>
                    <span className="text-sm font-medium">
                      {formatPercentage(review.confidence)}
                    </span>
                  </div>
                </div>

                <div>
                  <div className="text-sm font-medium text-gray-700 mb-3">Category Scores</div>
                  <div className="space-y-2">
                    {Object.entries(review.category_scores)
                      .sort(([, a], [, b]) => (b || 0) - (a || 0))
                      .map(([category, score]) => (
                        <div key={category} className="flex items-center justify-between">
                          <span className="text-sm text-gray-600 capitalize">
                            {category.replace(/_/g, " ")}
                          </span>
                          <div className="flex items-center gap-2">
                            <div className="w-32 bg-gray-200 rounded-full h-1.5">
                              <div
                                className="bg-blue-600 h-1.5 rounded-full"
                                style={{ width: `${(score || 0) * 100}%` }}
                              />
                            </div>
                            <span className="text-sm font-medium w-12 text-right">
                              {formatPercentage(score || 0)}
                            </span>
                          </div>
                        </div>
                      ))}
                  </div>
                </div>

                {review.reasons.length > 0 && (
                  <div>
                    <div className="text-sm font-medium text-gray-700 mb-2">Reasons</div>
                    <ul className="list-disc list-inside space-y-1">
                      {review.reasons.map((reason, idx) => (
                        <li key={idx} className="text-sm text-gray-600">
                          {reason}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {review.policy_name && (
                  <div>
                    <div className="text-sm font-medium text-gray-700 mb-2">Policy Triggered</div>
                    <p className="text-sm text-gray-600">{review.policy_name}</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Take Action</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Rationale (Optional)
                  </label>
                  <Textarea
                    value={rationale}
                    onChange={(e) => setRationale(e.target.value)}
                    placeholder="Explain your decision..."
                    rows={4}
                  />
                </div>

                <div className="space-y-2">
                  <Button
                    className="w-full"
                    onClick={() => handleAction(ReviewActionType.APPROVE)}
                    disabled={submitAction.isPending}
                  >
                    Approve Model Decision
                  </Button>

                  <Button
                    className="w-full"
                    variant="secondary"
                    onClick={() => handleAction(ReviewActionType.OVERRIDE_ALLOW)}
                    disabled={submitAction.isPending}
                  >
                    Override: Allow
                  </Button>

                  <Button
                    className="w-full"
                    variant="destructive"
                    onClick={() => handleAction(ReviewActionType.OVERRIDE_BLOCK)}
                    disabled={submitAction.isPending}
                  >
                    Override: Block
                  </Button>

                  <Button
                    className="w-full"
                    variant="outline"
                    onClick={() => handleAction(ReviewActionType.ESCALATE)}
                    disabled={submitAction.isPending}
                  >
                    Escalate to Senior Moderator
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>

          {review.actions.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>Action History</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {review.actions.map((action) => (
                    <div key={action.id} className="border-l-2 border-blue-600 pl-3 pb-3">
                      <div className="text-sm font-medium text-gray-900 capitalize">
                        {action.action_type.replace(/_/g, " ")}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {formatDate(action.created_at)}
                      </div>
                      {action.rationale && (
                        <p className="text-sm text-gray-600 mt-2">{action.rationale}</p>
                      )}
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
