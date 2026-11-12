// src/components/Navigation.jsx
import React, { useState } from 'react';
import { NavLink } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { logout } from '../services/auth';
import './Navigation.css';

export function Navigation() {
  const { currentUser } = useAuth();
  const [loggingOut, setLoggingOut] = useState(false);
  const [logoutError, setLogoutError] = useState(false);

  const handleLogout = async () => {
    setLoggingOut(true);
    setLogoutError(false);
    try {
      await logout();
    } catch {
      setLogoutError(true);
    } finally {
      setLoggingOut(false);
    }
  };

  return (
    <nav className="nav">
      <NavLink to="/" className="nav-brand">
        Media Hub
      </NavLink>
      <div className="nav-links">
        <NavLink to="/" className="nav-link" end>
          Home
        </NavLink>
        <NavLink to="/search" className="nav-link">
          Search
        </NavLink>
        {currentUser ? (
          <button className="nav-button" disabled={loggingOut} onClick={handleLogout}>
            Logout
          </button>
        ) : (
          <NavLink to="/login" className="nav-link">
            Login
          </NavLink>
        )}
        {currentUser && logoutError && <span role="alert">Could not sign out. Try again.</span>}
      </div>
    </nav>
  );
}

