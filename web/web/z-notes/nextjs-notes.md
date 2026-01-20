# Next.js Development Notes

> Quick reference for developers coming from Vite/React SPA background

---

## 🎯 TL;DR: Key Mindset Shifts

| Vite (React SPA) | Next.js (App Router) |
|------------------|----------------------|
| Everything runs in the browser | Components run on server by default |
| `useEffect` + `fetch` for data | Async components fetch data directly |
| One entry point, client routing | File-based routing with server rendering |
| All code ships to browser | Server code never reaches client |
| State lives in browser | State can be on server OR client |

---

## 1. Server Components vs Client Components

### The Big Difference

**Vite:** Everything is a "client component" - all React code runs in the browser.

**Next.js:** Components are **Server Components by default**. They run on the server, never ship JavaScript to the browser, and can directly access databases, file systems, and APIs.

### When to Use Each

| Use Server Component (default) | Use Client Component (`'use client'`) |
|-------------------------------|---------------------------------------|
| Fetching data | `useState`, `useEffect`, `useRef` |
| Accessing backend resources | Event handlers (`onClick`, `onChange`) |
| Rendering static content | Browser APIs (localStorage, etc.) |
| Keeping secrets on server | Interactivity and real-time updates |
| Large dependencies (keep off client) | Third-party client libraries |

### Syntax

```tsx
// Server Component (DEFAULT - no directive needed)
export default async function Page() {
  const data = await fetch('https://api.example.com/data')
  return <div>{data}</div>
}
```

```tsx
// Client Component (MUST add directive)
'use client'

import { useState } from 'react'

export default function Counter() {
  const [count, setCount] = useState(0)
  return <button onClick={() => setCount(count + 1)}>{count}</button>
}
```

### ⚠️ Common Gotcha

You cannot import a Server Component INTO a Client Component:

```tsx
// ❌ WRONG - This won't work
'use client'
import ServerComponent from './ServerComponent'

export default function ClientComponent() {
  return <ServerComponent /> // Error!
}
```

```tsx
// ✅ CORRECT - Pass as children from a Server Component
// page.tsx (Server Component)
import ClientComponent from './ClientComponent'
import ServerComponent from './ServerComponent'

export default function Page() {
  return (
    <ClientComponent>
      <ServerComponent />
    </ClientComponent>
  )
}
```

---

## 2. Data Fetching

### Vite Approach (Client-Side)

```tsx
// Vite - useEffect + fetch
import { useState, useEffect } from 'react'

export function Posts() {
  const [posts, setPosts] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/posts')
      .then(res => res.json())
      .then(data => {
        setPosts(data)
        setLoading(false)
      })
  }, [])

  if (loading) return <div>Loading...</div>
  return <ul>{posts.map(p => <li key={p.id}>{p.title}</li>)}</ul>
}
```

### Next.js Approach (Server-Side)

```tsx
// Next.js - Async Server Component (MUCH simpler!)
export default async function Posts() {
  const res = await fetch('https://api.example.com/posts')
  const posts = await res.json()

  return <ul>{posts.map(p => <li key={p.id}>{p.title}</li>)}</ul>
}
```

**Benefits:**
- No loading states to manage (handled by Suspense)
- No client-side JavaScript for fetching
- Data fetched at build time or request time
- Secure - API keys/secrets never reach client

### Caching Options

```tsx
// Static (cached until manually invalidated) - like getStaticProps
const data = await fetch('https://...', { cache: 'force-cache' })

// Dynamic (no cache, fresh every request) - like getServerSideProps
const data = await fetch('https://...', { cache: 'no-store' })

// Revalidate (cache for X seconds, then refresh)
const data = await fetch('https://...', { next: { revalidate: 60 } })
```

### Client-Side Fetching (When Needed)

For **interactive** data that changes frequently, use TanStack Query or SWR:

```tsx
'use client'

import { useQuery } from '@tanstack/react-query'

export function LiveData() {
  const { data, isLoading } = useQuery({
    queryKey: ['live-data'],
    queryFn: () => fetch('/api/live').then(r => r.json()),
    refetchInterval: 5000, // Poll every 5 seconds
  })

  if (isLoading) return <div>Loading...</div>
  return <div>{data.value}</div>
}
```

---

## 3. File-Based Routing

### Vite (Manual)

```tsx
// App.tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/projects/:id" element={<Project />} />
      </Routes>
    </BrowserRouter>
  )
}
```

### Next.js (Automatic)

```
app/
├── page.tsx              → /
├── dashboard/
│   └── page.tsx          → /dashboard
├── projects/
│   ├── page.tsx          → /projects
│   └── [id]/
│       └── page.tsx      → /projects/123, /projects/abc
```

**No router configuration needed!**

### Special Files in App Router

