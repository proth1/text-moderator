import { Link } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Badge from "../components/ui/Badge";
import Button from "../components/ui/Button";
import { usePolicies } from "../hooks/usePolicies";
import { PolicyStatus } from "../types";
import { formatDate } from "../lib/utils";
import { Plus, Edit } from "lucide-react";

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
    <div>
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Policies</h1>
          <p className="mt-2 text-gray-600">Manage moderation policies</p>
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
          <CardTitle>All Policies</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-gray-500">Loading policies...</div>
          ) : isError ? (
            <div className="text-center py-8 text-red-600">Failed to load policies</div>
          ) : !policies || policies.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <p className="mb-4">No policies found</p>
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
                  className="border border-gray-200 rounded-lg p-4 hover:border-blue-300 transition-colors"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <h3 className="text-lg font-semibold text-gray-900" data-testid="policy-name">{policy.name}</h3>
                        <Badge variant={getStatusVariant(policy.status)} data-testid="policy-status">
                          {policy.status.toUpperCase()}
                        </Badge>
                        <span className="text-sm text-gray-500" data-testid="policy-version">v{policy.version}</span>
                      </div>

                      {policy.description && (
                        <p className="text-sm text-gray-600 mb-3">{policy.description}</p>
                      )}

                      <div className="flex items-center gap-4 text-sm text-gray-500">
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
