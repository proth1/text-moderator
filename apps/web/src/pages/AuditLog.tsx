import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Badge from "../components/ui/Badge";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import { formatDate } from "../lib/utils";
import { Download, Search, ScrollText, Shield, Lock } from "lucide-react";
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
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Audit Log</h1>
          <p className="mt-1 text-slate-400">Immutable evidence trail for compliance and review</p>
        </div>
        <Button onClick={handleExport} data-testid="export-button">
          <Download className="h-4 w-4 mr-2" />
          Export
        </Button>
      </div>

      {/* Search */}
      <Card>
        <CardContent className="pt-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">
                Search Records
              </label>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-slate-500" />
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

      {/* Evidence Table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ScrollText className="w-5 h-5 text-emerald-400" />
            Evidence Records
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-12">
              <div className="w-8 h-8 border-2 border-blue-500/30 border-t-blue-500 rounded-full animate-spin mx-auto mb-4" />
              <p className="text-slate-500">Loading evidence records...</p>
            </div>
          ) : isError ? (
            <div className="text-center py-12 text-red-400" data-testid="error-message">
              Failed to load evidence records
            </div>
          ) : !filteredEvidence || filteredEvidence.length === 0 ? (
            <div className="text-center py-12">
              <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-slate-800 flex items-center justify-center">
                <Shield className="w-8 h-8 text-slate-600" />
              </div>
              <p className="text-slate-500">No evidence records found</p>
            </div>
          ) : (
            <div className="overflow-x-auto rounded-lg border border-slate-800">
              <table className="min-w-full divide-y divide-slate-800">
                <thead className="bg-slate-800/50">
                  <tr>
                    <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Control ID
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Decision ID
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Action
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Model
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Immutable
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">
                      Created
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800">
                  {filteredEvidence.map((record) => (
                    <tr key={record.id} data-testid="evidence-record" className="hover:bg-slate-800/30 transition-colors">
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-white" data-testid="control-id">
                        <code className="px-2 py-1 rounded bg-slate-800 text-blue-400">{record.control_id}</code>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-300 font-mono" data-testid="decision-id">
                        {record.decision_id?.slice(0, 8) || "N/A"}...
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap" data-testid="automated-action">
                        <Badge variant="default" className="capitalize">
                          {record.automated_action || "N/A"}
                        </Badge>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-400">
                        {record.model_name || "N/A"}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {record.immutable ? (
                          <Badge variant="success" data-testid="immutable-badge" className="flex items-center gap-1 w-fit">
                            <Lock className="w-3 h-3" />
                            Immutable
                          </Badge>
                        ) : (
                          <Badge variant="secondary">Mutable</Badge>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-500" data-testid="timestamp">
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
