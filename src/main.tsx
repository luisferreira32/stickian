import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './city.css'
//import App from './App.tsx'
import City from './City.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <City />
  </StrictMode>,
)
