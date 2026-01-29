import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Badge from "../components/ui/Badge";
import Button from "../components/ui/Button";
import Textarea from "../components/ui/Textarea";
import { useReviewDetail, useSubmitReviewAction } from "../hooks/useReviews";
import { ReviewActionType } from "../types";
import { formatDate, formatPercentage } from "../lib/utils";
import { ArrowLeft, FileText, Brain, Gavel, History, CheckCircle, XCircle, ArrowUpRight } from "lucide-react";

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

  const getScoreColor = (score: number) => {
    if (score < 0.3) return "from-emerald-500 to-emerald-400";
    if (score < 0.7) return "from-yellow-500 to-yellow-400";
    return "from-red-500 to-red-400";
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <div className="w-8 h-8 border-2 border-blue-500/30 border-t-blue-500 rounded-full animate-spin mx-auto mb-4" />
          <p className="text-slate-500">Loading review...</p>
        </div>
      </div>
    );
  }

  if (isError || !review) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <XCircle className="w-12 h-12 text-red-400 mx-auto mb-4" />
          <p className="text-red-400">Failed to load review</p>
          <Button variant="outline" onClick={() => navigate("/reviews")} className="mt-4">
            Return to Queue
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Back button */}
      <div>
        <Button variant="ghost" onClick={() => navigate("/reviews")}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Queue
        </Button>
      </div>

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Review Detail</h1>
          <p className="mt-1 text-slate-400">
            Decision ID: <code className="px-2 py-0.5 rounded bg-slate-800 text-blue-400">{review.decision_id.slice(0, 12)}...</code>
          </p>
        </div>
        <Badge variant={review.status === "pending" ? "warning" : "success"} className="text-sm px-4 py-1.5">
          {review.status.toUpperCase()}
        </Badge>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          {/* Original Content */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <FileText className="w-5 h-5 text-purple-400" />
                Original Content
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="bg-slate-800/50 border border-slate-700 p-4 rounded-lg">
                <p className="text-sm text-slate-200 whitespace-pre-wrap leading-relaxed">{review.text}</p>
              </div>
              <div className="mt-4 flex flex-wrap gap-2">
                <Badge variant="secondary">
                  Submission: {review.submission_id.slice(0, 12)}...
                </Badge>
                <Badge variant="secondary">
                  {formatDate(review.created_at)}
                </Badge>
              </div>
            </CardContent>
          </Card>

          {/* Model Analysis */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Brain className="w-5 h-5 text-blue-400" />
                Model Analysis
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-6">
                {/* Suggested Action */}
                <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700">
                  <div className="flex items-center justify-between">
                    <div>
                      <span className="text-sm text-slate-400">Suggested Action</span>
                      <div className="mt-2">
                        <Badge variant="warning" className="text-base px-4 py-1.5">
                          {review.suggested_action.toUpperCase()}
                        </Badge>
                      </div>
                    </div>
                    <div className="text-right">
                      <span className="text-sm text-slate-400">Confidence</span>
                      <div className="text-2xl font-bold text-white mt-1">
                        {formatPercentage(review.confidence)}
                      </div>
                    </div>
                  </div>
                </div>

                {/* Confidence Bar */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-slate-400">Confidence Score</span>
                    <span className="text-sm font-medium text-white">{formatPercentage(review.confidence)}</span>
                  </div>
                  <div className="h-3 rounded-full bg-slate-800 overflow-hidden">
                    <div
                      className="h-full rounded-full bg-gradient-to-r from-blue-500 to-purple-500 transition-all duration-500"
                      style={{ width: `${review.confidence * 100}%` }}
                    />
                  </div>
                </div>

                {/* Category Scores */}
                <div>
                  <div className="text-sm font-medium text-slate-300 mb-3">Category Scores</div>
                  <div className="space-y-3">
                    {Object.entries(review.category_scores)
                      .sort(([, a], [, b]) => (b || 0) - (a || 0))
                      .map(([category, score]) => (
                        <div key={category}>
                          <div className="flex items-center justify-between mb-1">
                            <span className="text-sm text-slate-400 capitalize">
                              {category.replace(/_/g, " ")}
                            </span>
                            <span className="text-sm font-medium text-white">
                              {formatPercentage(score || 0)}
                            </span>
                          </div>
                          <div className="h-2 rounded-full bg-slate-800 overflow-hidden">
                            <div
                              className={`h-full rounded-full bg-gradient-to-r ${getScoreColor(score || 0)} transition-all duration-500`}
                              style={{ width: `${(score || 0) * 100}%` }}
                            />
                          </div>
                        </div>
                      ))}
                  </div>
                </div>

                {/* Reasons */}
                {review.reasons.length > 0 && (
                  <div>
                    <div className="text-sm font-medium text-slate-300 mb-2">Reasons</div>
                    <ul className="space-y-1">
                      {review.reasons.map((reason, idx) => (
                        <li key={idx} className="text-sm text-slate-400 flex items-start gap-2">
                          <span className="text-blue-400 mt-1">•</span>
                          {reason}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {/* Policy Triggered */}
                {review.policy_name && (
                  <div className="p-3 rounded-lg bg-purple-500/10 border border-purple-500/30">
                    <div className="text-sm font-medium text-purple-400 mb-1">Policy Triggered</div>
                    <p className="text-sm text-slate-300">{review.policy_name}</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Action Panel */}
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Gavel className="w-5 h-5 text-emerald-400" />
                Take Action
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-2">
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
                    className="w-full justify-center"
                    onClick={() => handleAction(ReviewActionType.APPROVE)}
                    disabled={submitAction.isPending}
                  >
                    <CheckCircle className="w-4 h-4 mr-2" />
                    Approve Model Decision
                  </Button>

                  <Button
                    className="w-full justify-center"
                    variant="secondary"
                    onClick={() => handleAction(ReviewActionType.OVERRIDE_ALLOW)}
                    disabled={submitAction.isPending}
                  >
                    <CheckCircle className="w-4 h-4 mr-2" />
                    Override: Allow
                  </Button>

                  <Button
                    className="w-full justify-center"
                    variant="destructive"
                    onClick={() => handleAction(ReviewActionType.OVERRIDE_BLOCK)}
                    disabled={submitAction.isPending}
                  >
                    <XCircle className="w-4 h-4 mr-2" />
                    Override: Block
                  </Button>

                  <Button
                    className="w-full justify-center"
                    variant="outline"
                    onClick={() => handleAction(ReviewActionType.ESCALATE)}
                    disabled={submitAction.isPending}
                  >
                    <ArrowUpRight className="w-4 h-4 mr-2" />
                    Escalate to Senior Moderator
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Action History */}
          {review.actions.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <History className="w-5 h-5 text-slate-400" />
                  Action History
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {review.actions.map((action) => (
                    <div key={action.id} className="border-l-2 border-blue-500 pl-3 pb-3">
                      <div className="flex items-center gap-2">
                        <Badge variant="outline" className="capitalize text-xs">
                          {action.action_type.replace(/_/g, " ")}
                        </Badge>
                      </div>
                      <div className="text-xs text-slate-500 mt-1">
                        {formatDate(action.created_at)}
                      </div>
                      {action.rationale && (
                        <p className="text-sm text-slate-400 mt-2 bg-slate-800/50 p-2 rounded">
                          {action.rationale}
                        </p>
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
