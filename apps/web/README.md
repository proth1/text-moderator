# Civitas AI Text Moderator - Web Frontend

React + TypeScript frontend for the Civitas AI Text Moderator platform.

## Tech Stack

- **React 18** - UI library
- **TypeScript** - Type safety
- **Vite** - Build tooling and dev server
- **TailwindCSS v4** - Styling
- **React Router v6** - Client-side routing
- **TanStack Query** - Server state management
- **Zustand** - Client state management
- **Axios** - HTTP client
- **Lucide React** - Icons

## Project Structure

```
src/
├── components/
│   ├── ui/              # Reusable UI components (Button, Card, Badge, etc.)
│   └── layout/          # Layout components (MainLayout, Sidebar)
├── pages/               # Route pages
│   ├── Dashboard.tsx
│   ├── ModerationDemo.tsx
│   ├── ModeratorQueue.tsx
│   ├── ReviewDetail.tsx
│   ├── PolicyList.tsx
│   ├── PolicyEditor.tsx
│   ├── AuditLog.tsx
│   ├── Analytics.tsx
│   └── Login.tsx
├── hooks/               # Custom React hooks
│   ├── useModeration.ts
│   ├── usePolicies.ts
│   ├── useReviews.ts
│   └── useAuth.ts
├── api/                 # API client and endpoints
│   ├── client.ts
│   ├── moderation.ts
│   ├── policies.ts
│   ├── reviews.ts
│   └── evidence.ts
├── store/              # Zustand stores
│   └── authStore.ts
├── types/              # TypeScript type definitions
│   └── index.ts
├── lib/                # Utilities
│   └── utils.ts
├── App.tsx             # Main app component with routing
└── main.tsx            # App entry point
```

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn

### Installation

```bash
# Install dependencies
npm install
```

### Development

```bash
# Start dev server (default: http://localhost:5173)
npm run dev
```

### Build

```bash
# Build for production
npm run build

# Preview production build
npm run preview
```

## Environment Variables

Create a `.env` file in the root directory:

```env
VITE_API_URL=http://localhost:8080/api/v1
```

## Features

### 1. Dashboard
- Overview statistics
- Total moderations, blocked percentage
- Pending reviews count
- Active policies count

### 2. Moderation Demo
- Real-time text moderation
- Category score visualization
- Confidence indicators
- Action recommendations (Allow/Warn/Block/Review)

### 3. Review Queue
- List of flagged content requiring human review
- Filter by status (pending/reviewed)
- Sort by confidence, category, date
- Direct link to detailed review

### 4. Review Detail
- Full content view
- Model analysis (category scores, reasons)
- Policy information
- Action buttons (Approve/Override/Escalate)
- Action history

### 5. Policy Management
- Create and edit policies
- Configure category thresholds (sliders)
- Set actions per category
- Draft/Publish workflow
- Version tracking

### 6. Audit Log
- Complete history of moderation decisions
- Searchable and filterable
- Export functionality
- Compliance reporting

### 7. Analytics
- Moderation volume trends
- Model performance metrics
- False positive analysis
- Category distribution

## API Integration

The frontend communicates with the backend API at `/api/v1`. The Vite dev server proxies requests to `http://localhost:8080`.

### Authentication

API requests include an `X-API-Key` header. The key is stored in localStorage after login.

### Error Handling

- 401 responses trigger automatic logout
- Network errors display user-friendly messages
- Form validation prevents invalid submissions

## Code Style

- ESLint and TypeScript strict mode enabled
- Functional components with hooks
- Custom hooks for API calls
- Zustand for minimal client state
- TanStack Query for server state

## Development Notes

- Hot module replacement (HMR) enabled
- TypeScript strict mode
- All API calls typed with TypeScript interfaces
- Responsive design (mobile-friendly)
- Accessible UI components

## Next Steps

- [ ] Implement real authentication flow
- [ ] Add data visualization charts (Analytics page)
- [ ] Implement real-time updates (WebSocket)
- [ ] Add bulk actions in Review Queue
- [ ] Export audit logs (CSV/JSON)
- [ ] Add user management UI
- [ ] Implement policy comparison/diff view
- [ ] Add keyboard shortcuts
- [ ] Dark mode support
