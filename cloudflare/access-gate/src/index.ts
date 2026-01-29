/**
 * Civitas Access Gate Worker
 *
 * Complete authentication flow:
 * 1. User visits gate → enters email + password
 * 2. Password validated → magic link generated and emailed
 * 3. User clicks link → session created, redirected to target site
 * 4. All access logged to KV
 */

interface Env {
  ACCESS_LOG: KVNamespace;
  MAGIC_TOKENS: KVNamespace;
  SESSIONS: KVNamespace;
  ACCESS_PASSWORD: string;
  RESEND_API_KEY?: string;
  WORKER_URL: string;
}

const SESSION_DURATION = 24 * 60 * 60; // 24 hours in seconds
const MAGIC_LINK_EXPIRY = 10 * 60; // 10 minutes in seconds

// Target sites configuration
const SITES = {
  web: {
    name: 'Civitas Web App',
    url: 'https://civitas-web.pages.dev',
  },
  presentation: {
    name: 'Civitas Presentation',
    url: 'https://civitas-presentation.pages.dev',
  },
};

// HTML Templates
const loginPageHTML = (error?: string, targetSite = 'web') => `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Civitas AI - Request Access</title>
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="min-h-screen bg-gradient-to-br from-slate-900 via-blue-900 to-slate-900 flex items-center justify-center p-4">
  <div class="bg-white/10 backdrop-blur-lg rounded-2xl shadow-2xl p-8 w-full max-w-md border border-white/20">
    <div class="text-center mb-8">
      <div class="w-16 h-16 bg-blue-500 rounded-xl flex items-center justify-center mx-auto mb-4">
        <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
        </svg>
      </div>
      <h1 class="text-2xl font-bold text-white mb-1">Civitas AI</h1>
      <p class="text-blue-200 text-sm">Request access to ${SITES[targetSite as keyof typeof SITES]?.name || 'the platform'}</p>
    </div>

    ${error ? `
    <div class="bg-red-500/20 border border-red-500/50 text-red-200 px-4 py-3 rounded-lg mb-6 text-sm">
      ${error}
    </div>
    ` : ''}

    <form method="POST" action="/request-access" class="space-y-5">
      <input type="hidden" name="target" value="${targetSite}" />

      <div>
        <label class="block text-sm font-medium text-blue-100 mb-2">Email Address</label>
        <input
          type="email"
          name="email"
          required
          autocomplete="email"
          placeholder="you@company.com"
          class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-blue-300/50 focus:ring-2 focus:ring-blue-400 focus:border-transparent transition"
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-blue-100 mb-2">Access Code</label>
        <input
          type="password"
          name="password"
          required
          autocomplete="current-password"
          placeholder="Enter your access code"
          class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-blue-300/50 focus:ring-2 focus:ring-blue-400 focus:border-transparent transition"
        />
      </div>

      <button
        type="submit"
        class="w-full bg-blue-500 hover:bg-blue-600 text-white py-3 px-4 rounded-lg font-semibold transition duration-200 transform hover:scale-[1.02]"
      >
        Send Magic Link
      </button>
    </form>

    <p class="mt-6 text-center text-xs text-blue-300/60">
      A secure link will be sent to your email
    </p>
  </div>
</body>
</html>
`;

const magicLinkSentHTML = (email: string, magicLink?: string) => `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Civitas AI - Check Your Email</title>
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="min-h-screen bg-gradient-to-br from-slate-900 via-blue-900 to-slate-900 flex items-center justify-center p-4">
  <div class="bg-white/10 backdrop-blur-lg rounded-2xl shadow-2xl p-8 w-full max-w-md border border-white/20 text-center">
    <div class="w-16 h-16 bg-green-500 rounded-full flex items-center justify-center mx-auto mb-6">
      <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
      </svg>
    </div>
    <h1 class="text-2xl font-bold text-white mb-2">Check Your Email</h1>
    <p class="text-blue-200 mb-4">
      We've sent a magic link to<br/>
      <strong class="text-white">${email}</strong>
    </p>
    <p class="text-sm text-blue-300/70 mb-6">
      Click the link in the email to access Civitas AI.<br/>
      The link expires in 10 minutes.
    </p>
    ${magicLink ? `
    <div class="mt-6 p-4 bg-yellow-500/20 border border-yellow-500/50 rounded-lg">
      <p class="text-yellow-200 text-xs mb-2">Development Mode - Magic Link:</p>
      <a href="${magicLink}" class="text-yellow-300 text-sm break-all hover:underline">${magicLink}</a>
    </div>
    ` : ''}
  </div>
</body>
</html>
`;

