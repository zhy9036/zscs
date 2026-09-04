export interface ProjectFile {
  id: string
  filename: string
  content_type: string
  size: number
}

export interface Project {
  id: string
  title: string
  created_at: string
  updated_at: string
  files?: ProjectFile[]
}
