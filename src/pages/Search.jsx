// src/pages/Search.jsx
import React, { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { searchMovies, searchAnime, searchComics, searchBooks } from '../services/api';
import './Search.css';

const CATEGORIES = [
  { value: 'movies', label: 'Movies' },
  { value: 'anime', label: 'Anime' },
  { value: 'comics', label: 'Comics' },
  { value: 'books', label: 'Books' },
];

function normalizeResult(item, category) {
  if (category === 'movies') {
    return {
      id: item.id,
      title: item.title,
      image: item.poster_path
        ? `https://image.tmdb.org/t/p/w300${item.poster_path}`
        : null,
      subtitle: item.release_date ? item.release_date.slice(0, 4) : '',
    };
  }
  if (category === 'anime') {
    return {
      id: item.mal_id,
      title: item.title,
      image: item.images?.jpg?.image_url || null,
      subtitle: item.type || '',
    };
  }
  if (category === 'comics') {
    return {
      id: item.id,
      title: item.name,
      image: item.image?.medium_url || null,
      subtitle: item.volume?.name || '',
    };
  }
  if (category === 'books') {
    return {
      id: item.key,
      title: item.title,
      image: item.cover_i
        ? `https://covers.openlibrary.org/b/id/${item.cover_i}-M.jpg`
        : null,
      subtitle: item.author_name ? item.author_name[0] : '',
    };
  }
  return { id: item.id, title: String(item), image: null, subtitle: '' };
}

export function Search() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [query, setQuery] = useState(searchParams.get('q') || '');
  const [category, setCategory] = useState(
    searchParams.get('category') || 'movies'
  );
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    const q = searchParams.get('q');
    const cat = searchParams.get('category') || 'movies';
    if (q) {
      setQuery(q);
      setCategory(cat);
      runSearch(q, cat);
    }
  }, [searchParams]); // eslint-disable-line react-hooks/exhaustive-deps

  async function runSearch(q, cat) {
    if (!q.trim()) return;
    setLoading(true);
    setError(null);
    try {
      let raw = [];
      if (cat === 'movies') raw = await searchMovies(q);
      else if (cat === 'anime') raw = await searchAnime(q);
      else if (cat === 'comics') raw = await searchComics(q);
      else if (cat === 'books') raw = await searchBooks(q);
      setResults(raw.map((item) => normalizeResult(item, cat)));
    } catch (err) {
      setError(err.message);
      setResults([]);
    } finally {
      setLoading(false);
    }
  }

  function handleSubmit(e) {
    e.preventDefault();
    setSearchParams({ q: query, category });
    runSearch(query, category);
  }

  return (
    <div className="search">
      <div className="search-container">
        <h1>Search</h1>
        <form className="search-form" onSubmit={handleSubmit}>
          <select
            className="search-input"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            style={{ flex: '0 0 auto', width: '130px' }}
          >
            {CATEGORIES.map((c) => (
              <option key={c.value} value={c.value}>
                {c.label}
              </option>
            ))}
          </select>
          <input
            className="search-input"
            type="text"
            placeholder="Search..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <button className="search-button" type="submit">
            Search
          </button>
        </form>

        {loading && <p className="loading">Loading...</p>}
        {error && <p className="loading" style={{ color: '#e53e3e' }}>{error}</p>}

        {!loading && !error && results.length > 0 && (
          <div className="results-grid">
            {results.map((item) => (
              <div key={item.id} className="result-card">
                {item.image ? (
                  <img src={item.image} alt={item.title} />
                ) : (
                  <div
                    style={{
                      height: '300px',
                      background: 'linear-gradient(135deg, #667eea, #764ba2)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '3rem',
                    }}
                  >
                    🎬
                  </div>
                )}
                <h3>{item.title}</h3>
                {item.subtitle && <p>{item.subtitle}</p>}
              </div>
            ))}
          </div>
        )}

        {!loading && !error && results.length === 0 && query && (
          <p className="loading">No results found.</p>
        )}
      </div>
    </div>
  );
}
