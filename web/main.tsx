import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
//import App from './App.tsx'
//import City from './City.tsx'
import Dummy from './Dummy.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Dummy />
  </StrictMode>
)
