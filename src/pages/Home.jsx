// src/pages/Home.jsx
import React from 'react';
import { Link } from 'react-router-dom';
import { movieGenres } from '../movieGenres';
import './Home.css';

const CATEGORIES = [
  { label: 'Movies', icon: '🎬', query: 'movies' },
  { label: 'Anime', icon: '🎌', query: 'anime' },
  { label: 'Comics', icon: '💥', query: 'comics' },
  { label: 'Books', icon: '📚', query: 'books' },
];

const GENRE_ICONS = ['🎬', '🗺️', '🎭', '🚀', '🧙', '👻'];

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
          {[...movieGenres].map(([id, title], index) => (
            <Link
              key={id}
              to={`/search?category=movies&genre=${id}`}
              className="featured-card"
            >
              <div className="featured-card-placeholder">{GENRE_ICONS[index]}</div>
              <h3>{title}</h3>
            </Link>
          ))}
        </div>
      </section>
    </div>
  );
}

