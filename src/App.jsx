import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './hooks/useAuth';
import { Navigation } from './components/Navigation';
import { Home } from './pages/Home';
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
        </Routes>
      </AuthProvider>
    </Router>
  );
}

export default App;