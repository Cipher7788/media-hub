import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './hooks/useAuth';
import { Navigation } from './components/Navigation';
import { Home } from './pages/Home';
import { Collections } from './pages/Collections';
import Login from './pages/Login';
import { Search } from './pages/Search';
import './App.css';

function App() {
  return (
    <Router>
      <AuthProvider>
        <Navigation />
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/search" element={<Search />} />
          <Route path="/collections" element={<Collections />} />
          <Route path="/login" element={<Login />} />
          <Route path="*" element={<p role="status">Page not found.</p>} />
        </Routes>
      </AuthProvider>
    </Router>
  );
}

export default App;
