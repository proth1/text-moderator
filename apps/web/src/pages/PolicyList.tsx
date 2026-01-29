import { Link } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Badge from "../components/ui/Badge";
import Button from "../components/ui/Button";
import { usePolicies } from "../hooks/usePolicies";
import { PolicyStatus } from "../types";
import { formatDate } from "../lib/utils";
import { Plus, Edit, FileText, Shield } from "lucide-react";

export default function PolicyList() {
  const { data: policies, isLoading, isError } = usePolicies();

  const getStatusVariant = (status: PolicyStatus) => {
    switch (status) {
      case PolicyStatus.PUBLISHED:
        return "success";
      case PolicyStatus.DRAFT:
        return "secondary";
      case PolicyStatus.ARCHIVED:
        return "outline";
      default:
        return "default";
    }
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Policies</h1>
          <p className="mt-1 text-slate-400">Configure moderation rules and thresholds</p>
        </div>
        <Link to="/policies/new">
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            New Policy
          </Button>
        </Link>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="w-5 h-5 text-purple-400" />
            All Policies
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-12">
              <div className="w-8 h-8 border-2 border-blue-500/30 border-t-blue-500 rounded-full animate-spin mx-auto mb-4" />
              <p className="text-slate-500">Loading policies...</p>
            </div>
          ) : isError ? (
            <div className="text-center py-12 text-red-400">Failed to load policies</div>
          ) : !policies || policies.length === 0 ? (
            <div className="text-center py-12">
              <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-slate-800 flex items-center justify-center">
                <Shield className="w-8 h-8 text-slate-600" />
              </div>
              <p className="text-slate-500 mb-4">No policies found</p>
              <Link to="/policies/new">
                <Button>Create your first policy</Button>
              </Link>
            </div>
          ) : (
            <div className="space-y-4">
              {policies.map((policy) => (
                <div
                  key={policy.id}
                  data-testid="policy-item"
                  className="border border-slate-700 rounded-xl p-5 hover:border-blue-500/50 hover:bg-slate-800/50 transition-all group"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <h3 className="text-lg font-semibold text-white" data-testid="policy-name">{policy.name}</h3>
                        <Badge variant={getStatusVariant(policy.status)} data-testid="policy-status">
                          {policy.status.toUpperCase()}
                        </Badge>
                        <span className="text-sm text-slate-500" data-testid="policy-version">v{policy.version}</span>
                      </div>

                      {policy.description && (
                        <p className="text-sm text-slate-400 mb-3">{policy.description}</p>
                      )}

                      <div className="flex items-center gap-4 text-sm text-slate-500">
                        <span>Created: {formatDate(policy.created_at)}</span>
                        {policy.published_at && (
                          <span>Published: {formatDate(policy.published_at)}</span>
                        )}
                        {policy.effective_date && (
                          <span>Effective: {formatDate(policy.effective_date)}</span>
                        )}
                      </div>

                      <div className="mt-3 flex flex-wrap gap-2">
                        {Object.entries(policy.category_actions).map(([category, action]) => (
                          <Badge key={category} variant="outline" className="text-xs">
                            {category.replace(/_/g, " ")}: {action}
                          </Badge>
                        ))}
                      </div>
                    </div>

                    <div className="ml-4">
                      <Link to={`/policies/${policy.id}/edit`}>
                        <Button size="sm" variant="ghost">
                          <Edit className="h-4 w-4 mr-1" />
                          Edit
                        </Button>
                      </Link>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
