import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Badge from "../components/ui/Badge";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import { formatDate } from "../lib/utils";
import { Download, Search } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { listEvidence, exportEvidence } from "../api/evidence";

export default function AuditLog() {
  const [searchTerm, setSearchTerm] = useState("");

  const { data: evidence, isLoading, isError } = useQuery({
    queryKey: ["evidence"],
    queryFn: () => listEvidence(),
  });

  const handleExport = async () => {
    try {
      const blob = await exportEvidence();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `evidence_export_${new Date().toISOString().slice(0, 10)}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (err) {
      console.error("Export failed:", err);
    }
  };

  const filteredEvidence = evidence?.filter((record) => {
    if (!searchTerm) return true;
    const term = searchTerm.toLowerCase();
    return (
      record.control_id?.toLowerCase().includes(term) ||
      record.decision_id?.toLowerCase().includes(term) ||
      record.automated_action?.toLowerCase().includes(term) ||
      record.submission_hash?.toLowerCase().includes(term)
    );
  });

  return (
    <div>
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Audit Log</h1>
          <p className="mt-2 text-gray-600">Review all moderation actions and decisions</p>
        </div>
        <Button onClick={handleExport} data-testid="export-button">
          <Download className="h-4 w-4 mr-2" />
          Export
        </Button>
      </div>

      <Card className="mb-6">
        <CardContent className="pt-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Search
              </label>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <Input
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder="Search by control ID, decision ID, or action..."
                  className="pl-10"
                  data-testid="search-input"
                />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Evidence Records</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-gray-500">Loading evidence records...</div>
          ) : isError ? (
            <div className="text-center py-8 text-red-600" data-testid="error-message">
              Failed to load evidence records
            </div>
          ) : !filteredEvidence || filteredEvidence.length === 0 ? (
            <div className="text-center py-8 text-gray-500">No evidence records found</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Control ID
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Decision ID
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Action
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Model
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Immutable
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Created
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {filteredEvidence.map((record) => (
                    <tr key={record.id} data-testid="evidence-record">
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900" data-testid="control-id">
                        {record.control_id}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900" data-testid="decision-id">
                        {record.decision_id?.slice(0, 8) || "N/A"}...
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap" data-testid="automated-action">
                        <Badge variant="default" className="capitalize">
                          {record.automated_action || "N/A"}
                        </Badge>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {record.model_name || "N/A"}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {record.immutable ? (
                          <Badge variant="success" data-testid="immutable-badge">Immutable</Badge>
                        ) : (
                          <Badge variant="secondary">Mutable</Badge>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500" data-testid="timestamp">
                        {formatDate(record.created_at)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
