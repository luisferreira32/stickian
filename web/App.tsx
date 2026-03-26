import {
  Link,
  Route,
  BrowserRouter as Router,
  Routes,
  useNavigate,
} from 'react-router-dom'
import './App.css'
import City from './features/city/City'
import Dummy from './features/dummy/Dummy'
import Login from './features/login/Login'
import Signup from './features/login/Signup'
import { isAuthenticated, logout } from './shared/auth'
import WorldMap from './features/map/WorldMap'

const Navigation = () => {
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <nav className="navbar">
      <div className="nav-brand">
        <Link to="/">Stickian</Link>
      </div>
      <div className="nav-links">
        {isAuthenticated() ? (
          <>
            <Link to="/dummy">Home</Link>
            <Link to="/map">World</Link>
            <button onClick={handleLogout} className="logout-btn">
              Logout
            </button>
          </>
        ) : (
          <>
            <Link to="/login">Login</Link>
            <Link to="/signup">Sign Up</Link>
          </>
        )}
      </div>
    </nav>
  )
}

const Fallback = () => {
  const navigate = useNavigate()
  if (isAuthenticated()) {
    navigate('/dummy')
  } else {
    navigate('/login')
  }
  return null
}

const App = () => {
  const authed = isAuthenticated()

  return (
    <Router>
      <div className="app">
        <Navigation />
        <div className="main-content">
          <Routes>
            <Route path="/signup" element={<Signup />} />
            <Route path="/login" element={<Login />} />
            {authed && <Route path="/dummy" element={<Dummy />} />}
            {authed && <Route path="/city" element={<City />} />}
            {authed && <Route path="/map" element={<WorldMap />} />}
            <Route path="*" element={<Fallback />} />
          </Routes>
        </div>
      </div>
    </Router>
  )
}

export default App