| File | Purpose |
|------|---------|
| `page.tsx` | The UI for a route (required for route to be accessible) |
| `layout.tsx` | Shared UI that wraps child pages (persists across navigation) |
| `loading.tsx` | Loading UI (automatic Suspense fallback) |
| `error.tsx` | Error UI (automatic error boundary) |
| `not-found.tsx` | 404 UI |
| `route.ts` | API endpoint (replaces pages/api) |

### Dynamic Routes

```tsx
// app/projects/[id]/page.tsx
export default async function ProjectPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const project = await getProject(id)
  
  return <div>{project.name}</div>
}
```

---

## 4. Layouts (Persistent UI)

### Vite Approach

```tsx
// Manual layout wrapper
function Layout({ children }) {
  return (
    <div>
      <Sidebar />
      <main>{children}</main>
    </div>
  )
}

// Must wrap each page manually
function Dashboard() {
  return (
    <Layout>
      <h1>Dashboard</h1>
    </Layout>
  )
}
```

### Next.js Approach

```tsx
// app/dashboard/layout.tsx - Automatic!
export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div>
      <Sidebar />
      <main>{children}</main>
    </div>
  )
}

// app/dashboard/page.tsx - No wrapper needed
export default function Dashboard() {
  return <h1>Dashboard</h1>
}
```

**Key difference:** Layouts preserve state during navigation. The sidebar won't re-render when navigating between dashboard pages.

---

## 5. Navigation

### Vite (react-router-dom)

```tsx
import { Link, useNavigate } from 'react-router-dom'

function Nav() {
  const navigate = useNavigate()
  
  return (
    <>
      <Link to="/dashboard">Dashboard</Link>
      <button onClick={() => navigate('/projects')}>Go to Projects</button>
    </>
  )
}
```

### Next.js

```tsx
import Link from 'next/link'
import { useRouter } from 'next/navigation'

function Nav() {
  const router = useRouter()
  
  return (
    <>
      <Link href="/dashboard">Dashboard</Link>
      <button onClick={() => router.push('/projects')}>Go to Projects</button>
    </>
  )
}
```

### Router Methods

```tsx
'use client'

import { useRouter } from 'next/navigation'

export function Navigation() {
  const router = useRouter()
  
  // Navigate to a route
  router.push('/dashboard')
  
  // Replace current history entry
  router.replace('/dashboard')
  
  // Go back
  router.back()
  
  // Refresh current route (re-fetch data)
  router.refresh()
  
  // Prefetch a route
  router.prefetch('/projects')
}
```

### Getting Current Route Info

```tsx
'use client'

import { usePathname, useSearchParams } from 'next/navigation'

export function CurrentRoute() {
  const pathname = usePathname()        // "/dashboard/projects"
  const searchParams = useSearchParams() // URLSearchParams object
  
  const filter = searchParams.get('filter') // "active"
  
  return <div>Current path: {pathname}</div>
}
```

---

## 6. API Routes

### Vite (Requires Separate Backend)

You need Express, Fastify, or another server.

### Next.js (Built-in)

```tsx
// app/api/projects/route.ts
import { NextResponse } from 'next/server'

export async function GET() {
  const projects = await db.projects.findMany()
  return NextResponse.json(projects)
}

export async function POST(request: Request) {
  const body = await request.json()
  const project = await db.projects.create({ data: body })
  return NextResponse.json(project, { status: 201 })
}
```

**But wait!** With Server Components, you often don't need API routes:

```tsx
// Instead of: fetch('/api/projects')
// Just do this in a Server Component:

import { db } from '@/lib/db'

export default async function ProjectsPage() {
  const projects = await db.projects.findMany() // Direct DB access!
  return <ProjectList projects={projects} />
}
```

---

## 7. Server Actions (Form Handling)

### Vite Approach

```tsx
function CreateProject() {
  const handleSubmit = async (e) => {
    e.preventDefault()
    const formData = new FormData(e.target)
    await fetch('/api/projects', {
      method: 'POST',
      body: JSON.stringify(Object.fromEntries(formData)),
    })
  }

  return <form onSubmit={handleSubmit}>...</form>
}
```

### Next.js Server Actions

```tsx
// app/projects/actions.ts
'use server'

import { revalidatePath } from 'next/cache'

export async function createProject(formData: FormData) {
  const name = formData.get('name')
  await db.projects.create({ data: { name } })
  revalidatePath('/projects') // Refresh the projects page
}
```

```tsx
// app/projects/create-form.tsx
import { createProject } from './actions'

export function CreateProjectForm() {
  return (
    <form action={createProject}>
      <input name="name" />
      <button type="submit">Create</button>
    </form>
  )
}
```

**No `useState`, no `fetch`, no loading states to manage manually!**

---

## 8. Environment Variables

