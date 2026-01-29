import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import Textarea from "../components/ui/Textarea";
import { usePolicy, useCreatePolicy, useUpdatePolicy, usePublishPolicy } from "../hooks/usePolicies";
import { PolicyAction } from "../types";
import type { CategoryScores } from "../types";
import { ArrowLeft, FileText, Settings, Save, Send, X } from "lucide-react";

const CATEGORIES = [
  "hate_speech",
  "harassment",
  "violence",
  "sexual_content",
  "spam",
  "misinformation",
  "pii",
];

const ACTION_COLORS: Record<PolicyAction, string> = {
  [PolicyAction.ALLOW]: "bg-emerald-500/20 border-emerald-500/30 text-emerald-400",
  [PolicyAction.WARN]: "bg-yellow-500/20 border-yellow-500/30 text-yellow-400",
  [PolicyAction.BLOCK]: "bg-red-500/20 border-red-500/30 text-red-400",
  [PolicyAction.REVIEW]: "bg-blue-500/20 border-blue-500/30 text-blue-400",
};

export default function PolicyEditor() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isNew = !id;

  const { data: existingPolicy } = usePolicy(id || "");
  const createPolicy = useCreatePolicy();
  const updatePolicy = useUpdatePolicy();
  const publishPolicy = usePublishPolicy();

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [thresholds, setThresholds] = useState<CategoryScores>({});
  const [actions, setActions] = useState<Record<string, PolicyAction>>({});
  const [effectiveDate, setEffectiveDate] = useState("");

  useEffect(() => {
    if (existingPolicy) {
      setName(existingPolicy.name);
      setDescription(existingPolicy.description || "");
      setThresholds(existingPolicy.category_thresholds);
      setActions(existingPolicy.category_actions);
      setEffectiveDate(existingPolicy.effective_date || "");
    }
  }, [existingPolicy]);

  const handleThresholdChange = (category: string, value: number) => {
    setThresholds((prev) => ({ ...prev, [category]: value }));
  };

  const handleActionChange = (category: string, action: PolicyAction) => {
    setActions((prev) => ({ ...prev, [category]: action }));
  };

  const handleSave = async (publish = false) => {
    const policyData = {
      name,
      description: description || undefined,
      category_thresholds: thresholds,
      category_actions: actions,
      effective_date: effectiveDate || undefined,
    };

    try {
      if (isNew) {
        const newPolicy = await createPolicy.mutateAsync(policyData);
        if (publish) {
          await publishPolicy.mutateAsync(newPolicy.id);
        }
      } else if (id) {
        await updatePolicy.mutateAsync({ id, policy: policyData });
        if (publish) {
          await publishPolicy.mutateAsync(id);
        }
      }
      navigate("/policies");
    } catch (error) {
      console.error("Failed to save policy:", error);
    }
  };

  const getThresholdColor = (value: number) => {
    if (value < 0.3) return "from-emerald-500 to-emerald-400";
    if (value < 0.7) return "from-yellow-500 to-yellow-400";
    return "from-red-500 to-red-400";
  };

  return (
    <div className="space-y-6">
      {/* Back button */}
      <div>
        <Button variant="ghost" onClick={() => navigate("/policies")}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Policies
        </Button>
      </div>

      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-white">
          {isNew ? "Create New Policy" : "Edit Policy"}
        </h1>
        <p className="mt-1 text-slate-400">
          Configure category thresholds and automated actions
        </p>
      </div>

      <div className="max-w-4xl space-y-6">
        {/* Basic Information */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="w-5 h-5 text-purple-400" />
              Basic Information
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  Policy Name
                </label>
                <Input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g., Standard Moderation Policy"
                  data-testid="policy-name-input"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  Description
                </label>
                <Textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Describe the purpose of this policy..."
                  rows={3}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  Effective Date (Optional)
                </label>
                <Input
                  type="date"
                  value={effectiveDate}
                  onChange={(e) => setEffectiveDate(e.target.value)}
                  className="max-w-xs"
                />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Category Configuration */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Settings className="w-5 h-5 text-blue-400" />
              Category Configuration
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-6">
              {CATEGORIES.map((category) => (
                <div key={category} className="border-b border-slate-800 pb-6 last:border-0 last:pb-0">
                  <h4 className="text-sm font-semibold text-white mb-4 capitalize flex items-center gap-2">
                    <div className="w-2 h-2 rounded-full bg-blue-500" />
                    {category.replace(/_/g, " ")}
                  </h4>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    {/* Threshold Slider */}
                    <div>
                      <label className="block text-sm font-medium text-slate-400 mb-3">
                        Threshold
                      </label>
                      <div className="space-y-2">
                        <input
                          type="range"
                          min="0"
                          max="1"
                          step="0.01"
                          value={thresholds[category] || 0.5}
                          onChange={(e) =>
                            handleThresholdChange(category, parseFloat(e.target.value))
                          }
                          className="w-full h-2 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-blue-500"
                          data-testid={`threshold-${category}`}
                        />
                        <div className="flex items-center justify-between">
                          <div className="h-2 flex-1 rounded-full bg-slate-800 overflow-hidden mr-4">
                            <div
                              className={`h-full rounded-full bg-gradient-to-r ${getThresholdColor(thresholds[category] || 0.5)} transition-all duration-300`}
                              style={{ width: `${(thresholds[category] || 0.5) * 100}%` }}
                            />
                          </div>
                          <span className="text-sm font-mono font-medium text-white w-12 text-right">
                            {((thresholds[category] || 0.5) * 100).toFixed(0)}%
                          </span>
                        </div>
                      </div>
                    </div>

                    {/* Action Selector */}
                    <div>
                      <label className="block text-sm font-medium text-slate-400 mb-3">
                        Action When Exceeded
                      </label>
                      <div className="grid grid-cols-2 gap-2">
                        {Object.values(PolicyAction).map((action) => (
                          <button
                            key={action}
                            type="button"
                            onClick={() => handleActionChange(category, action)}
                            className={`px-3 py-2 text-sm font-medium rounded-lg border transition-all ${
                              actions[category] === action || (!actions[category] && action === PolicyAction.ALLOW)
                                ? ACTION_COLORS[action]
                                : "bg-slate-800/50 border-slate-700 text-slate-400 hover:border-slate-600"
                            }`}
                            data-testid={`action-${category}-${action}`}
                          >
                            {action.charAt(0).toUpperCase() + action.slice(1)}
                          </button>
                        ))}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Action Buttons */}
        <div className="flex gap-3 pb-8">
          <Button
            onClick={() => handleSave(false)}
            disabled={!name || createPolicy.isPending || updatePolicy.isPending}
            variant="secondary"
            data-testid="save-draft-button"
          >
            <Save className="w-4 h-4 mr-2" />
            Save as Draft
          </Button>
          <Button
            onClick={() => handleSave(true)}
            disabled={!name || createPolicy.isPending || updatePolicy.isPending || publishPolicy.isPending}
            data-testid="save-publish-button"
          >
            <Send className="w-4 h-4 mr-2" />
            Save & Publish
          </Button>
          <Button variant="outline" onClick={() => navigate("/policies")}>
            <X className="w-4 h-4 mr-2" />
            Cancel
          </Button>
        </div>
      </div>
    </div>
  );
}
