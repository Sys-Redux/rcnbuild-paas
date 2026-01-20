export interface Project {
  id: string
  name: string
  slug: string
  repo_full_name: string
  repo_url: string
  branch: string
  root_directory: string
  build_command: string
  start_command: string
  runtime: 'nodejs' | 'python' | 'go' | 'static' | 'docker' | 'unknown'
  port: number
  created_at: string
  updated_at: string
  live_deployment?: Deployment
}

export interface Deployment {
  id: string
  project_id: string
  commit_sha: string
  commit_message?: string
  commit_author?: string
  branch?: string
  status: 'pending' | 'building' | 'deploying' | 'live' | 'failed' | 'cancelled' | 'superseded'
  image_tag?: string
  url?: string
  error_message?: string
  created_at: string
  started_at?: string
  completed_at?: string
}

export interface EnvVar {
  key: string
  value: string // Masked in API responses
  created_at: string
}

export interface GitHubRepo {
  id: number
  name: string
  full_name: string
  private: boolean
  html_url: string
  default_branch: string
  pushed_at: string
}