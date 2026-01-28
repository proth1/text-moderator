import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Button from "../components/ui/Button";
import Textarea from "../components/ui/Textarea";
import Badge from "../components/ui/Badge";
import { useModerateText } from "../hooks/useModeration";
import { PolicyAction } from "../types";
import { formatPercentage } from "../lib/utils";
import { AlertTriangle, CheckCircle, XCircle } from "lucide-react";

export default function ModerationDemo() {
  const [text, setText] = useState("");
  const moderateMutation = useModerateText();

  const handleSubmit = () => {
    if (!text.trim()) return;
    moderateMutation.mutate({ text });
  };

  const result = moderateMutation.data;

  const getActionColor = (action: PolicyAction) => {
    switch (action) {
      case PolicyAction.ALLOW:
        return "success";
      case PolicyAction.WARN:
        return "warning";
      case PolicyAction.BLOCK:
        return "destructive";
      case PolicyAction.REVIEW:
        return "default";
      default:
        return "secondary";
    }
  };

  const getActionIcon = (action: PolicyAction) => {
    switch (action) {
      case PolicyAction.ALLOW:
        return <CheckCircle className="h-5 w-5 text-green-600" />;
      case PolicyAction.WARN:
        return <AlertTriangle className="h-5 w-5 text-yellow-600" />;
      case PolicyAction.BLOCK:
        return <XCircle className="h-5 w-5 text-red-600" />;
      default:
        return null;
    }
  };

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Moderation Demo</h1>
        <p className="mt-2 text-gray-600">Test content moderation in real-time</p>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Submit Content</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Enter text to moderate
                </label>
                <Textarea
                  name="content"
                  value={text}
                  onChange={(e) => setText(e.target.value)}
                  placeholder="Type your message here..."
                  rows={8}
                />
              </div>

              <div className="flex gap-3">
                <Button onClick={handleSubmit} disabled={!text.trim() || moderateMutation.isPending}>
                  {moderateMutation.isPending ? "Analyzing..." : "Submit for Moderation"}
                </Button>
                <Button
                  variant="outline"
                  onClick={() => {
                    setText("");
                    moderateMutation.reset();
                  }}
                >
                  Clear
                </Button>
              </div>

              {moderateMutation.isError && (
                <div className="rounded-md bg-red-50 p-4" data-testid="error-message">
                  <p className="text-sm text-red-800">
                    Failed to moderate text. Please try again.
                  </p>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Moderation Results</CardTitle>
          </CardHeader>
          <CardContent>
            {!result ? (
              <div className="text-center py-12 text-gray-500">
                <p>Submit content to see moderation results</p>
              </div>
            ) : (
              <div className="space-y-6" data-testid="moderation-feedback">
                {/* Action */}
                <div data-testid="moderation-action">
                  <div className="flex items-center gap-2 mb-2">
                    {getActionIcon(result.action)}
                    <span className="text-sm font-medium text-gray-700">Decision</span>
                  </div>
                  <Badge variant={getActionColor(result.action)} className="text-base px-3 py-1">
                    {result.action.toUpperCase()}
                  </Badge>
                </div>

                {/* Confidence */}
                <div>
                  <div className="text-sm font-medium text-gray-700 mb-2">
                    Confidence Score
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="flex-1 bg-gray-200 rounded-full h-2">
                      <div
                        className="bg-blue-600 h-2 rounded-full"
                        style={{ width: `${result.confidence * 100}%` }}
                      />
                    </div>
                    <span className="text-sm font-medium">
                      {formatPercentage(result.confidence)}
                    </span>
                  </div>
                </div>

                {/* Category Scores */}
                <div data-testid="category-scores">
                  <div className="text-sm font-medium text-gray-700 mb-3">Category Scores</div>
                  <div className="space-y-2">
                    {Object.entries(result.category_scores).map(([category, score]) => (
                      <div key={category} className="flex items-center justify-between">
                        <span className="text-sm text-gray-600 capitalize">
                          {category.replace(/_/g, " ")}
                        </span>
                        <div className="flex items-center gap-2">
                          <div className="w-24 bg-gray-200 rounded-full h-1.5">
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

                {/* Reasons */}
                {result.reasons.length > 0 && (
                  <div>
                    <div className="text-sm font-medium text-gray-700 mb-2">Reasons</div>
                    <ul className="list-disc list-inside space-y-1">
                      {result.reasons.map((reason, idx) => (
                        <li key={idx} className="text-sm text-gray-600">
                          {reason}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {/* Review Flag */}
                {result.requires_review && (
                  <div className="rounded-md bg-yellow-50 p-3">
                    <p className="text-sm text-yellow-800">
                      This content has been flagged for human review.
                    </p>
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