const errorPageHTML = (message: string) => `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Civitas AI - Error</title>
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="min-h-screen bg-gradient-to-br from-slate-900 via-red-900 to-slate-900 flex items-center justify-center p-4">
  <div class="bg-white/10 backdrop-blur-lg rounded-2xl shadow-2xl p-8 w-full max-w-md border border-white/20 text-center">
    <div class="w-16 h-16 bg-red-500 rounded-full flex items-center justify-center mx-auto mb-6">
      <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
      </svg>
    </div>
    <h1 class="text-2xl font-bold text-white mb-2">Access Denied</h1>
    <p class="text-red-200 mb-6">${message}</p>
    <a href="/" class="inline-block bg-white/20 hover:bg-white/30 text-white py-2 px-6 rounded-lg transition">
      Try Again
    </a>
  </div>
</body>
</html>
`;

// Generate secure random token
function generateToken(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
}

// Send magic link email via Resend
async function sendMagicLinkEmail(email: string, magicLink: string, siteName: string, apiKey: string): Promise<boolean> {
  try {
    const response = await fetch('https://api.resend.com/emails', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        from: 'Civitas AI <onboarding@resend.dev>',
        to: [email],
        subject: `Your Civitas AI Access Link`,
        headers: {
          'X-Entity-Ref-ID': crypto.randomUUID(), // Prevent threading
        },
        text: `Civitas AI - Access Link

Click or copy this link to access ${siteName}:

${magicLink}

This link expires in 10 minutes.

If you didn't request this link, you can safely ignore this email.`,
        html: `
          <div style="font-family: system-ui, sans-serif; max-width: 500px; margin: 0 auto; padding: 20px;">
            <h1 style="color: #3b82f6; margin-bottom: 20px;">Civitas AI</h1>
            <p style="color: #374151; font-size: 16px; line-height: 1.6;">
              Click the button below to access <strong>${siteName}</strong>:
            </p>
            <a href="${magicLink}" style="display: inline-block; background: #3b82f6; color: white; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-weight: 600; margin: 20px 0;">
              Access Civitas AI
            </a>
            <p style="color: #6b7280; font-size: 14px;">
              This link expires in 10 minutes.
            </p>
            <p style="color: #9ca3af; font-size: 12px; margin-top: 20px; word-break: break-all;">
              Or copy this link: ${magicLink}
            </p>
            <p style="color: #9ca3af; font-size: 12px; margin-top: 10px;">
              If you didn't request this link, you can safely ignore this email.
            </p>
          </div>
        `,
      }),
    });
    return response.ok;
  } catch (e) {
    console.error('Failed to send email:', e);
    return false;
  }
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const sessionCookie = getCookie(request, 'civitas_session');

    // Handle different routes
    switch (url.pathname) {
      case '/':
      case '/login': {
        const target = url.searchParams.get('site') || 'web';
        return new Response(loginPageHTML(undefined, target), {
          headers: { 'Content-Type': 'text/html' },
        });
      }

      case '/request-access': {
        if (request.method !== 'POST') {
          return new Response('Method not allowed', { status: 405 });
        }

        const formData = await request.formData();
        const email = (formData.get('email') as string)?.toLowerCase().trim();
        const password = formData.get('password') as string;
        const target = (formData.get('target') as string) || 'web';

        // Validate password
        if (password !== env.ACCESS_PASSWORD) {
          return new Response(loginPageHTML('Invalid access code.', target), {
            status: 401,
            headers: { 'Content-Type': 'text/html' },
          });
        }

        // Validate email
        if (!email || !email.includes('@')) {
          return new Response(loginPageHTML('Please enter a valid email address.', target), {
            status: 400,
            headers: { 'Content-Type': 'text/html' },
          });
        }

        // Generate magic link token
        const token = generateToken();
        const tokenData = {
          email,
          target,
          createdAt: Date.now(),
          ip: request.headers.get('CF-Connecting-IP') || 'unknown',
        };

        // Store token in KV
        await env.MAGIC_TOKENS.put(token, JSON.stringify(tokenData), {
          expirationTtl: MAGIC_LINK_EXPIRY,
        });

        // Log the access request
        const logEntry = {
          email,
          action: 'magic_link_requested',
          target,
          timestamp: new Date().toISOString(),
          ip: request.headers.get('CF-Connecting-IP') || 'unknown',
          userAgent: request.headers.get('User-Agent') || 'unknown',
          country: request.headers.get('CF-IPCountry') || 'unknown',
        };
        await env.ACCESS_LOG.put(`${Date.now()}-${email.replace(/[^a-zA-Z0-9]/g, '_')}`, JSON.stringify(logEntry), {
          expirationTtl: 90 * 24 * 60 * 60,
        });

        const magicLink = `${env.WORKER_URL}/verify?token=${token}`;
        const siteName = SITES[target as keyof typeof SITES]?.name || 'Civitas AI';

        // Try to send email
        let emailSent = false;
        if (env.RESEND_API_KEY) {
          emailSent = await sendMagicLinkEmail(email, magicLink, siteName, env.RESEND_API_KEY);
        }

        // Show success page (include link if email not configured)
        return new Response(magicLinkSentHTML(email, emailSent ? undefined : magicLink), {
          headers: { 'Content-Type': 'text/html' },
        });
      }

      case '/verify': {
        const token = url.searchParams.get('token');
        if (!token) {
          return new Response(errorPageHTML('Invalid or missing token.'), {
            status: 400,
            headers: { 'Content-Type': 'text/html' },
          });
        }

        // Get token data
        const tokenDataStr = await env.MAGIC_TOKENS.get(token);
        if (!tokenDataStr) {
          return new Response(errorPageHTML('This link has expired or is invalid. Please request a new one.'), {
            status: 401,
            headers: { 'Content-Type': 'text/html' },
          });
        }

        const tokenData = JSON.parse(tokenDataStr);

        // Delete the token (one-time use)
        await env.MAGIC_TOKENS.delete(token);

        // Create session
        const sessionId = generateToken();
        const sessionData = {
          email: tokenData.email,
          target: tokenData.target,
          createdAt: Date.now(),
          ip: request.headers.get('CF-Connecting-IP') || 'unknown',
        };
        await env.SESSIONS.put(sessionId, JSON.stringify(sessionData), {
          expirationTtl: SESSION_DURATION,
        });

        // Log successful authentication
        const logEntry = {
          email: tokenData.email,
          action: 'authenticated',
          target: tokenData.target,
          timestamp: new Date().toISOString(),
          ip: request.headers.get('CF-Connecting-IP') || 'unknown',
          country: request.headers.get('CF-IPCountry') || 'unknown',
        };
        await env.ACCESS_LOG.put(`${Date.now()}-auth-${tokenData.email.replace(/[^a-zA-Z0-9]/g, '_')}`, JSON.stringify(logEntry), {
          expirationTtl: 90 * 24 * 60 * 60,
        });

        // Redirect to target site with auth token in URL
        // The target site's middleware will validate this token and set its own session cookie
        const targetUrl = SITES[tokenData.target as keyof typeof SITES]?.url || SITES.web.url;
        return new Response(null, {
          status: 302,
          headers: {
            'Location': `${targetUrl}?civitas_auth=${sessionId}`,
          },
        });
      }

      case '/logout': {
        if (sessionCookie) {
          await env.SESSIONS.delete(sessionCookie);
        }
        return new Response(null, {
          status: 302,
          headers: {
            'Location': '/',
            'Set-Cookie': 'civitas_session=; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=0',
          },
        });
      }

      case '/logs': {
        // Admin endpoint - requires auth header
        const authHeader = request.headers.get('Authorization');
        if (authHeader !== `Bearer ${env.ACCESS_PASSWORD}`) {
          return new Response('Unauthorized', { status: 401 });
        }

        const logs = await env.ACCESS_LOG.list();
        const entries = await Promise.all(
          logs.keys.slice(0, 100).map(async (key) => {
            const value = await env.ACCESS_LOG.get(key.name);
            return { id: key.name, ...JSON.parse(value || '{}') };
          })
        );

        return new Response(JSON.stringify(entries, null, 2), {
          headers: { 'Content-Type': 'application/json' },
        });
      }

      case '/session': {
        // Check session validity
        if (!sessionCookie) {
          return new Response(JSON.stringify({ valid: false }), {
            headers: { 'Content-Type': 'application/json' },
          });
        }

        const sessionData = await env.SESSIONS.get(sessionCookie);
        if (!sessionData) {
          return new Response(JSON.stringify({ valid: false }), {
            headers: { 'Content-Type': 'application/json' },
          });
        }

        return new Response(JSON.stringify({ valid: true, ...JSON.parse(sessionData) }), {
          headers: { 'Content-Type': 'application/json' },
        });
      }

      case '/validate-token': {
        // Validate auth token from URL parameter (called by Pages Functions)
        const token = url.searchParams.get('token');
        if (!token) {
          return new Response(JSON.stringify({ valid: false, error: 'Missing token' }), {
            status: 400,
            headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' },
          });
        }

        const sessionData = await env.SESSIONS.get(token);
        if (!sessionData) {
          return new Response(JSON.stringify({ valid: false, error: 'Invalid or expired token' }), {
            status: 401,
            headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' },
          });
        }

        return new Response(JSON.stringify({ valid: true, ...JSON.parse(sessionData) }), {
          headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' },
        });
      }

      default:
        return new Response('Not Found', { status: 404 });
    }
  },
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
