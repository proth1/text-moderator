import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Button from "../components/ui/Button";
import Textarea from "../components/ui/Textarea";
import Badge from "../components/ui/Badge";
import { useModerateText } from "../hooks/useModeration";
import { PolicyAction } from "../types";
import { formatPercentage } from "../lib/utils";
import { AlertTriangle, CheckCircle, XCircle, Zap, Send, RotateCcw } from "lucide-react";

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
        return <CheckCircle className="h-6 w-6 text-emerald-400" />;
      case PolicyAction.WARN:
        return <AlertTriangle className="h-6 w-6 text-yellow-400" />;
      case PolicyAction.BLOCK:
        return <XCircle className="h-6 w-6 text-red-400" />;
      default:
        return null;
    }
  };

  const getScoreColor = (score: number) => {
    if (score < 0.3) return "from-emerald-500 to-emerald-400";
    if (score < 0.7) return "from-yellow-500 to-yellow-400";
    return "from-red-500 to-red-400";
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Moderation Demo</h1>
          <p className="mt-1 text-slate-400">Test content moderation with our ML-powered API</p>
        </div>
        <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-500/10 border border-blue-500/20">
          <Zap className="w-4 h-4 text-blue-400" />
          <span className="text-sm text-blue-400 font-medium">Live API</span>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Input Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Send className="w-5 h-5 text-blue-400" />
              Submit Content
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  Enter text to moderate
                </label>
                <Textarea
                  name="content"
                  value={text}
                  onChange={(e) => setText(e.target.value)}
                  placeholder="Type or paste your message here to analyze for toxicity, spam, harassment, and other harmful content..."
                  rows={8}
                />
              </div>

              <div className="flex gap-3">
                <Button onClick={handleSubmit} disabled={!text.trim() || moderateMutation.isPending}>
                  {moderateMutation.isPending ? (
                    <>
                      <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin mr-2" />
                      Analyzing...
                    </>
                  ) : (
                    <>
                      <Zap className="w-4 h-4 mr-2" />
                      Analyze Content
                    </>
                  )}
                </Button>
                <Button
                  variant="outline"
                  onClick={() => {
                    setText("");
                    moderateMutation.reset();
                  }}
                >
                  <RotateCcw className="w-4 h-4 mr-2" />
                  Clear
                </Button>
              </div>

              {moderateMutation.isError && (
                <div className="rounded-lg bg-red-500/10 border border-red-500/30 p-4" data-testid="error-message">
                  <p className="text-sm text-red-400">
                    Failed to moderate text. Please try again.
                  </p>
                </div>
              )}

              {/* Sample texts */}
              <div className="pt-4 border-t border-slate-800">
                <p className="text-xs text-slate-500 mb-2">Try sample texts:</p>
                <div className="flex flex-wrap gap-2">
                  <button
                    onClick={() => setText("This is a helpful and friendly message about the weather today.")}
                    className="text-xs px-3 py-1.5 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 hover:bg-emerald-500/20 transition-colors"
                  >
                    Safe content
                  </button>
                  <button
                    onClick={() => setText("You are absolutely terrible and should be ashamed of yourself!")}
                    className="text-xs px-3 py-1.5 rounded-lg bg-red-500/10 text-red-400 border border-red-500/30 hover:bg-red-500/20 transition-colors"
                  >
                    Toxic content
                  </button>
                  <button
                    onClick={() => setText("Buy cheap products now! Click here for amazing deals! Limited time offer!")}
                    className="text-xs px-3 py-1.5 rounded-lg bg-yellow-500/10 text-yellow-400 border border-yellow-500/30 hover:bg-yellow-500/20 transition-colors"
                  >
                    Spam content
                  </button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Results Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <CheckCircle className="w-5 h-5 text-emerald-400" />
              Moderation Results
            </CardTitle>
          </CardHeader>
          <CardContent>
            {!result ? (
              <div className="text-center py-16">
                <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-slate-800 flex items-center justify-center">
                  <Zap className="w-8 h-8 text-slate-600" />
                </div>
                <p className="text-slate-500">Submit content to see moderation results</p>
                <p className="text-xs text-slate-600 mt-2">Results will appear here instantly</p>
              </div>
            ) : (
              <div className="space-y-6" data-testid="moderation-feedback">
                {/* Decision */}
                <div data-testid="moderation-action" className="p-4 rounded-xl bg-slate-800/50 border border-slate-700">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      {getActionIcon(result.action)}
                      <div>
                        <span className="text-sm text-slate-400">Decision</span>
                        <div className="mt-1">
                          <Badge variant={getActionColor(result.action)} className="text-base px-4 py-1.5">
                            {result.action.toUpperCase()}
                          </Badge>
                        </div>
                      </div>
                    </div>
                    <div className="text-right">
                      <span className="text-sm text-slate-400">Confidence</span>
                      <div className="text-2xl font-bold text-white">
                        {formatPercentage(result.confidence)}
                      </div>
                    </div>
                  </div>
                </div>

                {/* Confidence Bar */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-slate-400">Overall Confidence</span>
                    <span className="text-sm font-medium text-white">{formatPercentage(result.confidence)}</span>
                  </div>
                  <div className="h-3 rounded-full bg-slate-800 overflow-hidden">
                    <div
                      className="h-full rounded-full bg-gradient-to-r from-blue-500 to-purple-500 transition-all duration-500"
                      style={{ width: `${result.confidence * 100}%` }}
                    />
                  </div>
                </div>

                {/* Category Scores */}
                <div data-testid="category-scores">
                  <div className="text-sm font-medium text-slate-300 mb-3">Category Scores</div>
                  <div className="space-y-3">
                    {Object.entries(result.category_scores).map(([category, score]) => (
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
                {result.reasons.length > 0 && (
                  <div>
                    <div className="text-sm font-medium text-slate-300 mb-2">Reasons</div>
                    <ul className="space-y-1">
                      {result.reasons.map((reason, idx) => (
                        <li key={idx} className="text-sm text-slate-400 flex items-start gap-2">
                          <span className="text-blue-400 mt-1">•</span>
                          {reason}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {/* Review Flag */}
                {result.requires_review && (
                  <div className="rounded-lg bg-yellow-500/10 border border-yellow-500/30 p-4 flex items-center gap-3">
                    <AlertTriangle className="w-5 h-5 text-yellow-400 flex-shrink-0" />
                    <p className="text-sm text-yellow-300">
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
