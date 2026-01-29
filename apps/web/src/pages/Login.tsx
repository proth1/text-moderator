import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import { useAuth } from "../hooks/useAuth";
import { UserRole } from "../types";

export default function Login() {
  const [apiKey, setApiKey] = useState("tk_admin_test_key_001");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const { login, isAuthenticated } = useAuth();
  const navigate = useNavigate();

  // Redirect to dashboard if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      navigate("/", { replace: true });
    }
  }, [isAuthenticated, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    try {
      // In a real app, this would validate the API key with the backend
      // For now, we'll just simulate a login
      if (!apiKey.trim()) {
        setError("Please enter an API key");
        setIsLoading(false);
        return;
      }

      // User object based on API key
      const usersByApiKey: Record<string, { id: string; email: string; role: UserRole }> = {
        "tk_admin_test_key_001": { id: "a0000000-0000-0000-0000-000000000001", email: "admin@civitas.test", role: UserRole.ADMIN },
        "tk_mod_test_key_002": { id: "a0000000-0000-0000-0000-000000000002", email: "moderator@civitas.test", role: UserRole.MODERATOR },
        "tk_viewer_test_key_003": { id: "a0000000-0000-0000-0000-000000000003", email: "viewer@civitas.test", role: UserRole.VIEWER },
      };

      const userInfo = usersByApiKey[apiKey] || {
        id: "a0000000-0000-0000-0000-000000000001",
        email: "admin@civitas.test",
        role: UserRole.ADMIN
      };

      const mockUser = {
        ...userInfo,
        created_at: new Date().toISOString(),
      };

      login(apiKey, mockUser);
      navigate("/");
    } catch (err) {
      setError("Invalid API key");
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="max-w-md w-full">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-blue-600 mb-2">Civitas AI</h1>
          <p className="text-gray-600">Text Moderator</p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Sign In</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  API Key
                </label>
                <Input
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  placeholder="Enter your API key"
                  disabled={isLoading}
                />
              </div>

              {error && (
                <div className="rounded-md bg-red-50 p-3">
                  <p className="text-sm text-red-800">{error}</p>
                </div>
              )}

              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? "Signing in..." : "Sign In"}
              </Button>

              <div className="mt-4 text-sm text-gray-600">
                <p className="font-medium">Demo credentials:</p>
                <p className="mt-1 font-mono text-xs bg-gray-100 p-2 rounded">tk_admin_test_key_001</p>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
