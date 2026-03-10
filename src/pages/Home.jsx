// src/pages/Home.jsx
import React from 'react';
import { Link } from 'react-router-dom';
import './Home.css';

const CATEGORIES = [
  { label: 'Movies', icon: '🎬', query: 'movies' },
  { label: 'Anime', icon: '🎌', query: 'anime' },
  { label: 'Comics', icon: '💥', query: 'comics' },
  { label: 'Books', icon: '📚', query: 'books' },
];

const FEATURED = [
  { id: 1, title: 'Action', icon: '🎬' },
  { id: 2, title: 'Adventure', icon: '🗺️' },
  { id: 3, title: 'Drama', icon: '🎭' },
  { id: 4, title: 'Sci-Fi', icon: '🚀' },
  { id: 5, title: 'Fantasy', icon: '🧙' },
  { id: 6, title: 'Horror', icon: '👻' },
];

export function Home() {
  return (
    <div className="home">
      <section className="hero">
        <h1>Welcome to Media Hub</h1>
        <p>Discover movies, anime, comics, and more — all in one place.</p>
        <Link to="/search" className="hero-link">Start Exploring</Link>
      </section>

      <section className="categories">
        <h2>Browse Categories</h2>
        <div className="categories-grid">
          {CATEGORIES.map((cat) => (
            <Link
              key={cat.label}
              to={`/search?category=${cat.query}`}
              className="category-card"
            >
              <span className="category-icon">{cat.icon}</span>
              {cat.label}
            </Link>
          ))}
        </div>
      </section>

      <section className="featured">
        <h2>Explore by Genre</h2>
        <div className="featured-grid">
          {FEATURED.map((item) => (
            <Link
              key={item.id}
              to={`/search?q=${encodeURIComponent(item.title)}`}
              className="featured-card"
            >
              <div className="featured-card-placeholder">{item.icon}</div>
              <h3>{item.title}</h3>
            </Link>
          ))}
        </div>
      </section>
    </div>
  );
}
