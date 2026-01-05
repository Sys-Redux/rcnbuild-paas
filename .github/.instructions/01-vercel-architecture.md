<!-- markdownlint-disable -->
# Vercel Architecture & Features Research

## Overview

Vercel is a frontend-focused platform-as-a-service (PaaS) that specializes in deploying modern web applications with serverless architecture, global edge delivery, and automatic scaling.

## Core Infrastructure Components

### 1. Build System

**Build Container Architecture:**
- Auto-scaling fleet of EC2 instances powered by AWS Fargate
- Supports 35+ frontend frameworks with automatic detection
- Build containers process files and ping status API endpoints
- Build output adheres to the Build Output API specification

**Build Process Flow:**
1. Files uploaded via POST request to Amazon S3 (static storage)
2. Deployment POST request triggers build scheduling
3. Authentication & configuration validation
4. Build container executes framework-specific builder
5. Provisions resources (Serverless Functions, Edge Functions, Static Assets)
6. Deployment metadata uploaded to static storage
7. Deployment ready for CDN serving

**Concurrency:**
- Hobby: 1 concurrent build
- Pro: Up to 12 concurrent builds
- Enterprise: Custom build slot allocation

### 2. Request Handling (Request Phase)

**DNS & Routing:**
- Uses anycast IP addresses (not GeoDNS)
- Routes via Amazon Global Accelerator
- Optimal data center selection based on:
  - Number of hops
  - Round-trip time
  - Available bandwidth

**Gateway Layer:**
- Kubernetes cluster entry point (Amazon EKS)
- Request inspection and malicious user filtering
- Virtual machine reverse proxy for rewriting/proxying
- Determines deployment version based on hostname
- Fetches deployment metadata for routing decisions

**Response Types:**
| Resource Type | Handling |
|--------------|----------|
| Static Resources | Downloaded from S3 static storage |
| Serverless Functions | AWS Lambda execution in deployed region |
| Edge Functions | Edge function execution at edge location |
| ISR (Incremental Static Regeneration) | Check cache → serve stale → regenerate in background |
| Optimized Images | Dedicated optimization service with edge caching |

### 3. Serverless Infrastructure

**Serverless Functions:**
- Built on AWS Lambda
- Support for API routes and server-side rendered pages
- Automatic scaling per-request

**Edge Functions:**
- Execute at edge locations (closest to users)
- Used for Middleware and edge runtime functions
- Near-instant cold start times

### 4. CDN & Edge Network

- Global edge network powered by AWS Global Network
- Caching based on Cache-Control headers
- Automatic DDoS protection via Amazon Global Accelerator
- Anycast routing for optimal performance

## Deployment Methods

### 1. Git Integration (Primary - "One-Click")
- Automatic deployment on commit/push
- Supports GitHub, GitLab, Bitbucket, Azure DevOps
- Preview deployments for every pull request
- Production deployment from main branch

### 2. Vercel CLI
```bash
npm i -g vercel
vercel --prod  # Deploy to production
```
- Links local directory to Vercel project
- Creates `.vercel` directory for project/org IDs

### 3. Deploy Hooks
- Unique URL per project
- HTTP GET/POST request triggers deployment
- Requires connected Git repository

### 4. REST API
- Generate SHA for each file
- Upload files to Vercel
- POST request to create deployment with file references

## Framework-Defined Infrastructure (FdI)

**Concept:** The deployment environment automatically provisions infrastructure derived from the framework and application code—eliminating manual infrastructure configuration.

**How It Works:**
1. Build-time parser analyzes framework source code
2. Understands developer intent from code patterns
3. Automatically generates IaC configuration
4. Deploys appropriate infrastructure primitives

**Example: Next.js Page Rendering**

```tsx
// getServerSideProps → Provisions Serverless Function
export async function getServerSideProps() {
  const posts = await getBlogPosts();
  return { props: { posts } }
}

// getStaticProps → Static HTML at build time (no compute needed)
export async function getStaticProps() {
  const posts = await getBlogPosts();
  return { props: { posts } }
}
```

**Automatic Infrastructure Mapping:**
| Framework Feature | Infrastructure Provisioned |
|------------------|---------------------------|
| `getServerSideProps` | Serverless Function (AWS Lambda) |
| `getStaticProps` | Static file storage |
| Middleware | Edge compute resources |
| Image optimization | Image optimization service |
| Form actions (Remix/SvelteKit) | Serverless Function |
| Deferred Static Generation (Gatsby) | Serverless + S3 + Edge cache |

## Environments

1. **Local Development** - Testing on local machine
2. **Preview** - Unique URL per commit/PR for QA/collaboration
3. **Production** - User-facing deployment with production domain

## Immutable Deployments

- Each `git commit` creates completely new virtual infrastructure
- Deployments are never changed after creation
- Unused deployments scale to zero (no compute cost)
- Enables instant rollbacks
- Production infrastructure maps perfectly to code version

## Key AWS Services Used

| Service | Purpose |
|---------|---------|
| Amazon S3 | Static file storage |
| Amazon SQS | Deployment scheduling queue |
| AWS Fargate/EC2 | Build containers |
| Amazon Global Accelerator | Anycast routing, DDoS protection |
| AWS Global Network | Edge network infrastructure |
| Amazon EKS | Kubernetes cluster (gateway) |
| AWS Lambda | Serverless function execution |

## Developer Experience Features

- Automatic framework detection (35+ frameworks)
- Git-triggered deployments
- Preview deployments with unique URLs
- Comments on Preview Deployments for team collaboration
- Build Output API for custom builders
- Instant rollbacks to any previous deployment