### Vite

```bash
# .env
VITE_API_URL=https://api.example.com  # Prefix with VITE_ to expose
```

```tsx
const apiUrl = import.meta.env.VITE_API_URL
```

### Next.js

```bash
# .env.local
# Server-only (secure - never sent to browser)
DATABASE_URL=postgres://...
API_SECRET=secret123

# Client-exposed (prefix with NEXT_PUBLIC_)
NEXT_PUBLIC_API_URL=https://api.example.com
```

```tsx
// Server Component - can access all env vars
const dbUrl = process.env.DATABASE_URL

// Client Component - only NEXT_PUBLIC_ vars
const apiUrl = process.env.NEXT_PUBLIC_API_URL
```

---

## 9. Loading & Error States

### Vite (Manual)

```tsx
function Page() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [data, setData] = useState(null)
  
  // ... lots of boilerplate
}
```

### Next.js (Automatic with Special Files)

```
app/dashboard/
├── page.tsx      # Your page
├── loading.tsx   # Shown while page loads
├── error.tsx     # Shown on error
└── not-found.tsx # Shown for 404
```

```tsx
// loading.tsx
export default function Loading() {
  return <div>Loading dashboard...</div>
}
```

```tsx
// error.tsx
'use client' // Error components must be Client Components

export default function Error({
  error,
  reset,
}: {
  error: Error
  reset: () => void
}) {
  return (
    <div>
      <h2>Something went wrong!</h2>
      <button onClick={() => reset()}>Try again</button>
    </div>
  )
}
```

---

## 10. Streaming & Suspense

### Partial Page Loading

```tsx
import { Suspense } from 'react'

export default function Dashboard() {
  return (
    <div>
      <h1>Dashboard</h1>
      
      {/* This loads immediately */}
      <UserInfo />
      
      {/* This streams in when ready */}
      <Suspense fallback={<div>Loading stats...</div>}>
        <SlowStats />
      </Suspense>
      
      {/* This also streams independently */}
      <Suspense fallback={<div>Loading activity...</div>}>
        <RecentActivity />
      </Suspense>
    </div>
  )
}
```

Users see the page immediately with loading indicators, then content streams in as it becomes available.

---

## 11. Common Patterns

### Passing Data from Server to Client

```tsx
// page.tsx (Server Component)
import ClientComponent from './client-component'

export default async function Page() {
  const data = await fetchData() // Runs on server
  
  // Pass serializable data as props
  return <ClientComponent initialData={data} />
}
```

```tsx
// client-component.tsx (Client Component)
'use client'

export default function ClientComponent({ initialData }) {
  const [data, setData] = useState(initialData)
  // Now you can use hooks, events, etc.
}
```

### Composing Server and Client Components

```tsx
// page.tsx (Server)
import Sidebar from './sidebar'        // Server Component
import InteractiveList from './list'   // Client Component

export default async function Page() {
  const items = await getItems()
  
  return (
    <div className="flex">
      <Sidebar />
      <InteractiveList items={items} />
    </div>
  )
}
```

---

## 12. Quick Reference

### Hooks Location

| Hook | Server Component | Client Component |
|------|-----------------|------------------|
| `useState` | ❌ | ✅ |
| `useEffect` | ❌ | ✅ |
| `useRef` | ❌ | ✅ |
| `useRouter` | ❌ | ✅ |
| `usePathname` | ❌ | ✅ |
| `useSearchParams` | ❌ | ✅ |
| `async/await` | ✅ | ❌ (use hooks) |

### Import Differences

```tsx
// Vite / Pages Router
import { useRouter } from 'next/router'

// App Router
import { useRouter } from 'next/navigation'
import { usePathname, useSearchParams } from 'next/navigation'
```

---

## 13. Development Commands

```bash
# Start dev server
npm run dev

# Build for production
npm run build

# Start production server
npm start

# Lint
npm run lint
```

---

## 14. Project Structure for RCNbuild

```
src/
├── app/                    # Routes and pages
│   ├── layout.tsx          # Root layout
│   ├── page.tsx            # Landing page
│   └── dashboard/          # Dashboard routes
├── components/             # React components
│   ├── ui/                 # Generic UI (buttons, cards)
│   └── features/           # Feature-specific components
├── lib/                    # Utilities and logic
│   ├── api.ts              # API client
│   ├── providers/          # React context providers
│   └── hooks/              # Custom hooks
└── types/                  # TypeScript types
```

---

## Resources

- [Next.js Docs](https://nextjs.org/docs)
- [App Router Introduction](https://nextjs.org/docs/app)
- [Server Components](https://nextjs.org/docs/app/building-your-application/rendering/server-components)
- [Data Fetching](https://nextjs.org/docs/app/building-your-application/data-fetching)
