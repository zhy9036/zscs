import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { Providers } from './providers'
import { RequireAuth } from './routes'
import { LoginPage } from '../pages/LoginPage'
import { HomePage } from '../pages/HomePage'
import { NewProjectPage } from '../pages/NewProjectPage'
import { ProjectPage } from '../pages/ProjectPage'

export function App() {
  return (
    <Providers>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<RequireAuth />}>
            <Route path="/" element={<HomePage />} />
            <Route path="/new" element={<NewProjectPage />} />
            <Route path="/chat/:projectId" element={<ProjectPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </Providers>
  )
}
