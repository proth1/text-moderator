/**
 * Civitas Presentation - Authentication Middleware
 *
 * Validates sessions from the Civitas Access Gate.
 * Users must authenticate through the gate to access this site.
 */

const GATE_URL = 'https://civitas-access-gate.paul-roth.workers.dev';
const SESSION_DURATION = 24 * 60 * 60; // 24 hours

interface Env {
  // No env vars needed - we call the gate worker for validation
}

export const onRequest: PagesFunction<Env> = async (context) => {
  const { request } = context;
  const url = new URL(request.url);

  // Allow static assets through without auth
  const staticExtensions = ['.js', '.css', '.png', '.jpg', '.jpeg', '.gif', '.svg', '.ico', '.woff', '.woff2', '.ttf', '.eot', '.html'];
  if (staticExtensions.some(ext => url.pathname.endsWith(ext)) && url.pathname !== '/index.html' && url.pathname !== '/') {
    return context.next();
  }

  // Check for auth token in URL (coming from gate after magic link)
  const authToken = url.searchParams.get('civitas_auth');
  if (authToken) {
    // Validate token with gate worker
    const validateResponse = await fetch(`${GATE_URL}/validate-token?token=${authToken}`);
    const validateData = await validateResponse.json() as { valid: boolean; email?: string };

    if (validateData.valid) {
      // Token is valid - set session cookie and redirect to clean URL
      url.searchParams.delete('civitas_auth');

      const response = new Response(null, {
        status: 302,
        headers: {
          'Location': url.toString(),
          'Set-Cookie': `civitas_session=${authToken}; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=${SESSION_DURATION}`,
        },
      });
      return response;
    }
    // Invalid token - continue to check for existing session
  }

  // Check for existing session cookie
  const sessionCookie = getCookie(request, 'civitas_session');
  if (sessionCookie) {
    // Validate session with gate worker
    const validateResponse = await fetch(`${GATE_URL}/validate-token?token=${sessionCookie}`);
    const validateData = await validateResponse.json() as { valid: boolean };

    if (validateData.valid) {
      // Session is valid - allow request through
      return context.next();
    }
    // Invalid session - clear cookie and redirect to gate
  }

  // No valid session - redirect to gate
  const gateUrl = `${GATE_URL}?site=presentation`;
  return new Response(null, {
    status: 302,
    headers: {
      'Location': gateUrl,
      // Clear any invalid session cookie
      'Set-Cookie': 'civitas_session=; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=0',
    },
  });
};

function getCookie(request: Request, name: string): string | null {
  const cookieHeader = request.headers.get('Cookie');
  if (!cookieHeader) return null;

  const cookies = cookieHeader.split(';').map(c => c.trim());
  for (const cookie of cookies) {
    const [key, value] = cookie.split('=');
    if (key === name) return value;
  }
  return null;
}
