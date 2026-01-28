import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import Textarea from "../components/ui/Textarea";
import { usePolicy, useCreatePolicy, useUpdatePolicy, usePublishPolicy } from "../hooks/usePolicies";
import { PolicyAction } from "../types";
import type { CategoryScores } from "../types";
import { ArrowLeft } from "lucide-react";

const CATEGORIES = [
  "hate_speech",
  "harassment",
  "violence",
  "sexual_content",
  "spam",
  "misinformation",
  "pii",
];

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

  return (
    <div>
      <div className="mb-6">
        <Button variant="ghost" onClick={() => navigate("/policies")}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Policies
        </Button>
      </div>

      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">
          {isNew ? "Create New Policy" : "Edit Policy"}
        </h1>
        <p className="mt-2 text-gray-600">
          Configure category thresholds and actions
        </p>
      </div>

      <div className="max-w-4xl space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Basic Information</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Policy Name
                </label>
                <Input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g., Standard Moderation Policy"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
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
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Effective Date (Optional)
                </label>
                <Input
                  type="date"
                  value={effectiveDate}
                  onChange={(e) => setEffectiveDate(e.target.value)}
                />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Category Configuration</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-6">
              {CATEGORIES.map((category) => (
                <div key={category} className="border-b border-gray-200 pb-6 last:border-0">
                  <h4 className="text-sm font-medium text-gray-900 mb-4 capitalize">
                    {category.replace(/_/g, " ")}
                  </h4>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Threshold
                      </label>
                      <div className="flex items-center gap-4">
                        <input
                          type="range"
                          min="0"
                          max="1"
                          step="0.01"
                          value={thresholds[category] || 0.5}
                          onChange={(e) =>
                            handleThresholdChange(category, parseFloat(e.target.value))
                          }
                          className="flex-1"
                        />
                        <span className="text-sm font-medium w-12 text-right">
                          {((thresholds[category] || 0.5) * 100).toFixed(0)}%
                        </span>
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Action
                      </label>
                      <select
                        value={actions[category] || PolicyAction.ALLOW}
                        onChange={(e) =>
                          handleActionChange(category, e.target.value as PolicyAction)
                        }
                        className="w-full h-10 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-600"
                      >
                        <option value={PolicyAction.ALLOW}>Allow</option>
                        <option value={PolicyAction.WARN}>Warn</option>
                        <option value={PolicyAction.BLOCK}>Block</option>
                        <option value={PolicyAction.REVIEW}>Review</option>
                      </select>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <div className="flex gap-3">
          <Button
            onClick={() => handleSave(false)}
            disabled={!name || createPolicy.isPending || updatePolicy.isPending}
          >
            Save as Draft
          </Button>
          <Button
            variant="default"
            onClick={() => handleSave(true)}
            disabled={!name || createPolicy.isPending || updatePolicy.isPending || publishPolicy.isPending}
          >
            Save & Publish
          </Button>
          <Button variant="outline" onClick={() => navigate("/policies")}>
            Cancel
          </Button>
        </div>
      </div>
    </div>
  );
}
