// src/components/Navigation.jsx
import React from 'react';
import { NavLink } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { logout } from '../services/auth';
import './Navigation.css';

export function Navigation() {
  const { currentUser } = useAuth();

  const handleLogout = async () => {
    try {
      await logout();
    } catch (error) {
      console.error('Logout failed:', error);
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
          <button className="nav-button" onClick={handleLogout}>
            Logout
          </button>
        ) : (
          <NavLink to="/login" className="nav-link">
            Login
          </NavLink>
        )}
      </div>
    </nav>
  );
}
