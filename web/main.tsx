import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
//import App from './App.tsx'
import City from './City.tsx'

// trigger application build

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <City />
  </StrictMode>
)
