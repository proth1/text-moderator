import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import { useAuth } from "../hooks/useAuth";
import { UserRole } from "../types";

export default function Login() {
  const [apiKey, setApiKey] = useState("");
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

      // Mock user object
      const mockUser = {
        id: "user_123",
        email: "moderator@civitas.ai",
        role: UserRole.MODERATOR,
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
                <p>Demo credentials:</p>
                <p className="mt-1">Use any API key to demo the application</p>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
